package main

import (
	"context"
	"crypto"
	"crypto/ed25519"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"maps"
	mathrand "math/rand/v2"
	"net"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/mjl-/bstore"
	"github.com/mjl-/sherpa"
	"github.com/mjl-/sherpaprom"
)

//go:embed web/*
var embedFS embed.FS

var fsys fs.FS = localFS()

func localFS() fs.FS {
	if _, err := os.Stat("web"); err == nil {
		return os.DirFS("web")
	}
	fs, err := fs.Sub(embedFS, "web")
	xcheckf(err, "embed fs")
	return fs
}

var tlsPrivKey crypto.Signer
var tlsCert tls.Certificate
var tlsConfig tls.Config

var shutdownCtx context.Context
var shutdownCancel func()

var logLevel slog.LevelVar

var database *bstore.DB
var databaseTypes = []any{Zone{}, ProviderConfig{}, Record{}, ZoneNotify{}, Credential{}, ZoneCredential{}}

var propagationFirstWait = time.Second / 10 // Set to 0 during testing.

// default file, created if absent
var tlskeypemDefault = "server.privkey-ed25519.pkcs8.pem"

// How/if printing of dns request/response messages should be done.
type traceDNS string

const (
	traceNone       traceDNS = ""
	traceText       traceDNS = "text"
	traceJSON       traceDNS = "json"
	traceJSONIndent traceDNS = "jsonindent"
)

// We assign each dns tcp connection and udp packet a cid and use it in logging.
type ctxKey string

var ctxKeyCID = ctxKey("cid")

func cidlog(cidctx context.Context) *slog.Logger {
	log := slog.Default()
	cid := cidctx.Value(ctxKeyCID)
	if cid != nil {
		log = log.With("cid", cid)
	}
	return log
}

var (
	metricDNSRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dnsclay_dns_request_total",
			Help: "DNS requests and response codes.",
		},
		[]string{
			"kind",  // "notify", "update", "axfr", "authoritative", "other"; Not opcode or type, since DNS encodes some commands as opcode and some as record type.
			"rcode", // known strings in lower-case, or "other".
		},
	)
	metricSyncErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "dnsclay_sync_errors_total",
			Help: "Number of errors during processing updated records during sync.",
		},
	)
	metricPropagateErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "dnsclay_propagate_errors_total",
			Help: "Number of errors ensuring dns changes have been propagated at provider.",
		},
	)
	metricProviderOp = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dnsclay_provider_op_duration_seconds",
			Help:    "Provider operation duration.",
			Buckets: []float64{0.05, 0.1, 0.5, 1, 5, 10, 20, 30},
		},
		[]string{
			"provider",
			"op", // "get", "append", "set", "delete"
		},
	)
	metricProviderOpErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dnsclay_provider_op_errors_total",
			Help: "Provider request errors",
		},
		[]string{
			"provider",
			"op", // "get", "append", "set", "delete"
		},
	)
	metricSOAGet = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "dnsclay_soa_get_total",
			Help: "Number requests for a soa record directly to authoritative name servers.",
		},
	)
	metricSOAGetErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "dnsclay_soa_get_errors_total",
			Help: "Number of errors for requests for a soa record directly to authoritative name servers.",
		},
	)
	metricPanics = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "dnsclay_panics_total",
			Help: "Number of unhandled panics.",
		},
	)
)

// Read from file at startup, or generated & written if the file is missing. Always set during operation.
var adminpassword string

func genpassword() string {
	var seed [32]byte
	_, err := cryptorand.Read(seed[:])
	if err != nil {
		panic(err)
	}
	secretRand := mathrand.New(mathrand.NewChaCha8(seed))

	var r string
	const chars = "abcdefghijklmnopqrstuwvxyzABCDEFGHIJKLMNOPQRSTUWVXYZ0123456789"
	for i := 0; i < 12; i++ {
		r += string(chars[secretRand.IntN(len(chars))])
	}
	return r
}

func slogInit() {
	slogOpts := slog.HandlerOptions{
		Level: &logLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == "time" {
				return slog.Attr{}
			}
			if a.Key == "cid" && a.Value.Kind() == slog.KindInt64 {
				return slog.String("cid", fmt.Sprintf("%x", a.Value.Int64()))
			}
			return a
		},
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slogOpts))
	slog.SetDefault(logger)
}

// Allowed operations for a connection/request. We don't do updates/xfr over UDP.
type listener struct {
	tls     bool
	notify  bool
	updates bool
	xfr     bool
	auth    bool
}

var serveTraceDNS []traceDNS

var errUnknownTLSPublicKey = errors.New("unknown tls public key")

func tlsServerConfig(cert tls.Certificate) tls.Config {
	return tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequestClientCert,
		NextProtos:   []string{"dot"},
		VerifyConnection: func(cs tls.ConnectionState) error {
			if len(cs.PeerCertificates) == 0 {
				return nil
			}
			sum := sha256.Sum256(cs.PeerCertificates[0].RawSubjectPublicKeyInfo)
			tlspubkey := base64.RawURLEncoding.EncodeToString(sum[:])
			q := bstore.QueryDB[Credential](shutdownCtx, database)
			q.FilterNonzero(Credential{Type: "tlspubkey", TLSPublicKey: tlspubkey})
			_, err := q.Get()
			if err == bstore.ErrAbsent {
				return fmt.Errorf("%w %s", errUnknownTLSPublicKey, tlspubkey)
			} else if err != nil {
				return fmt.Errorf("looking up tls public key %s: %w", tlspubkey, err)
			}
			return nil
		},
	}
}

func makeAdminMux() *http.ServeMux {
	// Read sherpa API documentation.
	ff, err := fsys.Open("api.json")
	xcheckf(err, "opening sherpa docs")
	err = json.NewDecoder(ff).Decode(&apiDoc)
	xcheckf(err, "parsing sherpa docs")
	err = ff.Close()
	xcheckf(err, "closing sherpa docs after parsing")

	collector, err := sherpaprom.NewCollector("dnsclay", nil)
	xcheckf(err, "creating sherpa prometheus collector")

	// Sherpa web API init.
	opts := &sherpa.HandlerOpts{
		Collector:           collector,
		AdjustFunctionNames: "none",
	}
	handler, err := sherpa.NewHandler("/api/", version, API{}, &apiDoc, opts)
	xcheckf(err, "making sherpa handler")
	adminMux := http.NewServeMux()
	// todo: better authentication mechanism than http basic, perhaps similar as in mox & ding.
	authedHandler := http.HandlerFunc(httpBasicAuth(handler.ServeHTTP))
	adminMux.Handle("GET /api/", authedHandler)
	adminMux.Handle("POST /api/", authedHandler)
	adminMux.Handle("OPTIONS /api/", authedHandler)
	adminMux.HandleFunc("GET /license", httpBasicAuth(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		err := licenseWrite(w)
		logCheck(slog.Default(), err, "respond with license")
	}))
	adminMux.HandleFunc("GET /dnsclay.db", httpBasicAuth(exportDatabase))
	adminMux.HandleFunc("GET /", httpBasicAuth(http.FileServerFS(fsys).ServeHTTP))
	return adminMux
}

func cmdServe(args []string) {
	flg := flag.NewFlagSet("dnsclay serve", flag.ExitOnError)

	var adminpasswordpath string
	var tcpdnsupxfrAddrs, tcpdnsnotifyAddrs, tlsdnsupxfrAddrs, tlsdnsnotifyAddrs, adminAddr, metricsAddr string
	var udpdnsAddrs string
	var tlskeypem, tlscertpem string
	var trace string

	flg.TextVar(&logLevel, "loglevel", &logLevel, "log level: error, warn, info, debug")
	flg.StringVar(&trace, "trace", "", "if non-empty, comma-separated formats to log dns request/response traces: text for textual format, json for json, jsonindent for multi-line indented json")
	flg.StringVar(&adminpasswordpath, "adminpasswordpath", "adminpassword", "file with admin password for http basic auth; if absent, a random password is generated and written")
	flg.StringVar(&udpdnsAddrs, "dns-udpaddr", "localhost:1053", "comma-separated udp address to serve dns notify and authoritative soa requests on")
	flg.StringVar(&tcpdnsupxfrAddrs, "dns-upxfr-tcpaddr", "localhost:1053", "comma-separated tcp address to serve dns update and axfr requests on")
	flg.StringVar(&tcpdnsnotifyAddrs, "dns-notify-tcpaddr", "", "comma-separated tcp address to listen for dns notify messages on")
	flg.StringVar(&tlsdnsupxfrAddrs, "dns-upxfr-tlsaddr", "localhost:1853", "comma-separated tls address to serve dns update and axfr requests on")
	flg.StringVar(&tlsdnsnotifyAddrs, "dns-notify-tlsaddr", "", "comma-separated tls address to listen for dns notify messages on")
	flg.StringVar(&tlskeypem, "tlskeypem", tlskeypemDefault, "path to pem file with pkcs#8 private key file, for dns tls server; if empty an ephemeral tls key is generated at startup; if left at default, file is created if missing")
	flg.StringVar(&tlscertpem, "tlscertpem", "", "path to pem file with one or more certificates; if empty, an ephemeral minimalistic certificate is generated for the private key")
	flg.StringVar(&adminAddr, "adminaddr", "localhost:8053", "address to serve admin interface on")
	flg.StringVar(&metricsAddr, "metricsaddr", "localhost:8053", "address to serve prometheus metrics on; can be same as adminaddr, no authentication needed")
	flg.Usage = func() {
		log.Printf("usage: dnsclay serve [flags]")
		flg.PrintDefaults()
		os.Exit(2)
	}
	flg.Parse(args)
	if trace != "" {
		for _, s := range strings.Split(trace, ",") {
			switch t := traceDNS(s); t {
			case traceNone, traceText, traceJSON, traceJSONIndent:
				serveTraceDNS = append(serveTraceDNS, t)
			default:
				log.Fatalf("unknown -trace value %q", trace)
			}
		}
	}
	args = flg.Args()
	if len(args) != 0 {
		log.Printf("no parameters allowed")
		flg.Usage()
	}

	slogInit()

	// Ensure we have an admin password for the web interface.
	pwbuf, err := os.ReadFile(adminpasswordpath)
	if err == nil {
		adminpassword = strings.TrimRight(string(pwbuf), "\n")
	} else if errors.Is(err, fs.ErrNotExist) {
		adminpassword = genpassword()
		err := os.WriteFile(adminpasswordpath, []byte(adminpassword+"\n"), 0600)
		xcheckf(err, "write adminpassword file")
		log.Printf("generated new admin password: %s", adminpassword)
	} else {
		log.Fatalf("reading adminpassword: %v", err)
	}

	shutdownCtx, shutdownCancel = context.WithCancel(context.Background())

	if tlskeypem != "" {
		tlsPrivKey = xprivatekey(tlskeypem, true)
	} else {
		_, tlsPrivKey, err = ed25519.GenerateKey(cryptorand.Reader)
		xcheckf(err, "generating ephemeral private key")
	}

	if tlscertpem != "" {
		tlsCert = xreadcert(tlscertpem, tlsPrivKey)
	} else {
		tlsCert = xminimalCert(tlsPrivKey)
		if tlskeypem == "" {
			slog.Debug("generated ephemeral private key")
		}
	}

	tlsConfig = tlsServerConfig(tlsCert)
	tlspubkeysum := sha256.Sum256(tlsCert.Leaf.RawSubjectPublicKeyInfo)
	tlspubkeyhash := base64.RawURLEncoding.EncodeToString(tlspubkeysum[:])

	// Possibly a shared handler for admin & metrics.
	adminMux := makeAdminMux()
	metricsMux := http.NewServeMux()
	if adminAddr != "" && adminAddr == metricsAddr {
		metricsMux = adminMux
	}

	metricsMux.Handle("GET /metrics", promhttp.Handler())

	// Open/initialize database.
	dbopts := bstore.Options{
		Timeout: 5 * time.Second,
	}
	database, err = bstore.Open(context.Background(), "dnsclay.db", &dbopts, databaseTypes...)
	xcheckf(err, "open database")

	slog.Info("dnsclay starting",
		"dns-udpaddr", udpdnsAddrs,
		"dns-upxfr-tcpaddr", tcpdnsupxfrAddrs,
		"dns-notify-tcpaddrs", tcpdnsnotifyAddrs,
		"dns-upxfr-tcpaddrs", tlsdnsupxfrAddrs,
		"dns-notify-tcpaddrs", tlsdnsnotifyAddrs,
		"tlspubkeyhash", tlspubkeyhash,
		"adminaddr", adminAddr,
		"metricsaddr", metricsAddr,
		"version", version)

	dnsListeners := map[string]listener{}

	addAddrs := func(s string, l listener) {
		if s == "" {
			return
		}
		for _, a := range strings.Split(s, ",") {
			x, ok := dnsListeners[a]
			if ok {
				if x.tls != l.tls {
					log.Fatalf("cannot serve plain tcp and tls on same address %s", a)
				}
			}
			l.notify = l.notify || x.notify
			l.updates = l.updates || x.updates
			l.xfr = l.xfr || x.xfr
			l.auth = l.auth || x.auth
			dnsListeners[a] = l
		}
	}
	addAddrs(tcpdnsupxfrAddrs, listener{false, false, true, true, true})
	addAddrs(tcpdnsnotifyAddrs, listener{false, true, false, false, false})
	addAddrs(tlsdnsupxfrAddrs, listener{true, false, true, true, true})
	addAddrs(tlsdnsupxfrAddrs, listener{true, true, false, false, false})

	// DNS NOTIFY is commonly done over UDP. AXFR clients may request the SOA before
	// initiating an AXFR, so we handle requests for authoritative SOA records over UDP
	// too. rfc/5936:1090
	if udpdnsAddrs != "" {
		for _, addr := range strings.Split(udpdnsAddrs, ",") {
			lconn, err := net.ListenPacket("udp", addr)
			xcheckf(err, "listen on udp %s", addr)

			// We just read one packet, handle it writing a response, then read the next. We
			// are only expecting NOTIFY messages and requests for authoritative SOA records.
			go func() {
				// Larger than needed, but easy to for shared code with tcp handling.
				buf := make([]byte, 2+64*1024)

				for {
					n, addr, err := lconn.ReadFrom(buf)
					if err != nil {
						slog.Error("read udp packet", "err", err)
						return
					}

					cid := connID.Add(1)
					c := &conn{
						cid:           cid,
						udpRemoteAddr: addr,
						udpconn:       lconn,
						log:           slog.With("cid", cid),
						listener:      listener{false, true, false, false, true},
						buf:           buf,
					}
					c.log.Debug("new request", "remoteaddr", addr)
					c.handleDNS(buf[:n])
				}
			}()
		}
	}

	for _, addr := range slices.Sorted(maps.Keys(dnsListeners)) {
		l := dnsListeners[addr]
		lconn, err := net.Listen("tcp", addr)
		xcheckf(err, "listening on tcp %s", addr)

		go func() {
			for {
				conn, err := lconn.Accept()
				xcheckf(err, "accept")
				go func() {
					defer recoverPanic(slog.Default(), "serving dns connection")
					serveDNS(conn, l)
				}()
			}
		}()
	}

	if adminAddr != "" && adminAddr == metricsAddr {
		// Single web server for both admin and metrics.
		conn, err := net.Listen("tcp", adminAddr)
		xcheckf(err, "listen for webserver")

		go func() {
			server := http.Server{
				Handler: adminMux,
				ConnContext: func(ctx context.Context, c net.Conn) context.Context {
					return context.WithValue(ctx, ctxKeyCID, connID.Add(1))
				},
			}
			err := server.Serve(conn)
			xcheckf(err, "serve webserver")
		}()
	} else {
		if adminAddr != "" {
			adminconn, err := net.Listen("tcp", adminAddr)
			xcheckf(err, "listen for admin webserver")

			go func() {
				server := http.Server{
					Handler: adminMux,
					ConnContext: func(ctx context.Context, c net.Conn) context.Context {
						return context.WithValue(ctx, ctxKeyCID, connID.Add(1))
					},
				}
				err := server.Serve(adminconn)
				xcheckf(err, "serve admin webserver")
			}()
		}

		if metricsAddr != "" {
			// For separate metrics webserver, redirect user to metrics.
			metricsMux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/metrics", http.StatusFound)
			})

			metricsconn, err := net.Listen("tcp", metricsAddr)
			xcheckf(err, "listening for metrics webserver")

			go func() {
				server := http.Server{
					Handler: metricsMux,
					ConnContext: func(ctx context.Context, c net.Conn) context.Context {
						return context.WithValue(ctx, ctxKeyCID, connID.Add(1))
					},
				}
				err := server.Serve(metricsconn)
				xcheckf(err, "serving metrics webserver")
			}()
		}
	}

	go func() {
		defer recoverPanic(slog.Default(), "periodic zone refresher")
		refresher()
	}()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGTERM)
	<-sigc
	slog.Info("shutting down")
	shutdownCancel()
	// todo: wait for all connections and operations to finish, then quit earlier if possible.
	time.Sleep(time.Second / 2)
	os.Exit(0)
}

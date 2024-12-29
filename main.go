package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"reflect"
	"runtime/debug"
	"strings"
	"time"

	"github.com/libdns/libdns"

	"github.com/miekg/dns"

	"github.com/mjl-/bstore"
)

type Provider struct {
	name string
	libdnsProvider
}

func (p Provider) withMetric(op string, fn func() ([]libdns.Record, error)) ([]libdns.Record, error) {
	t0 := time.Now()
	l, err := fn()
	metricProviderOp.WithLabelValues(p.name, op).Observe(float64(time.Since(t0) / time.Second))
	if err != nil {
		metricProviderOpErrors.WithLabelValues(p.name, op).Inc()
	}
	return l, err
}

func (p Provider) AppendRecords(ctx context.Context, zone string, recs []libdns.Record) (l []libdns.Record, err error) {
	return p.withMetric("append", func() ([]libdns.Record, error) {
		return p.libdnsProvider.AppendRecords(ctx, zone, recs)
	})
}

func (p Provider) DeleteRecords(ctx context.Context, zone string, recs []libdns.Record) (l []libdns.Record, err error) {
	return p.withMetric("delete", func() ([]libdns.Record, error) {
		return p.libdnsProvider.DeleteRecords(ctx, zone, recs)
	})
}

func (p Provider) SetRecords(ctx context.Context, zone string, recs []libdns.Record) (l []libdns.Record, err error) {
	return p.withMetric("set", func() ([]libdns.Record, error) {
		return p.libdnsProvider.SetRecords(ctx, zone, recs)
	})
}

func (p Provider) GetRecords(ctx context.Context, zone string) (l []libdns.Record, err error) {
	return p.withMetric("get", func() ([]libdns.Record, error) {
		return p.libdnsProvider.GetRecords(ctx, zone)
	})
}

type libdnsProvider interface {
	libdns.RecordAppender
	libdns.RecordDeleter
	libdns.RecordSetter
	libdns.RecordGetter
}

func xcheckf(err error, format string, args ...any) {
	if err != nil {
		log.Fatalf("%s: %s", fmt.Sprintf(format, args...), err)
	}
}

func logCheck(log *slog.Logger, err error, msg string, attrs ...any) {
	if err != nil {
		attrs = append([]any{"err", err}, attrs...)
		log.Error(msg, attrs...)
	}
}

func main() {
	log.SetFlags(0)

	flag.Usage = func() {
		log.Printf("usage: dnsclay serve [flags]")
		log.Printf("       dnsclay genkey >privkey-ed25519.pkcs8.pem")
		log.Printf("       dnsclay dns [flags] notify [flags] addr zone")
		log.Printf("       dnsclay dns [flags] update [flags] addr zone [add|del name type ttl value] ...")
		log.Printf("       dnsclay dns [flags] xfr [flags] addr zone")
		log.Printf("       dnsclay version")
		log.Printf("       dnsclay license")
		flag.PrintDefaults()
		os.Exit(2)
	}
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
	}

	cmd, args := args[0], args[1:]

	switch cmd {
	case "genkey":
		cmdGenkey(args)

	case "serve":
		cmdServe(args)

	case "dns":
		cmdDNS(args)

	case "version":
		if len(args) != 0 {
			flag.Usage()
		}
		fmt.Println(version)

	case "license":
		cmdLicense(args)

	default:
		flag.Usage()
	}
}

func cmdGenkey(args []string) {
	if len(args) != 0 {
		flag.Usage()
	}
	_, privKey, err := ed25519.GenerateKey(cryptorand.Reader)
	xcheckf(err, "generating ed25519 key")
	buf, err := x509.MarshalPKCS8PrivateKey(privKey)
	xcheckf(err, "marshal private key to pkcs8")
	b := pem.Block{Type: "PRIVATE KEY", Bytes: buf}
	err = pem.Encode(os.Stdout, &b)
	xcheckf(err, "write private key pkcs8 pem to stdout")

	tlsCert = xminimalCert(privKey)
	sum := sha256.Sum256(tlsCert.Leaf.RawSubjectPublicKeyInfo)
	log.Printf("tls public key hash: %s", base64.RawURLEncoding.EncodeToString(sum[:]))
}

func recordData(rr dns.RR) (hex, value string, rerr error) {
	gen := dns.RFC3597{Hdr: *rr.Header()}
	if err := gen.ToRFC3597(rr); err != nil {
		return "", "", fmt.Errorf("to generic rr: %v", err)
	}

	t := strings.SplitN(rr.String(), "\t", 5)
	if len(t) != 5 {
		return "", "", fmt.Errorf("cannot parse textual value from record")
	}
	value = t[4]

	return gen.Rdata, value, nil
}

func describe(v any) string {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetIndent("", "\t")
	err := enc.Encode(v)
	if err != nil {
		panic(err)
	}
	return b.String()
}

// zoneSOA returns the record for the zone if it can be found (not necessarily
// present for newly added zones), returning a zero Record if not found.
func zoneSOA(log *slog.Logger, tx *bstore.Tx, zone string) Record {
	q := bstore.QueryTx[Record](tx)
	q.FilterNonzero(Record{Type: Type(dns.TypeSOA), Zone: zone})
	q.FilterFn(func(r Record) bool { return r.Deleted == nil })
	soa, err := q.Get()
	if err != nil {
		log.Error("get soa record for zone, ignoring", "err", err, "zone", zone)
	}
	return soa
}

func recoverPanic(log *slog.Logger, action string) {
	x := recover()
	if x == nil {
		return
	}

	metricPanics.Inc()
	log.Error("unhandled panic", "err", x, "action", action)
	debug.PrintStack()
}

var errBadName = errors.New("invalid name")

// cleanAbsName verifies s adheres to domain name rules.
func cleanAbsName(s string) (string, error) {
	if !strings.HasSuffix(s, ".") {
		return "", fmt.Errorf("%w: %w %q: name must be absolute", errUser, errBadName, s)
	}
	if len(s) >= 255 {
		return "", fmt.Errorf("%w: %w %q: name too long", errUser, errBadName, s)
	}
	s = strings.ToLower(s)
	t := strings.Split(s, ".")
	for _, label := range t[:len(t)-1] {
		if label == "" {
			return "", fmt.Errorf("%w: %w: %q: invalid empty label", errUser, errBadName, s)
		}
		if len(label) > 63 {
			return "", fmt.Errorf("%w: %w: %q: label %q too long", errUser, errBadName, s, label)
		}
	}
	return s, nil
}

func _cleanAbsName(s string) string {
	var err error
	s, err = cleanAbsName(s)
	_checkf(err, "checking name")
	return s
}

var errProviderUserError = errors.New("bad provider")

// providerForConfig parses a JSON config into a provider.
func providerForConfig(name string, configJSON string) (Provider, error) {
	p, ok := providers[name]
	if !ok {
		return Provider{}, fmt.Errorf("%w: unknown provider %q", errProviderUserError, name)
	}

	t := reflect.TypeOf(p)
	v := reflect.New(t)
	dec := json.NewDecoder(strings.NewReader(configJSON))
	dec.DisallowUnknownFields()
	err := dec.Decode(v.Interface())
	if err != nil {
		return Provider{}, fmt.Errorf("%w: parsing provider config: %v", errProviderUserError, err)
	}
	p = v.Interface()
	provider, ok := p.(libdnsProvider)
	if !ok {
		return Provider{}, fmt.Errorf("provider %q with type %T does not implement provider interface", name, p)
	}
	return Provider{name, provider}, nil
}

func zoneProvider(tx *bstore.Tx, zone string) (Zone, Provider, error) {
	z := Zone{Name: zone}
	if err := tx.Get(&z); err != nil {
		return Zone{}, Provider{}, err
	}

	pc := ProviderConfig{Name: z.ProviderConfigName}
	if err := tx.Get(&pc); err != nil {
		return Zone{}, Provider{}, err
	}

	p, err := providerForConfig(pc.ProviderName, pc.ProviderConfigJSON)
	if err != nil {
		return Zone{}, Provider{}, err
	}

	return z, p, nil
}

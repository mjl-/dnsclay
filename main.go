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

// libdnsRR is a libdns.RR along with the optional ProviderData from the
// libdns.Record for known types.
type libdnsRR struct {
	libdns.RR
	ProviderData any // Optional, from libdns.Record.
}

type Provider struct {
	name string
	libdnsProvider
}

func (p Provider) withMetric(op string, fn func() ([]libdnsRR, error)) ([]libdnsRR, error) {
	t0 := time.Now()
	l, err := fn()
	metricProviderOp.WithLabelValues(p.name, op).Observe(float64(time.Since(t0) / time.Second))
	if err != nil {
		metricProviderOpErrors.WithLabelValues(p.name, op).Inc()
	}
	return l, err
}

// todo: get providerData/setProviderData in libdns? as method on type Record, or as an optional interface on such records.

func providerData(r libdns.Record) any {
	switch rr := r.(type) {
	case libdns.Address:
		return rr.ProviderData
	case libdns.CAA:
		return rr.ProviderData
	case libdns.CNAME:
		return rr.ProviderData
	case libdns.MX:
		return rr.ProviderData
	case libdns.NS:
		return rr.ProviderData
	case libdns.SRV:
		return rr.ProviderData
	case libdns.ServiceBinding:
		return rr.ProviderData
	case libdns.TXT:
		return rr.ProviderData
	}
	return nil
}

func setProviderData(r *libdns.Record, data any) {
	switch rr := (*r).(type) {
	case libdns.Address:
		rr.ProviderData = data
		*r = rr
	case libdns.CAA:
		rr.ProviderData = data
		*r = rr
	case libdns.CNAME:
		rr.ProviderData = data
		*r = rr
	case libdns.MX:
		rr.ProviderData = data
		*r = rr
	case libdns.NS:
		rr.ProviderData = data
		*r = rr
	case libdns.SRV:
		rr.ProviderData = data
		*r = rr
	case libdns.ServiceBinding:
		rr.ProviderData = data
		*r = rr
	case libdns.TXT:
		rr.ProviderData = data
		*r = rr
	}
}

func toRRs(l []libdns.Record) []libdnsRR {
	if l == nil {
		return nil
	}
	nl := make([]libdnsRR, len(l))
	for i, r := range l {
		nl[i] = libdnsRR{r.RR(), providerData(r)}
	}
	return nl
}

func toRecords(l []libdnsRR) ([]libdns.Record, error) {
	if l == nil {
		return nil, nil
	}
	nl := make([]libdns.Record, 0, len(l))
	for _, rr := range l {
		r, err := rr.Parse()
		if err != nil {
			return nil, fmt.Errorf("parsing rr %v: %v", rr, err)
		}
		setProviderData(&r, rr.ProviderData)
		nl = append(nl, r)
	}
	return nl, nil
}

func (p Provider) AppendRecords(ctx context.Context, zone string, rrs []libdnsRR) (l []libdnsRR, err error) {
	return p.withMetric("append", func() ([]libdnsRR, error) {
		recs, err := toRecords(rrs)
		if err != nil {
			return nil, fmt.Errorf("converting RRs to records: %v", err)
		}
		l, err := p.libdnsProvider.AppendRecords(ctx, zone, recs)
		return toRRs(l), err
	})
}

func (p Provider) DeleteRecords(ctx context.Context, zone string, rrs []libdnsRR) (l []libdnsRR, err error) {
	return p.withMetric("delete", func() ([]libdnsRR, error) {
		recs, err := toRecords(rrs)
		if err != nil {
			return nil, fmt.Errorf("converting RRs to records: %v", err)
		}
		l, err := p.libdnsProvider.DeleteRecords(ctx, zone, recs)
		return toRRs(l), err
	})
}

func (p Provider) SetRecords(ctx context.Context, zone string, rrs []libdnsRR) (l []libdnsRR, err error) {
	return p.withMetric("set", func() ([]libdnsRR, error) {
		recs, err := toRecords(rrs)
		if err != nil {
			return nil, fmt.Errorf("converting RRs to records: %v", err)
		}
		l, err := p.libdnsProvider.SetRecords(ctx, zone, recs)
		return toRRs(l), err
	})
}

func (p Provider) GetRecords(ctx context.Context, zone string) (rrs []libdnsRR, err error) {
	return p.withMetric("get", func() ([]libdnsRR, error) {
		l, err := p.libdnsProvider.GetRecords(ctx, zone)
		return toRRs(l), err
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
		log.Printf("       dnsclay dns [flags] update [flags] addr zone [add name type ttl value | del...] ...")
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
	seed := make([]byte, ed25519.SeedSize)
	if _, err := cryptorand.Read(seed); err != nil {
		panic(fmt.Errorf("read random: %w", err))
	}
	privKey := ed25519.NewKeyFromSeed(seed)
	buf, err := x509.MarshalPKCS8PrivateKey(privKey)
	xcheckf(err, "marshal private key to pkcs8")
	b := pem.Block{Type: "PRIVATE KEY", Bytes: buf}
	err = pem.Encode(os.Stdout, &b)
	xcheckf(err, "write private key pkcs8 pem to stdout")
	log.Printf("tls private key seed as base64url: %s", base64.RawURLEncoding.EncodeToString(seed))

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

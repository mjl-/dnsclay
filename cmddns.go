package main

import (
	"bytes"
	"crypto"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/miekg/dns"
)

// From command-line.
type tsigArgs struct {
	KeyName   string // Without trailing dot, added during DNS.
	Algorithm string // E.g. hmac-sha256, without trailing dot.
	Secret    string // As base64.
}

// From command-line.
type dnsArgs struct {
	TSIG                *tsigArgs
	TLSClientCert       *tls.Certificate
	TLSServerPubKeyHash []byte // base64-decoded sha256 hash of server subject public key info.
	TLSServerNoVerify   bool   // Don't do webpki verification.
	TLS                 bool   // Do TLS.
}

func (a dnsArgs) Client() dns.Client {
	var c dns.Client
	c.Net = "tcp"
	if a.TLSClientCert != nil || a.TLSServerPubKeyHash != nil || a.TLSServerNoVerify || a.TLS {
		c.Net = "tcp-tls"
		c.TLSConfig = &tls.Config{
			NextProtos: []string{"dot"},
		}
		if a.TLSClientCert != nil {
			c.TLSConfig.Certificates = []tls.Certificate{*a.TLSClientCert}
		}
		if a.TLSServerNoVerify || a.TLSServerPubKeyHash != nil {
			c.TLSConfig.InsecureSkipVerify = true
		}
		if a.TLSServerPubKeyHash != nil {
			c.TLSConfig.VerifyConnection = func(cs tls.ConnectionState) error {
				sum := sha256.Sum256(cs.PeerCertificates[0].RawSubjectPublicKeyInfo)
				if !bytes.Equal(sum[:], a.TLSServerPubKeyHash) {
					return fmt.Errorf("tls server public key hash mismatch")
				}
				return nil
			}
		}
	}
	if a.TSIG != nil {
		c.TsigSecret = map[string]string{
			a.TSIG.KeyName + ".": a.TSIG.Secret,
		}
	}

	return c
}

func cmdDNS(args []string) {
	flg := flag.NewFlagSet("dnsclay dns", flag.ExitOnError)

	var tsig, tlsclientkeypem, tlsclientkey, tlsclientcertpem, tlsserverpubkeyhash string
	var tlsnoverify, xtls bool
	flg.StringVar(&tsig, "tsig", "", `if non-empty, must be of form "keyname:algorithm:secret", where keyname must be known at the server in hostname syntax, with base64 secret, and algorithm eg hmac-sha256)`)
	flg.StringVar(&tlsclientkeypem, "tlsclientkeypem", "", "if non-empty, file with a tls private key to use for authentication to remote server")
	flg.StringVar(&tlsclientkey, "tlsclientkey", "", "if non-empty, a base64-encoded ed25519 key seed (32 bytes) used for authentication to remote server")
	flg.StringVar(&tlsclientcertpem, "tlsclientcertpem", "", "if non-empty, file with a tls certificate to use for authentication to remote server; if a tls client key is specified and no certificate, a minimal self-signed certificate is generated")
	flg.StringVar(&tlsserverpubkeyhash, "tlsserverpubkeyhash", "", "if non-empty, a base64-encoded (either raw-url or standard) sha256 hash of public key that remote is verified against instead of default webpki verification")
	flg.BoolVar(&tlsnoverify, "tlsnoverify", false, "if set, no verification of remote is done for tls connections; by default, tls pkix verification is done")
	flg.BoolVar(&xtls, "tls", false, "if set, make tls connection; implied by any of the tls flags")
	// todo: add flag for ca pem file to use for verification, and exact certificate
	flg.Usage = func() {
		log.Println("usage: dnsclay dns [flags] notify [flags] addr zone")
		log.Println("       dnsclay dns [flags] update [flags] addr zone [(add | delname | deltype | delrecord) ...] ...")
		log.Println("       dnsclay dns [flags] xfr [flags] addr zone")
		flg.PrintDefaults()
		os.Exit(2)
	}
	flg.Parse(args)
	args = flg.Args()
	if len(args) == 0 {
		flg.Usage()
	}

	// Set dnsargs based on flags, with tsig/tls authentication.
	var dnsargs dnsArgs
	if tsig != "" {
		t := strings.Split(tsig, ":")
		if len(t) != 3 {
			flg.Usage()
		}
		dnsargs.TSIG = &tsigArgs{
			strings.TrimSuffix(t[0], "."),
			strings.TrimSuffix(t[1], "."),
			t[2],
		}
	}
	if tlsclientkeypem != "" && tlsclientkey != "" {
		log.Println("cannot use both -tlsclientkeypem and -tlsclientkey")
		flg.Usage()
	}
	var tlsPrivKey crypto.Signer
	if tlsclientkeypem != "" {
		tlsPrivKey = xprivatekey(tlsclientkeypem, false)
	}
	if tlsclientkey != "" {
		seed, err := base64.RawURLEncoding.DecodeString(tlsclientkey)
		if err != nil {
			seed, err = base64.StdEncoding.DecodeString(tlsclientkey)
		}
		if err == nil && len(seed) != ed25519.SeedSize {
			err = fmt.Errorf("got %d bytes, need %d", len(seed), ed25519.SeedSize)
		}
		xcheckf(err, "parsing ed25519 base64 seed")
		tlsPrivKey = ed25519.NewKeyFromSeed(seed)
	}
	if tlsclientcertpem != "" {
		cert := xreadcert(tlsclientcertpem, tlsPrivKey)
		dnsargs.TLSClientCert = &cert
	} else if tlsPrivKey != nil {
		cert := xminimalCert(tlsPrivKey)
		dnsargs.TLSClientCert = &cert
	}

	if tlsserverpubkeyhash != "" {
		pkh, err := base64.RawURLEncoding.DecodeString(tlsserverpubkeyhash)
		if err != nil {
			pkh, err = base64.StdEncoding.DecodeString(tlsserverpubkeyhash)
		}
		if err == nil && len(pkh) != sha256.Size {
			err = fmt.Errorf("got %d bytes, expected %d", len(pkh), sha256.Size)
		}
		xcheckf(err, "parsing tls server public key hash")
		dnsargs.TLSServerPubKeyHash = pkh
	}
	dnsargs.TLSServerNoVerify = tlsnoverify
	dnsargs.TLS = xtls

	cmd, args := args[0], args[1:]
	switch cmd {
	case "notify":
		cmdDNSNotify(dnsargs, args)
	case "update":
		cmdDNSUpdate(dnsargs, args)
	case "xfr":
		cmdDNSXFR(dnsargs, args)
	default:
		flg.Usage()
	}
}

func cmdDNSNotify(dnsargs dnsArgs, args []string) {
	flg := flag.NewFlagSet("dnsclay dns notify", flag.ExitOnError)

	var xjson, trace bool
	flg.BoolVar(&xjson, "json", false, "print dns packets in json too")
	flg.BoolVar(&trace, "trace", false, "print dns packets")

	flg.Usage = func() {
		log.Println("usage: dnsclay dns [flags] notify [-trace] [-json] addr zone")
		flg.PrintDefaults()
		os.Exit(2)
	}
	flg.Parse(args)
	args = flg.Args()
	if len(args) != 2 {
		flg.Usage()
	}

	addr := args[0]
	zone := args[1]

	var om dns.Msg
	om.SetNotify(strings.TrimSuffix(zone, ".") + ".")

	c := dnsargs.Client()
	if tsig := dnsargs.TSIG; tsig != nil {
		om.SetTsig(tsig.KeyName+".", tsig.Algorithm+".", 300, time.Now().Unix())
	}

	im, _, err := c.Exchange(&om, addr)
	if err == nil {
		err = responseError(im)
	}

	if trace {
		log.Println("# request")
		log.Println(om.String())
		if xjson {
			log.Println(describe(om))
		}

		log.Println()
		log.Println("# response")
		log.Println(im.String())
		if xjson {
			log.Println(describe(im))
		}
	}

	xcheckf(err, "send notify")
}

func cmdDNSUpdate(dnsargs dnsArgs, args []string) {
	flg := flag.NewFlagSet("dnsclay dns update", flag.ExitOnError)

	var xjson, trace bool
	flg.BoolVar(&xjson, "json", false, "print dns packets in json too")
	flg.BoolVar(&trace, "trace", false, "print dns packets")

	flg.Usage = func() {
		log.Println(`usage: dnsclay dns [flags] update [-trace] [-json] addr zone [add name ttl type value | delname name | deltype name type | delrecord name type value] ...`)
		flg.PrintDefaults()
		os.Exit(2)
	}
	flg.Parse(args)
	args = flg.Args()

	if len(args) < 2 {
		flg.Usage()
	}

	addr := args[0]
	zone := args[1]

	args = args[2:]

	// todo: add prerequisites. figure out how to specify them on the command-line.

	if len(args) == 0 {
		log.Println("need at least one add or delete statement")
		flg.Usage()
	}

	var om dns.Msg
	om.SetUpdate(strings.TrimSuffix(zone, ".") + ".")

	absname := func(name string) string {
		if strings.HasSuffix(name, ".") {
			return name
		}
		return name + "." + zone
	}

	for len(args) > 0 {
		op := args[0]
		args = args[1:]
		switch op {
		case "add":
			if len(args) < 4 {
				log.Println("add needs 4 arguments")
				flg.Usage()
			}
			text := fmt.Sprintf("%s %s %s %s", absname(args[0]), args[1], args[2], args[3])
			rr, err := dns.NewRR(text)
			xcheckf(err, "parsing record %q", text)

			h := rr.Header()
			if h.Class != dns.ClassINET {
				log.Fatalf("cannot set class %q, need class inet (IN)", dns.Class(h.Class))
			}
			if h.Rrtype == dns.TypeANY || h.Rrtype == dns.TypeNone {
				log.Fatalf("cannot set type %q, NONE and ANY have special meaning for deletion", dns.Type(h.Rrtype))
			}

			args = args[4:]
			om.Ns = append(om.Ns, rr)

		case "delname":
			if len(args) < 1 {
				log.Println("delname needs an argument")
				flg.Usage()
			}
			name, err := cleanAbsName(absname(args[0]))
			xcheckf(err, "parsing name %q", name)
			rr := &dns.RFC3597{
				Hdr: dns.RR_Header{
					Name:   name,
					Rrtype: dns.TypeANY,
					Class:  dns.ClassANY,
				},
			}
			args = args[1:]
			om.Ns = append(om.Ns, rr)

		case "deltype":
			if len(args) < 2 {
				log.Println("deltype needs two arguments")
				flg.Usage()
			}
			t, ok := dns.StringToType[strings.ToUpper(args[1])]
			if !ok {
				log.Fatalf("unknown type %q", args[1])
			}
			name, err := cleanAbsName(absname(args[0]))
			xcheckf(err, "parsing name %q", name)
			rr := &dns.RFC3597{
				Hdr: dns.RR_Header{
					Name:   name,
					Rrtype: t,
					Class:  dns.ClassANY,
				},
			}
			args = args[2:]
			om.Ns = append(om.Ns, rr)

		case "delrecord":
			if len(args) < 3 {
				log.Println("delrecord needs three arguments")
				flg.Usage()
			}

			text := fmt.Sprintf("%s 0 %s %s", absname(args[0]), args[1], args[2])
			rr, err := dns.NewRR(text)
			xcheckf(err, "parsing record %q", text)
			rr.Header().Class = dns.ClassNONE
			args = args[3:]
			om.Ns = append(om.Ns, rr)

		default:
			log.Printf("unknown command %q, need add/del", args[0])
			flg.Usage()
		}
	}

	c := dnsargs.Client()
	if tsig := dnsargs.TSIG; tsig != nil {
		om.SetTsig(tsig.KeyName+".", tsig.Algorithm+".", 300, time.Now().Unix())
	}

	im, _, err := c.Exchange(&om, addr)
	if err == nil {
		err = responseError(im)
	}

	if trace {
		log.Println("# query")
		log.Println(om.String())
		if xjson {
			log.Println(describe(om))
		}

		log.Println()
		log.Println("# response")
		log.Println(im.String())
		if xjson {
			log.Println(describe(im))
		}
	}

	xcheckf(err, "send update")
}

func cmdDNSXFR(dnsargs dnsArgs, args []string) {
	flg := flag.NewFlagSet("dnsclay dns xfr", flag.ExitOnError)

	var serial uint
	var query, xjson bool
	flg.BoolVar(&xjson, "json", false, "print dns packets in json too")
	flg.BoolVar(&query, "query", false, "print dns query")
	flg.UintVar(&serial, "i", 0, "serial number for IXFR request instead of default AXFR")

	flg.Usage = func() {
		log.Println("usage: dnsclay dns [flags] xfr [-i serial] [-json] addr zone")
		flg.PrintDefaults()
		os.Exit(2)
	}
	flg.Parse(args)
	args = flg.Args()
	if len(args) != 2 {
		flg.Usage()
	}

	addr := args[0]
	zone := args[1]

	var om dns.Msg
	if serial > 0 {
		om.SetIxfr(zone, uint32(serial), "dnsclay.example.", "dnsclay.example.")
	} else {
		om.SetAxfr(zone)
	}

	// We can't (easily) verify multiple message with tsig with package dns, we can't
	// set "timersonly" for messages after the first. So we use the dns.Transfer.
	c := dnsargs.Client()
	conn, err := c.Dial(addr)
	xcheckf(err, "dial")
	conn.TsigSecret = c.TsigSecret
	t := dns.Transfer{
		Conn:       conn,
		TsigSecret: c.TsigSecret,
	}
	if tsig := dnsargs.TSIG; tsig != nil {
		om.SetTsig(tsig.KeyName+".", tsig.Algorithm+".", 300, time.Now().Unix())
	}

	if query {
		log.Println("# query")
		log.Println(om.String())
		if xjson {
			log.Println(describe(om))
		}
	}

	envc, err := t.In(&om, "")
	xcheckf(err, "axfr transaction")

	for env := range envc {
		xcheckf(env.Error, "get message")
		if query {
			log.Println()
			log.Println(" # response")
		}

		// Make up a response message with the answer RR's, we don't get it from dns.Transfer.
		var im dns.Msg
		im.MsgHdr = om.MsgHdr
		im.Response = true
		im.Question = om.Question
		im.Answer = env.RR
		fmt.Println(im.String())

		if xjson {
			log.Println(describe(im))
		}
	}
}

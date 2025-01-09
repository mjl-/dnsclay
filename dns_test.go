package main

import (
	"crypto/ed25519"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/libdns/libdns"
	"github.com/miekg/dns"
)

var testUseEDNS0 bool

type dnsclient struct {
	t    *testing.T
	c    *dns.Client
	addr string
}

func (dc dnsclient) exchange(om *dns.Msg, expErr any, expRcode int) *dns.Msg {
	t := dc.t
	t.Helper()
	im, _, err := dc.c.Exchange(om, dc.addr)
	if expErr != nil && (err == nil || !(err == expErr || errors.As(err, expErr))) {
		t.Fatalf("got err %#v, expected %#v", err, expErr)
	} else if expErr == nil {
		tcheck(t, err, "dns transaction")
		tcompare(t, im.Rcode, expRcode)
		if expRcode != dns.RcodeSuccess {
			if err := responseError(im); err == nil {
				t.Fatalf("got success, expected not implemented")
			}
		}
	}
	return im
}

func possiblyEnableEDNS0(om *dns.Msg) {
	if testUseEDNS0 {
		om.SetEdns0(1232, false)
	}
}

func msgNotify(zone string) *dns.Msg {
	var om dns.Msg
	om.SetNotify(zone)
	possiblyEnableEDNS0(&om)
	return &om
}

func msgAXFR(zone string) *dns.Msg {
	var om dns.Msg
	om.SetAxfr(zone)
	possiblyEnableEDNS0(&om)
	return &om
}

func msgAXFRTSIG(name string, tsigKey string, tm time.Time) *dns.Msg {
	om := msgAXFR(name)
	possiblyEnableEDNS0(om)
	om.SetTsig(tsigKey+".", "hmac-sha256.", 300, tm.Unix())
	return om
}

func msgUpdate(zone string) *dns.Msg {
	var om dns.Msg
	om.SetUpdate(zone)
	possiblyEnableEDNS0(&om)
	return &om
}

func msgQuery(name string, typ uint16) *dns.Msg {
	var om dns.Msg
	om.SetQuestion(name, typ)
	possiblyEnableEDNS0(&om)
	return &om
}

func TestNotifyTCP(t *testing.T) {
	testDNS(t, func(te testEnv, z Zone) {
		lconn, err := net.Listen("tcp", "127.0.0.1:0")
		tcheck(t, err, "listen tcp")
		defer lconn.Close()

		l, err := te.z0.p.AppendRecords(ctxbg, z.Name, []libdns.Record{ldr("", "nhost", 300, "A", "10.0.0.3")})
		tcheck(t, err, "add record")
		tcompare(t, len(l), 1)

		tdc := dnsclient{t, &dns.Client{Net: "tcp"}, lconn.Addr().String()}

		// First send notify with same SOA serial. Will not change anything.
		soaRR, err := te.z0.soa.RR()
		tcheck(t, err, "soa rr")
		go func() {
			om := msgNotify(z.Name)
			om.Answer = []dns.RR{soaRR}
			tdc.exchange(om, nil, dns.RcodeSuccess)
		}()

		conn, err := lconn.Accept()
		tcheck(t, err, "accept")

		te.zoneUnchanged(func() {
			serveDNS(conn, listener{notify: true})
		})

		result := make(chan error, 1)
		go func() {
			im, _, err := tdc.c.Exchange(msgNotify(z.Name), tdc.addr)
			if err == nil {
				err = responseError(im)
			}
			result <- err
		}()

		conn, err = lconn.Accept()
		tcheck(t, err, "accept")

		te.zoneChanged(func() {
			serveDNS(conn, listener{notify: true})
		})
		err = <-result
		tcheck(t, err, "send notify")

		// Regular UPDATE/AXFR handler does not implement notify.
		tdc.addr = te.tcpaddr
		tdc.exchange(msgNotify(z.Name), nil, dns.RcodeNotImplemented)
	})
}

func TestNotifyUDP(t *testing.T) {
	testDNS(t, func(te testEnv, z Zone) {
		lconn, err := net.ListenPacket("udp", "127.0.0.1:0")
		tcheck(t, err, "listen udp")
		defer lconn.Close()

		l, err := te.z0.p.AppendRecords(ctxbg, z.Name, []libdns.Record{ldr("", "nhost", 300, "A", "10.0.0.3")})
		tcheck(t, err, "add record")
		tcompare(t, len(l), 1)

		tdc := dnsclient{t, &dns.Client{Net: "udp"}, lconn.LocalAddr().String()}

		result := make(chan error, 1)
		go func() {
			im, _, err := tdc.c.Exchange(msgNotify(z.Name), tdc.addr)
			if err == nil {
				err = responseError(im)
			}
			result <- err
		}()

		buf := make([]byte, 1024)
		n, raddr, err := lconn.ReadFrom(buf)
		tcheck(t, err, "read")

		cid := connID.Add(1)
		c := conn{
			cid:           cid,
			udpRemoteAddr: raddr,
			udpconn:       lconn,
			log:           slog.With("cid", cid),
			listener:      listener{notify: true},
			buf:           buf,
		}

		te.zoneChanged(func() {
			testSyncNotify = true
			c.handleDNS(buf[:n])
			testSyncNotify = false
		})
		err = <-result
		tcheck(t, err, "send notify")
	})
}

func TestDNSMsg(t *testing.T) {
	testDNS(t, func(te testEnv, z Zone) {
		tdc := dnsclient{t, &dns.Client{Net: "tcp"}, te.tcpaddr}

		// Send response instead of query.
		om := msgQuery(z.Name, dns.TypeSOA)
		om.Response = true
		tdc.exchange(om, nil, dns.RcodeFormatError)

		// Send class other than INET.
		om = msgQuery(z.Name, dns.TypeSOA)
		om.Question[0].Qclass = dns.ClassCHAOS
		tdc.exchange(om, nil, dns.RcodeRefused)

		// Send truncated request.
		om = msgQuery(z.Name, dns.TypeSOA)
		om.Truncated = true
		tdc.exchange(om, nil, dns.RcodeFormatError)

		// Send unimplemented opcode.
		om = msgQuery(z.Name, dns.TypeSOA)
		om.Opcode = dns.OpcodeStatus
		tdc.exchange(om, nil, dns.RcodeNotImplemented)
	})
}

func TestAXFR(t *testing.T) {
	testDNS(t, func(te testEnv, z Zone) {
		tdc := dnsclient{t, &dns.Client{Net: "tcp"}, te.tcpaddr}

		tdc.c.TsigSecret = map[string]string{
			te.z0.credTSIG.Name + ".": te.z0.credTSIG.TSIGSecret,
		}

		// Credentials not valid for unknown zone.
		te.zoneUnchanged(func() {
			om := msgAXFRTSIG("bogus.example.", te.z0.credTSIG.Name, time.Now())
			tdc.exchange(om, nil, dns.RcodeRefused)
		})

		// Zone should be up to date. AXFR should not change anything, like a new SOA.
		te.zoneUnchanged(func() {
			om := msgAXFRTSIG(z.Name, te.z0.credTSIG.Name, time.Now())
			tdc.exchange(om, nil, dns.RcodeSuccess)
		})

		// Add record directly via provider.
		l, err := te.z0.p.AppendRecords(ctxbg, z.Name, []libdns.Record{ldr("", "nhost", 300, "A", "10.0.0.3")})
		tcheck(t, err, "add record")
		tcompare(t, len(l), 1)

		// AXFR again, we should see the updated records.
		te.zoneChanged(func() {
			om := msgAXFRTSIG(z.Name, te.z0.credTSIG.Name, time.Now())
			tdc.exchange(om, nil, dns.RcodeSuccess)
		})

		// Add many large records, for testing multi-message axfr.
		many := make([]libdns.Record, 64*4)
		for i := range many {
			many[i] = libdns.Record{
				Name:  "bigtxt",
				Type:  "TXT",
				TTL:   time.Hour,
				Value: fmt.Sprintf(`"%255d"`, i), // 4 of these are 1KB.
			}
		}
		l, err = te.z0.p.AppendRecords(ctxbg, z.Name, many)
		tcheck(t, err, "add record")
		tcompare(t, len(l), len(many))

		// AXFR with multiple messages.
		te.zoneChanged(func() {
			om := msgAXFR(z.Name)
			dc, err := dns.DialWithTLS("tcp", te.tlsaddr, te.z0.tlsConfig)
			tcheck(t, err, "dial dns")
			defer dc.Close()
			err = dc.WriteMsg(om)
			tcheck(t, err, "write axfr request")
			n := 0
			for {
				err = dc.SetDeadline(time.Now().Add(3 * time.Second))
				tcheck(t, err, "set deadline")
				im, err := dc.ReadMsg()
				tcheck(t, err, "read axfr response")
				tcompare(t, im.Rcode, dns.RcodeSuccess)
				tcompare(t, len(im.Answer) > 0, true)
				if n == 0 {
					tcompare(t, im.Answer[0].Header().Rrtype, dns.TypeSOA)
				}
				n++
				if im.Answer[len(im.Answer)-1].Header().Rrtype == dns.TypeSOA && (len(im.Answer) > 1 || n > 1) {
					break
				}
			}
			tcompare(t, n > 1, true)
			err = dc.SetDeadline(time.Now().Add(time.Second / 10))
			tcheck(t, err, "set deadline")
			_, err = dc.ReadMsg()
			if err == nil || !errors.Is(err, os.ErrDeadlineExceeded) {
				t.Fatalf("reading past last axfr message, got err %#v, expected deadline exceeded", err)
			}
		})

		te.zoneUnchanged(func() {
			conn, err := tdc.c.Dial(te.tcpaddr)
			tcheck(t, err, "dial dns")
			defer conn.Close()
			xfr := dns.Transfer{
				Conn:       conn,
				TsigSecret: tdc.c.TsigSecret,
			}
			om := msgAXFRTSIG(z.Name, te.z0.credTSIG.Name, time.Now())

			envc, err := xfr.In(om, "")
			tcheck(t, err, "axfr transaction")

			n := 0
			var end bool
			for env := range envc {
				tcheck(t, env.Error, "get axfr message")

				tcompare(t, len(env.RR) > 0, true)
				if n == 0 {
					tcompare(t, env.RR[0].Header().Rrtype, dns.TypeSOA)
				}
				tcompare(t, end, false)
				end = env.RR[len(env.RR)-1].Header().Rrtype == dns.TypeSOA && (len(env.RR) > 1 || n > 0)
				n++
			}
			tcompare(t, end, true)
		})
	})
}

func TestDNSAuthentication(t *testing.T) {
	testDNS(t, func(te testEnv, z Zone) {
		goodCred := map[string]string{
			te.z0.credTSIG.Name + ".": te.z0.credTSIG.TSIGSecret,
			te.z1.credTSIG.Name + ".": te.z1.credTSIG.TSIGSecret,
			// Not valid because key name is from TLS public key, but for testing:
			te.z0.credTLS.Name + ".": te.z0.credTSIG.TSIGSecret,
			"unknown.":               te.z0.credTSIG.TSIGSecret,
		}
		badCred := map[string]string{
			te.z0.credTSIG.Name + ".": "YmFkIGNyZWRz",
		}
		tdc := dnsclient{t, &dns.Client{Net: "tcp", TsigSecret: goodCred}, te.tcpaddr}

		// AXFR needs auth.
		tdc.exchange(msgAXFR(z.Name), nil, dns.RcodeRefused)

		// TSIG auth is OK for AXFR.
		tdc.exchange(msgAXFRTSIG(z.Name, te.z0.credTSIG.Name, time.Now()), nil, dns.RcodeSuccess)

		// Mismatching TSIG is detected.
		tdc.c.TsigSecret = badCred
		tdc.exchange(msgAXFRTSIG(z.Name, te.z0.credTSIG.Name, time.Now()), &dns.ErrAuth, 0)
		// todo: check tsig error code
		tdc.c.TsigSecret = goodCred

		// Unknown key fails.
		tdc.exchange(msgAXFRTSIG(z.Name, "unknown", time.Now()), nil, dns.RcodeNotAuth)
		// todo: check tsig error code

		// Signature is beyond allowed time.
		tdc.exchange(msgAXFRTSIG(z.Name, te.z0.credTSIG.Name, time.Now().Add(-600*time.Second)), &dns.ErrSig, 0)
		// todo: check tsig error code

		// Key used for TSIG that's known as tls fails.
		tdc.exchange(msgAXFRTSIG(z.Name, te.z0.credTLS.Name, time.Now()), nil, dns.RcodeNotAuth)
		// todo: check tsig error code

		// TSIG auth is for specific zone, other zones give REFUSED.
		tdc.exchange(msgAXFRTSIG(te.z1.z.Name, te.z0.credTSIG.Name, time.Now()), nil, dns.RcodeRefused)

		tdc.c.Net = "tcp-tls"
		tdc.c.TLSConfig = te.z0.tlsConfig
		tdc.addr = te.tlsaddr

		// TLS auth is also fine.
		tdc.exchange(msgAXFR(z.Name), nil, dns.RcodeSuccess)

		// TLS auth is also for specific zone, others give REFUSED.
		tdc.exchange(msgAXFR(te.z1.z.Name), nil, dns.RcodeRefused)

		tdc.c.TLSConfig = &tls.Config{InsecureSkipVerify: true}

		// TLS without auth is not enough.
		tdc.exchange(msgAXFR(z.Name), nil, dns.RcodeRefused)

		// TLS without tls client auth but with tsig is ok.
		tdc.exchange(msgAXFRTSIG(z.Name, te.z0.credTSIG.Name, time.Now()), nil, dns.RcodeSuccess)

		tdc.c.TLSConfig = te.z0.tlsConfig

		// Both TLS and TSIG is also fine.
		tdc.exchange(msgAXFRTSIG(z.Name, te.z0.credTSIG.Name, time.Now()), nil, dns.RcodeSuccess)

		// TLS and TSIG must both give access if both are present.
		tdc.exchange(msgAXFRTSIG(te.z1.z.Name, te.z1.credTSIG.Name, time.Now()), nil, dns.RcodeRefused)

		// Unknown TLS client cert fails the connection.
		badseed := make([]byte, 32)
		badseed[0] = 2
		badkey := ed25519.NewKeyFromSeed(badseed)
		badcert := xminimalCert(badkey)
		tdc.c.TLSConfig = &tls.Config{InsecureSkipVerify: true, Certificates: []tls.Certificate{badcert}}
		var opErr *net.OpError
		tdc.exchange(msgAXFR(z.Name), &opErr, 0)
	})
}

func TestUpdate(t *testing.T) {
	newRR := func(z Zone, s string) dns.RR {
		zp := dns.NewZoneParser(strings.NewReader(s), z.Name, "")
		zp.SetDefaultTTL(300)
		rr, ok := zp.Next()
		tcheck(t, zp.Err(), "parse rr")
		tcompare(t, ok, true)
		_, ok = zp.Next()
		tcompare(t, ok, false)
		return rr
	}

	testUpdate := func(te testEnv, om *dns.Msg, expRcode int) {
		t.Helper()
		c := dns.Client{Net: "tcp-tls", TLSConfig: te.z0.tlsConfig}
		tdc := dnsclient{t, &c, te.tlsaddr}
		tdc.exchange(om, nil, expRcode)
	}

	testDNS(t, func(te testEnv, z Zone) {
		tdc := dnsclient{t, &dns.Client{Net: "tcp"}, te.tcpaddr}

		om := msgUpdate(z.Name)
		om.Insert([]dns.RR{newRR(z, "testhost A 10.0.0.3")})

		// Update requires authentication.
		te.zoneUnchanged(func() {
			tdc.exchange(om, nil, dns.RcodeRefused)
		})

		// TSIG is ok. Other tests use TLS, more convenient. Add a record to be picked up before making changes.
		tc := te.zoneChanged(func() {
			_, err := te.z0.p.AppendRecords(ctxbg, z.Name, []libdns.Record{ldr("", "text", 300, "TXT", `"test"`)})
			tcheck(t, err, "append record")

			tdc.c.TsigSecret = map[string]string{
				te.z0.credTSIG.Name + ".": te.z0.credTSIG.TSIGSecret,
			}
			om = msgUpdate(z.Name)
			om.Insert([]dns.RR{newRR(z, "testhost A 10.0.0.3")})
			om.SetTsig(te.z0.credTSIG.Name+".", "hmac-sha256.", 300, time.Now().Unix())
			tdc.exchange(om, nil, dns.RcodeSuccess)
		})
		tc.checkRecordDelta(typecounts{"A": 2}, typecounts{"TXT": 1, "A": 3})
	})

	testDNS(t, func(te testEnv, z Zone) {
		om := msgUpdate(z.Name)
		om.Insert([]dns.RR{newRR(z, "testhost A 10.0.0.3")})

		// Can only do class INET.
		om.Question[0].Qclass = dns.ClassCHAOS
		testUpdate(te, om, dns.RcodeRefused)
		om.Question[0].Qclass = dns.ClassINET

		// Cannot add a non-INET record.
		om.Ns[0].(*dns.A).Hdr.Class = dns.ClassCHAOS
		testUpdate(te, om, dns.RcodeFormatError)
		om.Ns[0].(*dns.A).Hdr.Class = dns.ClassINET

		// Credentials not valid (unknown zone).
		om = msgUpdate("bogus.example.")
		om.Insert([]dns.RR{newRR(z, "testhost A 10.0.0.3")})
		testUpdate(te, om, dns.RcodeRefused)

		om = msgUpdate(z.Name)
		om.Insert([]dns.RR{newRR(z, "testhost A 10.0.0.3")})

		// Various failing prerequisites.
		te.zoneUnchanged(func() {
			// Cannot change record of other zone.
			om.Answer = nil
			om.NameNotUsed([]dns.RR{newRR(z, "testhost.z1.example. A 10.0.0.1")})
			testUpdate(te, om, dns.RcodeNotZone)

			// Check that name (any type) does not exist. It does.
			om.Answer = nil
			om.NameNotUsed([]dns.RR{newRR(z, "testhost A 10.0.0.1")})
			testUpdate(te, om, dns.RcodeYXDomain)

			// Check that a name is used (any type). It does not.
			om.Answer = nil
			om.NameUsed([]dns.RR{newRR(z, "notused A 10.0.0.1")})
			testUpdate(te, om, dns.RcodeNameError)

			// Check that the name+type does not exist. It does.
			om.Answer = nil
			om.RRsetNotUsed([]dns.RR{newRR(z, "testhost A 10.0.0.1")})
			testUpdate(te, om, dns.RcodeYXRrset)

			// Check that name+type exists. Name doesn't exist at at all.
			om.Answer = nil
			om.RRsetUsed([]dns.RR{newRR(z, "notused A 10.0.0.3")})
			testUpdate(te, om, dns.RcodeNXRrset)

			// Check that name+type exists. Name exists, but with different type.
			om.Answer = nil
			om.RRsetUsed([]dns.RR{newRR(z, "testhost AAAA ::1")})
			testUpdate(te, om, dns.RcodeNXRrset)

			// Check that rrset has same name+type+data.
			om.Answer = nil
			om.Used([]dns.RR{newRR(z, "testhost A 10.0.0.1")})
			testUpdate(te, om, dns.RcodeNXRrset)

			// Check that rrset has same name+type+data.
			om.Answer = nil
			om.Used([]dns.RR{newRR(z, "testhost A 10.0.0.1")})
			testUpdate(te, om, dns.RcodeNXRrset)
		})

		// Succeeding prerequisite.
		tc := te.zoneChanged(func() {
			om.Answer = nil
			om.Used([]dns.RR{newRR(z, "testhost A 10.0.0.1"), newRR(z, "testhost A 10.0.0.2")})

			testUpdate(te, om, dns.RcodeSuccess)
		})
		tc.checkRecordDelta(typecounts{"A": 2}, typecounts{"A": 3})

		// Cannot change record in other zone.
		te.zoneUnchanged(func() {
			om.Answer = nil
			om.RRsetUsed([]dns.RR{newRR(z, "testhost A 10.0.0.1")})

			om.Ns = nil
			om.Insert([]dns.RR{newRR(z, "testhost.z1.example. A 10.0.0.1")})
			testUpdate(te, om, dns.RcodeNotZone)
		})

		// Adding a CNAME when a non-CNAME exists is ignored.
		te.zoneUnchanged(func() {
			om.Answer = nil
			om.Ns = nil
			om.Insert([]dns.RR{newRR(z, "testhost CNAME dangling."+z.Name)})
			testUpdate(te, om, dns.RcodeSuccess)
		})

		// First add a CNAME. Then we insert it with a different target. It should be replaced.
		tc = te.zoneChanged(func() {
			om.Answer = nil
			om.Ns = nil
			om.Insert([]dns.RR{newRR(z, "cname CNAME dangling."+z.Name)})
			testUpdate(te, om, dns.RcodeSuccess)
		})
		tc.checkRecordDelta(typecounts{}, typecounts{"CNAME": 1})
		tc = te.zoneChanged(func() {
			om.Answer = nil
			om.Ns = nil
			om.Insert([]dns.RR{newRR(z, "cname CNAME testhost."+z.Name)})
			testUpdate(te, om, dns.RcodeSuccess)
		})
		tc.checkRecordDelta(typecounts{"CNAME": 1}, typecounts{"CNAME": 1})
	})

	testDNS(t, func(te testEnv, z Zone) {
		om := msgUpdate(z.Name)

		tc := te.zoneChanged(func() {
			om.Remove([]dns.RR{newRR(z, "testhost A 10.0.0.1")})
			testUpdate(te, om, dns.RcodeSuccess)
		})
		// Two A records marked deleted, one created again.
		tc.checkRecordDelta(typecounts{"A": 2}, typecounts{"A": 1})
	})

	testDNS(t, func(te testEnv, z Zone) {
		om := msgUpdate(z.Name)

		// First add AAAA under existing name.
		tc := te.zoneChanged(func() {
			te.api.RecordSetAdd(ctxbg, z.Name, RecordSetChange{"testhost", 300, Type(dns.TypeAAAA), []string{"::1"}})
		})
		tc.checkRecordDelta(typecounts{}, typecounts{"AAAA": 1})

		// Only remove "testhost A", will keep "AAAA".
		tc = te.zoneChanged(func() {
			om.RemoveRRset([]dns.RR{newRR(z, "testhost A 10.0.0.1")})
			testUpdate(te, om, dns.RcodeSuccess)
		})
		tc.checkRecordDelta(typecounts{"A": 2}, typecounts{})

		// Now remove all types for name.
		tc = te.zoneChanged(func() {
			om.RemoveName([]dns.RR{newRR(z, "testhost A 10.0.0.1")})
			testUpdate(te, om, dns.RcodeSuccess)
		})
		tc.checkRecordDelta(typecounts{"AAAA": 1}, typecounts{})
	})
}

func TestDNSAuthoritative(t *testing.T) {
	testDNS(t, func(te testEnv, z Zone) {
		// Get authoritative SOA.
		tdc := dnsclient{t, &dns.Client{Net: "tcp"}, te.tcpaddr}

		om := msgQuery(z.Name, dns.TypeSOA)
		om.RecursionDesired = false
		im := tdc.exchange(om, nil, dns.RcodeSuccess)
		tcompare(t, len(im.Answer), 1)
		value := strings.TrimPrefix(im.Answer[0].String(), im.Answer[0].Header().String())
		tcompare(t, te.z0.soa.Value, value)
		tcompare(t, im.Authoritative, true)

		// Other zones return NOTAUTH.
		om = msgQuery("bogus.example.", dns.TypeSOA)
		om.RecursionDesired = false
		tdc.exchange(om, nil, dns.RcodeNotAuth)

		// Other types result in SERVFAIL.
		om = msgQuery(z.Name, dns.TypeNS)
		tdc.exchange(om, nil, dns.RcodeServerFailure)
	})
}

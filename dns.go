package main

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/libdns/libdns"

	"github.com/miekg/dns"

	"github.com/mjl-/bstore"
)

// Enabled during tests.
var testSyncNotify = false
var testSyncUpdate = false

var connID atomic.Int64

func init() {
	connID.Store(int64(time.Now().UnixMilli()))
}

// conn is a "connection", either for tcp/tls, or for a single udp packet.
type conn struct {
	cid           int64
	udpRemoteAddr net.Addr // If not nil, this is a udp request and the response should be written to this addr.
	udpconn       net.PacketConn
	conn          net.Conn // Initially tcp connection, replaced with tlsconn after handshake.
	tlsconn       *tls.Conn
	log           *slog.Logger
	listener      listener
	// Verified TLS public key/certificate credential. Credentials are checked for
	// authorization when processing dns requests.
	credTLS *Credential

	// Per request/message fields.
	reqKind       string      // For metric.
	respRcode     int         // For metric.
	credTSIG      *Credential // Authenticated if set.
	tsigIn        *dns.TSIG   // Used for making tsig response signature.
	tsigErrorCode uint16      // For TSIG errors, the code to send in the response.
	notify        bool        // If set, dns notify to zone is sent after operation.
	zone          string      // Absolute name.
	outOpt        *dns.OPT    // If set, added to response, for edns0.
	buf           []byte      // 2+64k, for request & response.
	im            dns.Msg     // Incoming message currently being processed (we handle one at a time).
	nresp         int         // Number of response message for the current message. For multi-message XFR with TSIG.
}

// ServeDNS serves a tcp connection, a loop that reads one request, processes it,
// and writes a response message. No concurrent handling of multiple messages in
// flight on a single connection. Requests on other TCP connections, or with UDP,
// are processed concurrently.
func serveDNS(nc net.Conn, l listener) {
	cid := connID.Add(1)
	c := &conn{
		cid:      cid,
		conn:     nc,
		log:      slog.With("cid", cid),
		listener: l,
		buf:      make([]byte, 2+64*1024),
	}
	defer func() {
		if c.tlsconn != nil {
			// Close quickly, don't wait too long for close alert message being sent/timing out.
			err := c.conn.SetDeadline(time.Now().Add(1 * time.Second))
			logCheck(c.log, err, "setting connection io deadline")
			err = c.tlsconn.Close()
			logCheck(c.log, err, "closing tls connection")
		} else {
			err := nc.Close()
			logCheck(c.log, err, "closing tcp connection")
		}
	}()
	c.log.Debug("new connection", "remoteaddr", nc.RemoteAddr())
	defer c.log.Debug("connection closed")

	// If we are doing TLS, do the handshake explicitly and verify a certificate if
	// present (which is optional).
	if l.tls {
		c.tlsconn = tls.Server(c.conn, &tlsConfig)

		tlsctx, cancel := context.WithTimeout(shutdownCtx, 15*time.Second)
		defer cancel()
		err := c.tlsconn.HandshakeContext(tlsctx)
		if err != nil {
			if errors.Is(err, errUnknownTLSPublicKey) {
				c.log.Debug("tls handshake", "err", err)
			} else {
				c.log.Info("tls handshake", "err", err)
			}
			return
		}
		cancel()

		cs := c.tlsconn.ConnectionState()
		if len(cs.PeerCertificates) > 0 {
			sum := sha256.Sum256(cs.PeerCertificates[0].RawSubjectPublicKeyInfo)
			tlspubkey := base64.RawURLEncoding.EncodeToString(sum[:])

			q := bstore.QueryDB[Credential](shutdownCtx, database)
			q.FilterNonzero(Credential{Type: "tlspubkey", TLSPublicKey: tlspubkey})
			cred, err := q.Get()
			if err != nil {
				c.log.Info("get client certificate, closing connection", "err", err)
				return
			}
			c.credTLS = &cred
			c.log.Debug("tls connection with tls client auth", "credentials", c.credTLS)
		} else {
			c.log.Debug("tls connection without tls client auth")
		}

		c.conn = c.tlsconn
	}

	for {
		// Deadline includes reads and writes for TLS connections.
		err := nc.SetDeadline(time.Now().Add(30 * time.Second))
		logCheck(c.log, err, "setting read deadline")

		// Read dns message size.
		_, err = io.ReadFull(c.conn, c.buf[:2])
		if err != nil {
			if !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
				c.log.Debug("reading tcp dns message size", "err", err)
			}
			return
		}
		// Read dns message.
		size := int(c.buf[0])<<8 | int(c.buf[1])
		n, err := io.ReadFull(c.conn, c.buf[2:2+size])
		if err != nil {
			c.log.Debug("reading tcp dns message", "err", err, "size", size, "got", n)
			return
		}

		// Handle message. For fatal errors, abort the connection.
		ok := c.handleDNS(c.buf[2 : 2+size])
		if !ok {
			break
		}
	}
}

// printTrace prints a trace of the DNS message to standard error, if enabled
// through the -trace flag. The packet is potentially written in multiple formats
// (canonical text format, json formats).
func (c *conn) printTrace(prefix string, m *dns.Msg) {
	if len(serveTraceDNS) > 0 {
		fmt.Fprintf(os.Stderr, "\n%s (cid %x)\n", prefix, c.cid)
		defer fmt.Fprintln(os.Stderr)
	}
	for _, t := range serveTraceDNS {
		var s string
		switch t {
		case traceNone:
			return
		case traceText:
			s = m.String()
		case traceJSON:
			buf, err := json.Marshal(m)
			if err != nil {
				s = "trace"
			} else {
				s = string(buf)
			}
		case traceJSONIndent:
			s = describe(m)
		}
		fmt.Fprintln(os.Stderr, s)
	}
}

// respond with a "SERVFAIL" error and extended "other error" code and error message.
func (c *conn) respondErrorf(format string, args ...any) (ok bool) {
	return c.respondExtErrorf(dns.RcodeServerFailure, dns.ExtendedErrorCodeOther, format, args...)
}

// respond with the provided rcode and "other error" extended code and a message.
func (c *conn) respondCodeErrorf(rcode int, format string, args ...any) (ok bool) {
	return c.respondExtErrorf(rcode, dns.ExtendedErrorCodeOther, format, args...)
}

// respond with error rcode and extended error code and error message.
func (c *conn) respondExtErrorf(rcode int, exterrcode uint16, format string, args ...any) (ok bool) {
	var xm dns.Msg
	om := xm.SetRcode(&c.im, rcode)
	om.Authoritative = true
	om.AuthenticatedData = false
	msg := fmt.Sprintf(format, args...)
	if c.outOpt != nil {
		c.outOpt.Option = append(c.outOpt.Option, &dns.EDNS0_EDE{InfoCode: exterrcode, ExtraText: msg})
	}
	c.log.Debug("error response", "msg", msg, "rcode", rcode, "exterrcode", exterrcode)
	return c.respond(om)
}

// respond to request message, and potentially schedule a DNS NOTIFY in case of the
// request resulting in changes to stored records.
func (c *conn) respond(om *dns.Msg) (ok bool) {
	c.respRcode = om.Rcode

	// If we need to notify, we do it after having written the respond message to the
	// requester. If we were to schedule the NOTIFY immediately when the change was
	// made, there could potentially be a race where the requester tries to get the
	// changed result that isn't committed/available for new requests yet.
	if c.notify && c.zone != "" {
		zone := c.zone
		c.zone = ""
		c.notify = false
		go func() {
			defer recoverPanic(c.log, "sending dns notifications for zone")
			sendZoneNotify(c.log, zone)
		}()
	}

	// Add OPT to outgoing responses (if it was present on incoming request).
	if c.outOpt != nil {
		if opt := om.IsEdns0(); opt != nil {
			panic("edns0 already set on response?")
		}
		om.Extra = append(om.Extra, c.outOpt)
	}

	// Add TSIG to response if we have verified TSIG credentials.
	var osize int
	if c.tsigIn != nil && c.credTSIG != nil {
		tsigOut := &dns.TSIG{
			Hdr:        c.tsigIn.Hdr,
			Algorithm:  c.tsigIn.Algorithm,
			TimeSigned: uint64(time.Now().Unix()),
			Fudge:      300,
			OrigId:     c.tsigIn.OrigId,
			Error:      c.tsigErrorCode,
		}
		om.Extra = append(om.Extra, tsigOut)

		// rfc/8945:591 Package dns does not generate a signature when tsigOut.Error is
		// RcodeBadSig. For RcodeBadSig, we would not have valid c.credTSIG, so would not
		// get here.
		// rfc/8945:605 We pass in the mac from the incoming message, which we replace
		// below with the newly generated mac, to chain in case of multiple response
		// messages.
		outbuf, mac, err := dns.TsigGenerate(om, c.credTSIG.TSIGSecret, c.tsigIn.MAC, c.nresp > 0)
		if err != nil {
			// On errors, package dns will already have removed the TSIG record from
			// om.Extra...
			c.tsigIn = nil // Prevent TSIG record being added again.
			return c.respondErrorf("generating tsig response: %v", err)
		}
		// Multiple response messages, for AXFR, use the previous response MAC as input for
		// calculating the signature.
		c.nresp++
		c.tsigIn.MAC = mac
		osize = len(outbuf)

		c.printTrace("# >>> outgoing dns response", om)

		c.buf[0] = byte(osize >> 8)
		c.buf[1] = byte(osize)
		copy(c.buf[2:], outbuf) // Should always fit.
	} else {
		osize = om.Len()

		c.printTrace("# >>> outgoing dns response", om)

		c.buf[0] = byte(osize >> 8)
		c.buf[1] = byte(osize)
		if _, err := om.PackBuffer(c.buf[2:]); err != nil {
			c.log.Debug("packing dns response, aborting connection", "err", err)
			return false
		}
	}

	c.log.Debug("writing response packet", "size", osize)
	if c.udpRemoteAddr != nil {
		_, err := c.udpconn.WriteTo(c.buf[2:2+osize], c.udpRemoteAddr)
		if err != nil {
			c.log.Debug("writing dns response", "err", err)
			return false
		}
	} else {
		// Deadline includes both writes and reads, for TLS connections.
		err := c.conn.SetDeadline(time.Now().Add(30 * time.Second))
		logCheck(c.log, err, "setting write deadline")
		if _, err := c.conn.Write(c.buf[:2+osize]); err != nil {
			c.log.Debug("writing dns response, aborting connection", "err", err)
			return false
		}
	}
	return true
}

// handleDNS handles a single dns request message, either from UDP or TCP. imbuf
// does not include the TCP packet size prefix. If handleDNS returns false (for
// fatal connection errors, like unparsable packet), the TCP connection should be
// dropped.
func (c *conn) handleDNS(imbuf []byte) (ok bool) {
	c.reqKind = "n/a"
	c.respRcode = -1
	defer func() {
		var rcodestr string
		if s, ok := dns.RcodeToString[c.respRcode]; ok {
			rcodestr = strings.ToLower(s)
		} else {
			rcodestr = "other"
		}
		metricDNSRequests.WithLabelValues(c.reqKind, rcodestr).Inc()
	}()

	// Reset per-request/message fields.
	c.credTSIG = nil
	c.tsigIn = nil
	c.tsigErrorCode = 0
	c.notify = false
	c.zone = ""
	c.outOpt = nil
	c.im = dns.Msg{}
	c.nresp = 0
	if err := c.im.Unpack(imbuf); err != nil {
		c.log.Debug("parsing dns message, aborting connection", "err", err)
		return false
	}

	c.printTrace("# <<< incoming dns request", &c.im)

	if c.im.Response {
		if c.udpRemoteAddr != nil {
			// Not responding to potential misdirected response.
			return false
		}
		c.respondCodeErrorf(dns.RcodeFormatError, "only dns requests allowed")
		return false // Drop connection.
	}
	if c.im.Rcode != 0 {
		return c.respondCodeErrorf(dns.RcodeFormatError, "rcode must be zero")
	}
	if c.im.Truncated {
		return c.respondCodeErrorf(dns.RcodeFormatError, "do not ask truncated questions")
	}

	// Handle EDNS for future versions (> 0).
	// note: Package dns returns any OPT record, not only version 0 of edns.
	opt := c.im.IsEdns0()
	if opt != nil {
		c.outOpt = &dns.OPT{
			Hdr:    dns.RR_Header{Name: ".", Rrtype: dns.TypeOPT},
			Option: []dns.EDNS0{},
		}
		// 1232 is recommended since the dns edns0 flag day.
		c.outOpt.SetUDPSize(1232)
		if opt.Version() != 0 {
			// rfc/8906:312
			return c.respondCodeErrorf(dns.RcodeBadVers, "dns eopt with version %d not supported (only edns0)", opt.Version())
		}
	}

	// Check whether TSIG is present, and ensure it is the only and last record.
	for i, rr := range c.im.Extra {
		if tsig, ok := rr.(*dns.TSIG); ok && i != len(c.im.Extra)-1 {
			// rfc/8945:472
			return c.respondCodeErrorf(dns.RcodeFormatError, "tsig must be last extra record")
		} else if ok {
			c.tsigIn = tsig
		}
	}

	// Check TSIG authentication. Authorization is checked later.
	if c.tsigIn != nil {
		ctx := shutdownCtx
		var cred Credential
		var err error
		err = database.Read(ctx, func(tx *bstore.Tx) error {
			q := bstore.QueryTx[Credential](tx)
			q.FilterNonzero(Credential{Name: strings.TrimSuffix(c.tsigIn.Hdr.Name, ".")})
			cred, err = q.Get()
			if err == nil && cred.Type != "tsig" {
				err = fmt.Errorf("not a tsig key: %w", bstore.ErrAbsent)
			}
			return err
		})
		if err != nil && errors.Is(err, bstore.ErrAbsent) {
			// todo: since we didn't set credTSIG yet, we are currently not responding with a TSIG RR.
			// rfc/8945:496
			c.tsigErrorCode = dns.RcodeBadKey
			return c.respondCodeErrorf(dns.RcodeNotAuth, "unknown key")
		} else if err != nil {
			return c.respondErrorf("checking tsig: %v", err)
		}
		c.credTSIG = &cred
		// Package dns implements hmac-sha1 and later, not hmac-md5. So we don't check
		// which hmac is used. Package dns always checks the received mac against the full
		// mac it calculated, resulting in RcodeSig. Should be dns.RcodeBadTrunc... rfc/8945:582
		if err := dns.TsigVerify(imbuf, cred.TSIGSecret, "", false); err != nil {
			if errors.Is(err, dns.ErrTime) {
				// rfc/8945:506 Package dns ensures the response isn't tsig-signed.
				c.tsigErrorCode = dns.RcodeBadTime
			} else if errors.Is(err, dns.ErrSig) {
				// rfc/8945:553 Package dns ensures the response isn't tsig-signed.
				c.tsigErrorCode = dns.RcodeBadSig
			}
			// TsigVerify can return dns.ErrKeyAlg for unsupported mac algorithm, but there is
			// no specific tsig error code to return.
			return c.respondCodeErrorf(dns.RcodeNotAuth, "verifying tsig: %v", err)
		}
	}

	if len(c.im.Question) != 1 {
		return c.respondCodeErrorf(dns.RcodeFormatError, "request must have 1 question, not %d", len(c.im.Question))
	}
	q := c.im.Question[0]

	// We allow CHAOS queries, for returning our version.
	switch {
	case q.Qclass == dns.ClassINET,
		q.Qclass == dns.ClassCHAOS && c.listener.auth && c.im.Opcode == dns.OpcodeQuery:
	default:
		return c.respondCodeErrorf(dns.RcodeRefused, "only class inet allowed")
	}

	if c.listener.notify && c.im.Opcode == dns.OpcodeNotify {
		c.reqKind = "notify"
		return c.handleNotify(shutdownCtx)
	} else if c.listener.updates && c.im.Opcode == dns.OpcodeUpdate {
		c.reqKind = "update"
		return c.handleUpdate(shutdownCtx)
	} else if c.listener.xfr && c.im.Opcode == dns.OpcodeQuery && q.Qtype == dns.TypeAXFR {
		c.reqKind = "axfr"
		return c.handleXFR(shutdownCtx)
	} else if c.listener.auth && c.im.Opcode == dns.OpcodeQuery {
		c.reqKind = "authoritative"
		// We serve "authoritative" queries for SOA. For AXFR clients that check if they
		// are up to date before initiating the transfer.
		return c.handleAuth(shutdownCtx)
	} else {
		c.reqKind = "other"
		// rfc/8906:268
		return c.respondExtErrorf(dns.RcodeNotImplemented, dns.ExtendedErrorCodeNotSupported, "request not implemented")
	}
}

// DNS UPDATE and AXFR require authentication (checked early when processing the
// packet), and the credentials must be authorized for the operation (checking
// while processing the request).
var errAuthcRequired = errors.New("tls public key and/or tsig authentication required")
var errPermission = errors.New("permission denied")

func verifyZoneCredentials(tx *bstore.Tx, zoneName string, credTLS, credTSIG *Credential) error {
	if credTLS == nil && credTSIG == nil {
		return errAuthcRequired
	}

	if credTLS != nil {
		q := bstore.QueryTx[ZoneCredential](tx)
		q.FilterNonzero(ZoneCredential{Zone: zoneName, CredentialID: credTLS.ID})
		_, err := q.Get()
		if err == bstore.ErrAbsent {
			return fmt.Errorf("%w: tls public key not authorized for this zone", errPermission)
		} else if err != nil {
			return fmt.Errorf("verifying tls public key: %v", err)
		}
	}

	if credTSIG != nil {
		q := bstore.QueryTx[ZoneCredential](tx)
		q.FilterNonzero(ZoneCredential{Zone: zoneName, CredentialID: credTSIG.ID})
		_, err := q.Get()
		if err == bstore.ErrAbsent {
			return fmt.Errorf("%w: tsig key not authorized for this zone", errPermission)
		} else if err != nil {
			return fmt.Errorf("verifying tsig key: %v", err)
		}
	}

	return nil
}

// DNS NOTIFY requests cause us to do an immediate check for SOA freshness. If the
// request message has a SOA record, we check it against what we have. If it is
// already the same, we don't have to do anything. Otherwise, do a full sync.
func (c *conn) handleNotify(ctx context.Context) (ok bool) {
	if len(c.im.Question) != 1 {
		return c.respondCodeErrorf(dns.RcodeFormatError, "exactly 1 question required")
	}

	// rfc/1996:178 We ignore any authoritative/ns and data/extra sections.

	// No authorization for now. Could require it in the future, or based on ip
	// address. We are not trusting anything the request says.

	// todo: rfc/1996:186 we could have an IP-based allowlist for processing dns notify messages.

	var z Zone
	var provider Provider
	var soa *Record
	err := database.Read(ctx, func(tx *bstore.Tx) error {
		var err error
		z, provider, err = zoneProvider(tx, c.im.Question[0].Name)
		if err != nil {
			return err
		}
		q := bstore.QueryTx[Record](tx)
		q.FilterNonzero(Record{Zone: z.Name, AbsName: z.Name, Type: Type(dns.TypeSOA)})
		q.FilterFn(func(r Record) bool { return r.Deleted == nil })
		rr, err := q.Get()
		if err == bstore.ErrAbsent {
			return nil // No SOA yet, not necessarily an error.
		} else if err != nil {
			return fmt.Errorf("lookup local soa: %v", err)
		}
		soa = &rr
		return nil
	})
	if err != nil && err == bstore.ErrAbsent {
		return c.respondExtErrorf(dns.RcodeNotAuth, dns.ExtendedErrorCodeNotAuthoritative, "unknown zone")
	} else if err != nil {
		return c.respondErrorf("get zone and provider: %v", err)
	}

	// rfc/1996:157
	if len(c.im.Answer) == 1 {
		if nsoa, ok := c.im.Answer[0].(*dns.SOA); ok && soa != nil && Serial(nsoa.Serial) == soa.SerialFirst {
			c.log.Debug("received dns notify with soa record with serial we already have")
			var xm dns.Msg
			om := xm.SetRcode(&c.im, dns.RcodeSuccess)
			om.Authoritative = true
			om.AuthenticatedData = false
			return c.respond(om)
		}
	}

	done := make(chan struct{}, 1)

	go func() {
		defer recoverPanic(c.log, "syncing zone after dns notify")
		defer func() {
			done <- struct{}{}
		}()

		// todo: rfc/1996:198 instead of requesting the full list of records, we could be querying the SOA record directly at the authoritative servers, and only sync when it has changed.
		// todo: rfc/1996:265 we should only have a single sync in flight at a time.

		log := c.log.With("notifyzone", z.Name)

		unlock := lockZone(z.Name)
		defer unlock()

		ctx, cancel := context.WithTimeout(shutdownCtx, 30*time.Second)
		defer cancel()
		latest, err := getRecords(ctx, c.log, provider, z.Name, false)
		if err != nil {
			log.Error("get records from provider", "err", err)
			return
		}

		var notify bool
		defer possiblyZoneNotify(log, z.Name, &notify)

		c.zone = z.Name
		err = database.Write(ctx, func(tx *bstore.Tx) error {
			notify, _, _, _, err = syncRecords(log, tx, z, latest)
			return err
		})
		if err != nil {
			log.Error("updating records", "err", err)
		}
	}()

	// Hook for testing.
	if testSyncNotify {
		<-done
	}

	var xm dns.Msg
	om := xm.SetRcode(&c.im, dns.RcodeSuccess)
	om.Authoritative = true
	om.AuthenticatedData = false
	return c.respond(om)
}

// DNS UPDATE requests reference the zone they operate on in the "question" name.
// The answer section holds prerequisites (records/rrsets/names that must (not)
// exist that we check before making changes. The authoritative/ns section holds
// the records to be added, or name or name+type or name+type+data that needs to be
// removed. We always sync with remote before evaluating prerequisites. Then we
// prepare the changes we're going to send to remote. The changes are supposed to
// be atomic, but the libdns API (and likely most underlying cloud APIs) does not
// allow for atomic additions/deletions. After making the changes, in the
// background we check that the changes were propagated as expected, refreshing
// records from the provider a few times as needed.
func (c *conn) handleUpdate(ctx context.Context) (ok bool) {
	if len(c.im.Question) != 1 || c.im.Question[0].Qtype != dns.TypeSOA {
		return c.respondCodeErrorf(dns.RcodeFormatError, "exactly 1 soa question needed")
	}

	var z Zone
	var provider Provider
	var soa Record
	err := database.Read(ctx, func(tx *bstore.Tx) error {
		if err := verifyZoneCredentials(tx, c.im.Question[0].Name, c.credTLS, c.credTSIG); err != nil {
			return err
		}

		var err error
		z, provider, err = zoneProvider(tx, c.im.Question[0].Name)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil && err == bstore.ErrAbsent {
		// rfc/2136:544
		return c.respondExtErrorf(dns.RcodeNotAuth, dns.ExtendedErrorCodeNotAuthoritative, "unknown zone")
	} else if err != nil && (errors.Is(err, errAuthcRequired) || errors.Is(err, errPermission)) {
		// rfc/8914:349
		return c.respondExtErrorf(dns.RcodeRefused, dns.ExtendedErrorCodeProhibited, "%v", err)
	} else if err != nil {
		return c.respondErrorf("get zone and provider: %v", err)
	}

	unlock := lockZone(z.Name)
	defer func() {
		// May have been cleared when passing control over to ensurePropagate.
		if unlock != nil {
			unlock()
		}
	}()

	// Sync latest zone before attempting to make any changes.
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	latest, err := getRecords(ctx, c.log, provider, z.Name, false)
	if err != nil {
		return c.respondExtErrorf(dns.RcodeServerFailure, dns.ExtendedErrorCodeNetworkError, "get records from provider: %v", err)
	}

	// We keep these up to date while removing/adding records. So logic like "remove
	// all records for a given name" and various checks on current records (eg CNAME,
	// last NS) work.
	known := map[recordKey]Record{}
	rrsets := map[rrsetKey][]Record{}
	adjustAdd := func(r Record) {
		known[r.recordKey()] = r
		k := r.rrsetKey()
		rrsets[k] = append(rrsets[k], r)
	}
	adjustDel := func(r Record) {
		delete(known, r.recordKey())
		k := r.rrsetKey()
		l := rrsets[k]
		for i := 0; i < len(l); i++ {
			if l[i].recordKey() == r.recordKey() {
				copy(l[i:], l[i+1:])
				l = l[:len(l)-1]
				if len(l) == 0 {
					delete(rrsets, k)
				} else {
					rrsets[k] = l
				}
				return
			}
		}
		panic("adjustDel: no record deleted?")
	}

	c.zone = z.Name // Used along with c.notify
	err = database.Write(ctx, func(tx *bstore.Tx) error {
		c.notify, _, _, _, err = syncRecords(c.log, tx, z, latest)
		if err != nil {
			return err
		}

		soa = zoneSOA(c.log, tx, z.Name)

		q := bstore.QueryTx[Record](tx)
		q.FilterNonzero(Record{Zone: z.Name})
		q.FilterFn(func(r Record) bool { return r.Deleted == nil })
		q.FilterNotEqual("Type", Type(dns.TypeSOA))
		current, err := q.List()
		if err != nil {
			return fmt.Errorf("list current records: %w", err)
		}

		for _, r := range current {
			adjustAdd(r)
		}

		return nil
	})
	if err != nil {
		return c.respondErrorf("ensuring records are fresh: %v", err)
	}

	// For checking as a group (cannot be check individually).
	rrsetsCheck := map[rrsetKey][]Record{}

	// rfc/2136:324 Description
	// rfc/2136:623 Pseudocode
	// Check prerequisites.
	for _, r := range c.im.Answer {
		h := r.Header()
		if h.Ttl != 0 {
			return c.respondCodeErrorf(dns.RcodeFormatError, "ttl of prerequisites must be 0")
		}

		name, err := cleanAbsName(h.Name)
		if err != nil {
			return c.respondCodeErrorf(dns.RcodeFormatError, "bad name %s", h.Name)
		}

		if !(name == z.Name || strings.HasSuffix(name, "."+z.Name)) {
			// rfc/2136:552
			return c.respondCodeErrorf(dns.RcodeNotZone, "name must be in zone")
		}

		var ok bool
		var rcode int
		if h.Class == dns.ClassANY {
			if h.Rdlength != 0 {
				// rfc/2136:568
				return c.respondCodeErrorf(dns.RcodeFormatError, "prereq with class any must have rr with rdlength 0 != %d", h.Rdlength)
			}
			// Exists/in use.
			ok = false
			rcode = dns.RcodeNXRrset // rfc/2136:573
			if h.Rrtype == dns.TypeANY {
				rcode = dns.RcodeNameError // rfc/2136:571
			}
			for _, cr := range known {
				if cr.AbsName == name && (h.Rrtype == dns.TypeANY || Type(h.Rrtype) == cr.Type) {
					ok = true
					break
				}
			}
		} else if h.Class == dns.ClassNONE {
			if h.Rdlength != 0 {
				// rfc/2136:577
				return c.respondCodeErrorf(dns.RcodeFormatError, "prereq with class any must have rr with rdlength 0 != %d", h.Rdlength)
			}
			// Rrset does not exist/name not in use.
			ok = true
			rcode = dns.RcodeYXRrset // rfc/2136:582
			if h.Rrtype == dns.TypeANY {
				rcode = dns.RcodeYXDomain // rfc/2136:580
			}
			for _, cr := range known {
				if cr.AbsName == name && (h.Rrtype == dns.TypeANY || Type(h.Rrtype) == cr.Type) {
					ok = false
					break
				}
			}
		} else {
			// rfc/2136:591
			if h.Class != dns.ClassINET {
				return c.respondCodeErrorf(dns.RcodeFormatError, "class must be inet")
			}

			hex, value, err := recordData(r)
			if err != nil {
				return c.respondErrorf("parsing record for prerequisite comparison: %v", err)
			}

			r := Record{0, z.Name, 0, 0, time.Time{}, nil, name, Type(h.Rrtype), Class(dns.ClassINET), TTL(0), hex, value, ""}
			k := r.rrsetKey()
			rrsetsCheck[k] = append(rrsetsCheck[k], r)
			ok = true // Checked later.
		}

		if !ok {
			return c.respondCodeErrorf(rcode, "prerequisite failed")
		}
	}

	// Compare if two rrsets are equal, not taking TTL into account.
	rrsetEqual := func(a, b []Record) bool {
		if len(a) != len(b) {
			return false
		}
		akeys := map[recordKey]int{}
		bkeys := map[recordKey]int{}
		for _, e := range a {
			e.TTL = 0
			akeys[e.recordKey()]++
		}
		for _, e := range b {
			e.TTL = 0
			bkeys[e.recordKey()]++
		}
		for k, n := range akeys {
			if n != bkeys[k] {
				return false
			}
		}
		return true
	}

	// rfc/2136:590
	for k, rrset := range rrsetsCheck {
		if !rrsetEqual(rrset, rrsets[k]) {
			return c.respondCodeErrorf(dns.RcodeNXRrset, "prerequisite failed for %v", k)
		}
	}

	c.log.Debug("dns update prerequisites are ok")

	// rfc/2136:664 todo: we could implement an acl with rules which records these credentials are allowed to update. if not allowed, we respond with REFUSED and extended code Prohibited.

	// rfc/2136:40 DNS UPDATE is supposed be atomic, but that's not possible with the
	// libdns API (and likely with the underlying APIs). We could try to rollback
	// changes we've made after an error, but that's error prone too. We'll leave it as
	// a limitation.

	var add, set, remove []Record
	for _, rr := range c.im.Ns {
		h := rr.Header()

		// rfc/2136:704
		switch h.Class {
		case dns.ClassANY, dns.ClassNONE, dns.ClassINET:
		default:
			return c.respondCodeErrorf(dns.RcodeFormatError, "can only add records with class INET")
		}

		name, err := cleanAbsName(h.Name)
		if err != nil {
			return c.respondCodeErrorf(dns.RcodeFormatError, "bad name %s", h.Name)
		}
		// rfc/2136:706
		if !(name == z.Name || strings.HasSuffix(name, "."+z.Name)) {
			return c.respondCodeErrorf(dns.RcodeNotZone, "name must be in zone")
		}

		// rfc/2136:709
		switch h.Rrtype {
		case dns.TypeNone, dns.TypeAXFR, dns.TypeIXFR, dns.TypeMAILA, dns.TypeMAILB, dns.TypeTKEY, dns.TypeNXNAME:
			return c.respondCodeErrorf(dns.RcodeFormatError, "meta record types not allowed")
		}
		if h.Class != dns.ClassANY && h.Rrtype == dns.TypeANY {
			return c.respondCodeErrorf(dns.RcodeFormatError, "record type any not allowed for class other than any")
		}

		if h.Class == dns.ClassANY {
			if h.Ttl != 0 {
				return c.respondCodeErrorf(dns.RcodeFormatError, "ttl must be zero for class any")
			}
			if h.Rdlength != 0 {
				return c.respondCodeErrorf(dns.RcodeFormatError, "rdlength must be zero for class any")
			}
			for _, cr := range known {
				// Delete All RRsets From A Name, or Delete an RRset.
				// rfc/2136:777 ANY deletes all, except when name is zone and type SOA or NS.
				// todo: should we also not delete dnssec-signing records?
				if name == cr.AbsName && (h.Rrtype == dns.TypeANY && (name != z.Name || cr.Type != Type(dns.TypeSOA) && cr.Type != Type(dns.TypeNS)) || Type(h.Rrtype) == cr.Type) {
					remove = append(remove, cr)
					adjustDel(cr)
				}
			}
			continue
		}

		r := Record{0, z.Name, 0, 0, time.Time{}, nil, name, Type(h.Rrtype), Class(dns.ClassINET), TTL(h.Ttl), "", "", ""}

		hex, value, err := recordData(rr)
		if err != nil {
			return c.respondErrorf("parsing record to add/delete: %v", err)
		}
		r.DataHex = hex
		r.Value = value

		c.log.Debug("looking to add/remove record", "record", r)

		switch h.Class {
		case dns.ClassNONE:
			// rfc/2136:793 Deleting SOA for zone is ignored.
			if r.AbsName == z.Name && r.Type == Type(dns.TypeSOA) {
				c.log.Debug("removing soa for zone is ignored", "delrecord", r)
				continue
			}
			// rfc/2136:779 With 1 NS remaining for zone, attempts to delete the last are ignored.
			if r.AbsName == z.Name && r.Type == Type(dns.TypeNS) && len(rrsets[rrsetKey{r.AbsName, r.Type, r.Class}]) == 1 {
				c.log.Debug("removing last ns record for zone is ignored", "delrecord", r)
				continue
			}

			// Remove all records that match name,type,data (TTL ignored, so we cannot look up in "known").
			for _, cr := range known {
				if cr.AbsName == r.AbsName && cr.Type == r.Type && cr.DataHex == r.DataHex {
					// Delete An RR From An RRset
					remove = append(remove, cr)
					adjustDel(cr)
				}
			}

		case dns.ClassINET:
			// If a record already exists exactly as is, we don't make any changes.
			if _, ok := known[r.recordKey()]; ok {
				continue
			}

			if r.Type == Type(dns.TypeSOA) {
				// todo: we could try implementing setting a new soa. would have to check with libdns providers if they implement it.
				return c.respondCodeErrorf(dns.RcodeRefused, "setting soa not implemented")
			}

			if r.Type != Type(dns.TypeCNAME) {
				// If CNAME for this name already exists, ignore new records.
				if len(rrsets[rrsetKey{r.AbsName, Type(dns.TypeCNAME), r.Class}]) > 0 {
					c.log.Info("attempt to add record for name that has a cname")
					continue
				}

				add = append(add, r)
				adjustAdd(r)
				continue
			}

			// rfc/2136:119 For CNAME, only a single value can exist. If a CNAME already
			// exists, we "set" it, which should overwrite the current record.
			// rfc/2136:769 For a CNAME update, if a non-CNAME exists, we ignore the update.
			rrset := rrsets[rrsetKey{r.AbsName, r.Type, r.Class}]
			if len(rrset) > 0 {
				r.ProviderID = rrset[0].ProviderID
				set = append(set, r)
				for _, d := range rrset {
					adjustDel(d)
				}
				adjustAdd(r)
			} else {
				ignore := false
				for _, cr := range known {
					if cr.AbsName == r.AbsName {
						ignore = true
						break
					}
				}
				if !ignore {
					add = append(add, r)
					adjustAdd(r)
				}
			}

		default:
			panic("missing case for class")
		}
	}

	// todo: it may be better to batch adds/sets and deletes separately, and potentially do multiple of them. eg when update requests to add a request which it then deletes. we currently first try to delete it, then add it. hopefully sane clients never do that.

	c.log.Debug("adding/setting/removing records", "add", add, "set", set, "remove", remove)

	var added, xset, removed []libdns.Record
	if len(add) > 0 {
		added, err = appendRecords(ctx, c.log, provider, z.Name, libdnsRecords(add))
		if err != nil {
			return c.respondErrorf("adding records: %v", err)
		}
	}
	if len(set) > 0 {
		xset, err = setRecords(ctx, c.log, provider, z.Name, libdnsRecords(set))
		if err != nil {
			return c.respondErrorf("setting records: %v", err)
		}
	}
	if len(remove) > 0 {
		removed, err = deleteRecords(ctx, c.log, provider, z.Name, libdnsRecords(remove))
		if err != nil {
			return c.respondErrorf("removing records: %v", err)
		}
	}
	c.log.Debug("records added/set/removed", "added", added, "set", xset, "removed", removed)

	done := make(chan struct{}, 1)

	xunlock := unlock
	unlock = nil
	go func() {
		defer recoverPanic(c.log, "ensuring updated zone after dns update")
		defer xunlock()
		defer func() {
			done <- struct{}{}
		}()

		adds := make([]recordKey, len(add))
		for i, a := range add {
			adds[i] = a.recordKey()
		}

		_, _, err := ensurePropagate(shutdownCtx, c.log, provider, z, adds, remove, soa.SerialFirst)
		if err != nil {
			c.log.Error("ensuring propagation of dns update", "err", err)
		}
	}()

	if testSyncUpdate {
		<-done
	}

	var xm dns.Msg
	om := xm.SetRcode(&c.im, dns.RcodeSuccess)
	om.Authoritative = true
	om.AuthenticatedData = false
	return c.respond(om)
}

// Handle XFR (AXFR only for now). We will send the full zone. We start and end
// with the SOA record. We may be sending multiple messages (each DNS message is
// max 64KB), each potentially TSIG signed.
func (c *conn) handleXFR(ctx context.Context) (ok bool) {
	// rfc/5936:439
	if len(c.im.Answer) != 0 || len(c.im.Ns) != 0 {
		return c.respondCodeErrorf(dns.RcodeFormatError, "answer and authority section must be empty for xfr")
	}

	// rfc/5936:532 We don't check the RRs in the additional section. We already
	// handled TSIG records. The rest are ignored for now. Perhaps we should become
	// more strict in the future.

	// todo: for ixfr, we would require an auth/ns section with 1 record, a soa. we would be bringing the client up to date from their current serial to the next serial by sending deletions before the new soa, and additions followed by the soa again after the new serial. rfc/1995

	// Get zone and verify credentials have access.
	var z Zone
	var provider Provider
	err := database.Read(ctx, func(tx *bstore.Tx) error {
		if err := verifyZoneCredentials(tx, c.im.Question[0].Name, c.credTLS, c.credTSIG); err != nil {
			return err
		}

		var err error
		z, provider, err = zoneProvider(tx, c.im.Question[0].Name)
		return err
	})
	if err != nil && err == bstore.ErrAbsent {
		return c.respondExtErrorf(dns.RcodeNotAuth, dns.ExtendedErrorCodeNotAuthoritative, "unknown zone") // rfc/5936:717
	} else if err != nil && (errors.Is(err, errAuthcRequired) || errors.Is(err, errPermission)) {
		// rfc/8914:349
		return c.respondExtErrorf(dns.RcodeRefused, dns.ExtendedErrorCodeProhibited, "%v", err)
	} else if err != nil {
		return c.respondErrorf("get zone and provider: %v", err)
	}

	unlock := lockZone(z.Name)
	defer unlock()

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	latest, err := getRecords(ctx, c.log, provider, z.Name, false)
	if err != nil {
		return c.respondExtErrorf(dns.RcodeServerFailure, dns.ExtendedErrorCodeNetworkError, "get records from provider: %v", err)
	}

	c.zone = z.Name // Used along with c.notify.
	var latestSOA *Record
	var current []Record
	err = database.Write(ctx, func(tx *bstore.Tx) error {
		c.notify, latestSOA, _, _, err = syncRecords(c.log, tx, z, latest)
		if err != nil {
			return err
		}

		q := bstore.QueryTx[Record](tx)
		q.FilterNonzero(Record{Zone: z.Name})
		q.FilterFn(func(r Record) bool { return r.Deleted == nil })
		q.FilterNotEqual("Type", Type(dns.TypeSOA))
		current, err = q.List()
		if err != nil {
			return fmt.Errorf("list current records: %w", err)
		}

		return nil
	})
	if err != nil {
		return c.respondErrorf("%v", err)
	}

	c.log.Debug("current", "records", current)

	soa, err := latestSOA.RR()
	if err != nil {
		return c.respondErrorf("soa rr: %v", err)
	}

	// rfc/5936:589 Prepare the full response records first. We may have to write
	// multiple output messages. We start and end with the SOA record.
	answer := make([]dns.RR, 2+len(current))
	answer[0] = soa
	for i := 0; i < len(current); i++ {
		rr, err := current[i].RR()
		if err != nil {
			return c.respondErrorf("db record rr: %v", err)
		}
		answer[1+i] = rr
	}
	answer[len(answer)-1] = soa

	// todo: rfc/5936:1034 we always lower-case domain names for convenience of implementation (looking up records by name in the database). we could also store the original case and use it in axfr, but probably not worth the trouble.

	for len(answer) > 0 {
		var xm dns.Msg
		om := xm.SetReply(&c.im)
		om.Authoritative = true // rfc/5936:696 Mark authoritative. We don't have error conditions anymore.
		om.AuthenticatedData = false

		// rfc/5936:611 Fill up the message with answers to a reasonable extent. The TSIG
		// record is added later, when writing the response. We could try a bit harder to
		// determine the TSIG RR size, but we'll just leave enough room for now.
		use := len(answer)
		om.Answer = answer[:use]
		const maxSize = 64*1024 - 512 // Some slack for possible extra TSIG record.
		for size := om.Len(); size > maxSize; size = om.Len() {
			nuse := maxSize * use / size
			if nuse >= use {
				use--
			} else {
				use = nuse
			}
			om.Answer = answer[:use]
		}
		if use == 0 {
			c.log.Error("internal error, no records fit in axfr response")
			return false
		}
		answer = answer[use:]
		if ok := c.respond(om); !ok {
			return false
		}
	}
	return true
}

// For regular dns queries for authoritative data. We only answer requests for SOA
// records or CHAOS version.bind. Useful because XFR clients may check if they are
// up to date before initiating a full zone transfer.
func (c *conn) handleAuth(ctx context.Context) (ok bool) {
	q := c.im.Question[0]

	if len(c.im.Answer) != 0 || len(c.im.Ns) != 0 {
		return c.respondCodeErrorf(dns.RcodeFormatError, "answer and authority section must be empty")
	}

	if q.Qclass == dns.ClassCHAOS {
		if q.Qtype != dns.TypeTXT || q.Name != "version.bind." {
			return c.respondCodeErrorf(dns.RcodeRefused, "only the version.bind txt can be requested for chaos")
		}

		txt := dns.TXT{
			Hdr: dns.RR_Header{Name: "version.bind.", Rrtype: dns.TypeTXT, Class: dns.ClassCHAOS},
			Txt: []string{version},
		}
		var xm dns.Msg
		om := xm.SetReply(&c.im)
		om.Authoritative = true
		om.AuthenticatedData = false
		om.Answer = []dns.RR{&txt}
		return c.respond(om)
	}

	// rfc/8906:234 We should respond with NOERROR/NXDOMAIN, but we don't want to
	// mislead. Better tell clients something is wrong.
	if q.Qtype != dns.TypeSOA {
		return c.respondErrorf("only soa records can be requested")
	}

	// Get zone & SOA. The found zone may be for a parent name, so we can return
	// NXDOMAIN to the request instead of NOTAUTH.
	var z Zone
	var soa Record
	err := database.Read(ctx, func(tx *bstore.Tx) error {
		name := strings.ToLower(q.Name)
		for {
			z = Zone{Name: name}
			if err := tx.Get(&z); err == nil {
				q := bstore.QueryTx[Record](tx)
				q.FilterNonzero(Record{Zone: z.Name, AbsName: z.Name, Type: Type(dns.TypeSOA)})
				q.FilterFn(func(r Record) bool { return r.Deleted == nil })
				soa, err = q.Get()
				if err != nil {
					return fmt.Errorf("get soa record for zone: %v", err)
				}
				return nil
			} else if err != bstore.ErrAbsent {
				return err
			}
			t := strings.SplitN(name, ".", 2)
			if len(t) == 1 || t[1] == "" {
				break
			}
			name = t[1]
		}
		return bstore.ErrAbsent
	})
	if err != nil && err == bstore.ErrAbsent {
		return c.respondExtErrorf(dns.RcodeNotAuth, dns.ExtendedErrorCodeNotAuthoritative, "unknown zone")
	} else if err != nil {
		return c.respondErrorf("get zone and soa: %v", err)
	} else if !strings.EqualFold(z.Name, q.Name) {
		return c.respondCodeErrorf(dns.RcodeNameError, "no soa record for this subdomain")
	}

	soarr, err := soa.RR()
	if err != nil {
		return c.respondErrorf("making soa for zone: %v", err)
	}

	var xm dns.Msg
	om := xm.SetReply(&c.im)
	om.Authoritative = true
	om.AuthenticatedData = false
	om.Answer = []dns.RR{soarr}
	return c.respond(om)
}

// responseError returns a non-nil error if a dns response message indicates a failure.
func responseError(m *dns.Msg) error {
	if m.Rcode == dns.RcodeSuccess {
		return nil
	}
	err := fmt.Errorf("dns response error code %q (%d)", dns.RcodeToString[m.Rcode], m.Rcode)
	if opt := m.IsEdns0(); opt != nil {
		for _, o := range opt.Option {
			if ede, ok := o.(*dns.EDNS0_EDE); ok {
				err = fmt.Errorf("%s: %s (%d): %s", err, dns.ExtendedErrorCodeToString[ede.InfoCode], ede.InfoCode, ede.ExtraText)
			}
		}
	}
	return err
}

func deleteRecords(ctx context.Context, log *slog.Logger, provider Provider, zone string, records []libdns.Record) ([]libdns.Record, error) {
	l, err := provider.DeleteRecords(ctx, zone, records)
	log.Debug("deleting records at provider", "err", err, "zone", zone, "records", records, "deleted", l)
	return l, err
}

func appendRecords(ctx context.Context, log *slog.Logger, provider Provider, zone string, records []libdns.Record) ([]libdns.Record, error) {
	l, err := provider.AppendRecords(ctx, zone, records)
	log.Debug("appending records at provider", "err", err, "zone", zone, "records", records, "appended", l)
	return l, err
}

func setRecords(ctx context.Context, log *slog.Logger, provider Provider, zone string, records []libdns.Record) ([]libdns.Record, error) {
	l, err := provider.SetRecords(ctx, zone, records)
	log.Debug("setting records at provider", "err", err, "zone", zone, "records", records, "set", l)
	return l, err
}

package main

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/libdns/libdns"
	"github.com/miekg/dns"

	"github.com/mjl-/bstore"
	"github.com/mjl-/sherpa"
)

var ctxbg = context.Background()

func tcompare(t *testing.T, got, exp any) {
	t.Helper()
	if !reflect.DeepEqual(got, exp) {
		panic(fmt.Sprintf("got:\n%v\nexpected:\n%v\n\n%#v\n%#v", got, exp, got, exp))
	}
}

func tcheck(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %s", msg, err)
	}
}

var fakeProvidersMutex sync.Mutex
var fakeProviders = map[string]*fakeProvider{}

// Testing is done with a fake in-memory provider that implements the Provider
// interface. We reference them through the "ID" field in the serialized JSON
// config, looking them up in the fakeProviders map.
type fakeProvider struct {
	ID          string
	AbsNames    bool // If set, GetRecords returns absolute names.
	FixedSerial bool // If set, serial is fixed, always 1.
	NoSOA       bool // If set, GetRecords does not return a SOA record.

	sync.Mutex
	Records []libdns.Record
}

var _ libdnsProvider = (*fakeProvider)(nil)

func newFakeProvider(p *fakeProvider) {
	fakeProvidersMutex.Lock()
	defer fakeProvidersMutex.Unlock()
	fakeProviders[p.ID] = p
}

// getProvider gets a singleton provider by ID.
func getProvider(pp *fakeProvider) (*fakeProvider, error) {
	fakeProvidersMutex.Lock()
	defer fakeProvidersMutex.Unlock()
	p := fakeProviders[pp.ID]
	if p == nil {
		// errUser is added so we don't cause a server error that prints a stack trace on successful tests.
		return nil, fmt.Errorf("%w: unknown provider instance %q", errUser, pp.ID)
	}
	return p, nil
}

func (pp *fakeProvider) GetRecords(ctx context.Context, zone string) ([]libdns.Record, error) {
	p, err := getProvider(pp)
	if err != nil {
		return nil, err
	}
	p.Lock()
	defer p.Unlock()
	l := make([]libdns.Record, 0, len(p.Records))
	for _, r := range p.Records {
		if p.NoSOA && r.Type == "SOA" {
			continue
		}
		if p.FixedSerial && r.Type == "SOA" {
			// Third field in SOA value is serial, we change it to 1.
			t := strings.Split(r.Value, " ")
			t[2] = "1"
			r.Value = strings.Join(t, " ")
		}
		if p.AbsNames {
			r.Name = libdns.AbsoluteName(r.Name, zone)
		}
		l = append(l, r)
	}
	return l, nil
}

func (p *fakeProvider) remove(r libdns.Record) bool {
	i := p.find(r)
	if i >= 0 {
		copy(p.Records[i:], p.Records[i+1:])
		p.Records = p.Records[:len(p.Records)-1]
	}
	return i >= 0
}

func (p *fakeProvider) find(r libdns.Record) int {
	for i, or := range p.Records {
		if strings.EqualFold(or.Name, r.Name) && or.Type == r.Type && or.TTL == r.TTL && or.Value == r.Value {
			return i
		}
	}
	return -1
}

func (p *fakeProvider) findName(name string) (l []libdns.Record) {
	for _, r := range p.Records {
		if strings.EqualFold(r.Name, name) {
			l = append(l, r)
		}
	}
	return l
}

func (p *fakeProvider) SOA() (libdns.Record, bool) {
	p.Lock()
	defer p.Unlock()
	for _, r := range p.Records {
		if r.Name == "" && r.Type == "SOA" {
			return r, true
		}
	}
	return libdns.Record{}, false
}

func (p *fakeProvider) changeSerial() {
	for i, r := range p.Records {
		if r.Name == "" && r.Type == "SOA" {
			rr, err := dns.NewRR(". 300 IN SOA " + r.Value)
			if err != nil {
				panic("parsing soa rr: " + err.Error())
			}
			soa := rr.(*dns.SOA)
			soa.Serial++
			p.Records[i].Value = strings.TrimPrefix(soa.String(), soa.Header().String())
			return
		}
	}
}

func (pp *fakeProvider) DeleteRecords(ctx context.Context, zone string, l []libdns.Record) ([]libdns.Record, error) {
	p, err := getProvider(pp)
	if err != nil {
		return nil, err
	}
	p.Lock()
	defer p.Unlock()
	var deleted []libdns.Record
	for _, r := range l {
		if p.remove(r) {
			deleted = append(deleted, r)
		}
	}
	if len(deleted) > 0 {
		p.changeSerial()
	}
	return deleted, nil
}

// todo: could implement special behaviour for adding cname when one already exists, or another type of record already exists.
func (pp *fakeProvider) AppendRecords(ctx context.Context, zone string, l []libdns.Record) ([]libdns.Record, error) {
	p, err := getProvider(pp)
	if err != nil {
		return nil, err
	}
	p.Lock()
	defer p.Unlock()
	p.Records = append(p.Records, l...)
	if len(l) > 0 {
		p.changeSerial()
	}
	return l, nil
}

// todo: could implement special behaviour for adding cname when another type of record already exists.
func (pp *fakeProvider) SetRecords(ctx context.Context, zone string, l []libdns.Record) ([]libdns.Record, error) {
	p, err := getProvider(pp)
	if err != nil {
		return nil, err
	}
	p.Lock()
	defer p.Unlock()
	var changed bool
	for _, r := range l {
		if r.Type == "CNAME" {
			nl := p.findName(r.Name)
			if len(nl) > 1 || nl[0].Type != "CNAME" {
				continue
			} else if len(nl) == 1 {
				// Replace current cname.
				p.remove(nl[0])
			}
			p.Records = append(p.Records, r)
			changed = true
		} else if p.find(r) < 0 {
			p.Records = append(p.Records, r)
			changed = true
		}
	}
	if changed {
		p.changeSerial()
	}
	return l, nil
}

// listener for tcp notify messages.
type notify struct {
	t    *testing.T
	c    chan struct{}
	conn net.Listener
	addr string
}

// read a single notify message, with basic checks.
func processNotify(conn net.Conn) error {
	buf := make([]byte, 1024)
	if _, err := io.ReadFull(conn, buf[:2]); err != nil {
		return err
	}
	size := int(buf[0])<<8 | int(buf[1])
	n, err := io.ReadFull(conn, buf[:size])
	if err != nil {
		return err
	}
	var im dns.Msg
	if err := im.Unpack(buf[:n]); err != nil {
		return err
	}
	if im.Opcode != dns.OpcodeNotify {
		return fmt.Errorf("got op %v, expected notify %d", im.Opcode, dns.OpcodeNotify)
	}
	var om dns.Msg
	om.SetReply(&im)
	obuf, err := om.PackBuffer(buf[2:])
	if err != nil {
		return err
	}
	buf[0] = byte(len(obuf) >> 8)
	buf[1] = byte(len(obuf))
	if _, err := conn.Write(buf[:2+len(obuf)]); err != nil {
		return err
	}
	return nil
}

// newNotify starts a new tcp dns notify listener on a random port, sending on
// the returned notify.c when a request came in.
func newNotify(t *testing.T) *notify {
	lconn, err := net.Listen("tcp", "127.0.0.1:0")
	tcheck(t, err, "listen")
	addr := lconn.Addr().String()
	n := &notify{t, make(chan struct{}, 2), lconn, addr}
	go func() {
		for {
			conn, err := lconn.Accept()
			if err != nil {
				// Listening connection is closed when we need to stop.
				return
			}

			if err := processNotify(conn); err != nil {
				slog.Error("processing notify for test", "err", err)
			}
			conn.Close()

			select {
			case n.c <- struct{}{}:
			default:
			}
		}
	}()
	return n
}

func (n *notify) close() {
	n.conn.Close()
}

func (n *notify) drain() {
	for {
		select {
		case <-n.c:
		default:
			return
		}
	}
}

func (n *notify) wait() {
	select {
	case <-time.After(5 * time.Second):
		n.t.Helper()
		panic("no notification within 5s")
	case <-n.c:
	}
}

// tests all get a testEnv, which has a fresh database with two zones and some
// records, a tcp/tls/udp dns servers.
type testEnv struct {
	t       *testing.T
	api     API
	z0, z1  testZone
	tcpaddr string
	tlsaddr string
	udpaddr string
}

// testZone is a zone for testing, with records, valid credentials.
type testZone struct {
	z         Zone
	pc        ProviderConfig
	p         *fakeProvider
	n         *notify
	records   []Record
	soa       Record
	credTSIG  Credential
	credTLS   Credential
	tlsConfig *tls.Config
}

func ldr(id, relname string, ttl int, typ, value string) libdns.Record {
	return libdns.Record{ID: id, Name: relname, TTL: time.Duration(ttl) * time.Second, Type: typ, Value: value}
}

// testDNS runs fn in a new testEnv. fn should make changes and verify they're
// all made correctly.
func testDNS(t *testing.T, fn func(te testEnv, z Zone)) {
	t.Helper()

	testDNSProvider0(t, fn)

	testUseEDNS0 = true
	defer func() {
		testUseEDNS0 = false
	}()
	testDNSProvider1(t, fn)
}

func testDNSProvider0(t *testing.T, fn func(te testEnv, z Zone)) {
	testDNSProvider(t, fn, false, false, false)
}

func testDNSProvider1(t *testing.T, fn func(te testEnv, z Zone)) {
	testDNSProvider(t, fn, true, true, true)
}

func testDNSProvider(t *testing.T, fn func(te testEnv, z Zone), absNames, fixedSerial, noSOA bool) {
	os.Remove("testdata/dnsclay.db")

	dbopts := bstore.Options{
		Timeout: 5 * time.Second,
	}
	var err error
	database, err = bstore.Open(ctxbg, "testdata/dnsclay.db", &dbopts, databaseTypes...)
	xcheckf(err, "open database")
	defer func() {
		err := database.Close()
		logCheck(slog.Default(), err, "close db")
		database = nil
	}()

	shutdownCtx, shutdownCancel = context.WithCancel(ctxbg)
	defer shutdownCancel()

	// Register & add fake providers.
	providers["fake"] = fakeProvider{}

	// With SOA.
	z0p := &fakeProvider{
		ID:          "z0",
		AbsNames:    absNames,
		FixedSerial: fixedSerial,
		NoSOA:       noSOA,
		Records: []libdns.Record{
			ldr("", "", 300, "SOA", "ns0.example. z0.example. 2024010100 3600 300 1209600 300"),
			ldr("", "testhost", 300, "A", "10.0.0.1"),
			ldr("", "testhost", 300, "A", "10.0.0.2"),
		},
	}
	// Without SOA, case-sensitive.
	z1p := &fakeProvider{
		ID: "z1",
		Records: []libdns.Record{
			ldr("", "Testhost", 300, "A", "10.0.0.1"),
		},
	}
	// todo: test with providers that keep IDs

	newFakeProvider(z0p)
	newFakeProvider(z1p)

	z0n := newNotify(t)
	z1n := newNotify(t)
	defer z0n.close()
	defer z1n.close()

	api := API{}

	z0n.drain()
	pc0 := ProviderConfig{Name: "z0.example", ProviderName: "fake", ProviderConfigJSON: `{"ID": "z0"}`}
	pc0 = api.ProviderConfigAdd(ctxbg, pc0)
	z0 := Zone{Name: "z0.example.", ProviderConfigName: pc0.Name}
	z0 = api.ZoneAdd(ctxbg, z0, []ZoneNotify{{0, time.Time{}, z0.Name, z0n.addr, "tcp"}})
	z0n.wait()
	z0, _, _, creds0, sets0 := api.Zone(ctxbg, z0.Name)
	tcompare(t, len(sets0), 2)
	rl0 := api.ZoneRecords(ctxbg, z0.Name)
	tcompare(t, len(rl0), 3)

	z1n.drain()
	pc1 := ProviderConfig{Name: "z1.example", ProviderName: "fake", ProviderConfigJSON: `{"ID": "z1"}`}
	pc1 = api.ProviderConfigAdd(ctxbg, pc1)
	z1 := Zone{Name: "z1.example.", ProviderConfigName: pc1.Name}
	z1 = api.ZoneAdd(ctxbg, z1, []ZoneNotify{{0, time.Time{}, z1.Name, z1n.addr, "tcp"}})
	z1n.wait()
	z1, _, _, creds1, sets1 := api.Zone(ctxbg, z1.Name)
	tcompare(t, len(sets1), 2) // SOA should have been created.
	rl1 := api.ZoneRecords(ctxbg, z1.Name)
	tcompare(t, len(rl1), 2)

	findSOA := func(z Zone, l []Record) Record {
		for _, r := range l {
			if r.Type == Type(dns.TypeSOA) && r.Deleted == nil && r.AbsName == z.Name {
				return r
			}
		}
		t.Fatalf("missing soa record for zone")
		return Record{}
	}
	soa0 := findSOA(z0, rl0)
	soa1 := findSOA(z1, rl1)

	// Generate fake private keys.
	seed0 := make([]byte, 32)
	seed1 := make([]byte, 32)
	seed1[0] = 1
	makeTLS := func(seed []byte, z Zone, name string) (Credential, *tls.Config) {
		cert := xminimalCert(ed25519.NewKeyFromSeed(seed))
		r := sha256.Sum256(cert.Leaf.RawSubjectPublicKeyInfo)
		fp := base64.RawURLEncoding.EncodeToString(r[:])

		tlspubkey := api.ZoneCredentialAdd(ctxbg, z.Name, Credential{0, time.Time{}, name, "tlspubkey", "", fp})
		config := tls.Config{
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{cert},
			NextProtos:         []string{"dot"},
		}

		return tlspubkey, &config
	}

	tlspubkey0, tlsConfig0 := makeTLS(seed0, z0, "tlspubkey0")
	tlspubkey1, tlsConfig1 := makeTLS(seed1, z1, "tlspubkey1")

	te := testEnv{
		t,
		api,
		testZone{z0, pc0, z0p, z0n, rl0, soa0, creds0[0], tlspubkey0, tlsConfig0},
		testZone{z1, pc1, z1p, z1n, rl1, soa1, creds1[0], tlspubkey1, tlsConfig1},
		testtcpconn.Addr().String(),
		testtlsconn.Addr().String(),
		testudpconn.LocalAddr().String(),
	}

	fn(te, z0)

	// Check nothing changed to the other zone.
	nz1, _, _, _, _ := api.Zone(ctxbg, z1.Name)
	tcompare(t, nz1, z1)
	nrl1 := api.ZoneRecords(ctxbg, z1.Name)
	tcompare(t, nrl1, rl1)
}

func (te testEnv) sherpaError(expCode string, fn func()) {
	t := te.t
	t.Helper()
	defer func() {
		x := recover()
		if x == nil {
			panic("did not get expected sherpa error")
		}
		if err, ok := x.(*sherpa.Error); !ok {
			panic(fmt.Sprintf("got panic %#v, expected sherpa error", x))
		} else if err.Code != expCode {
			panic(fmt.Sprintf("got sherpa error code %q, expected %q (%s)", err.Code, expCode, err.Message))
		}
	}()
	fn()
}

func (te testEnv) zoneUnchanged(fn func()) {
	t := te.t
	z0, nr0 := te.z0.z, te.z0.p.Records
	z1, nr1 := te.z1.z, te.z1.p.Records
	fn()
	nz0, nnr0 := te.z0.z, te.z0.p.Records
	nz1, nnr1 := te.z1.z, te.z1.p.Records
	nz0.LastSync = z0.LastSync
	nz1.LastSync = z1.LastSync
	tcompare(t, nz0, z0)
	tcompare(t, nz1, z1)
	tcompare(t, nr0, nnr0)
	tcompare(t, nr1, nnr1)
}

// testChange represents changes done in a call to zoneChanged, for easy comparison.
type testChange struct {
	t *testing.T

	ldrSOAOld libdns.Record
	ldrSOANew libdns.Record
	ldrDel    []libdns.Record
	ldrAdd    []libdns.Record

	rSOAOld []Record
	rSOANew *Record
	rDel    []Record
	rAdd    []Record
}

// check a new SOA with new serial was created.
func (tc testChange) checkNewSOA() {
	t := tc.t
	if tc.rSOANew == nil {
		t.Fatalf("no soa record after change")
	}
	if len(tc.rSOAOld) == 0 {
		t.Fatalf("no old soa record after change")
	}
	for _, old := range tc.rSOAOld {
		if old.SerialFirst == tc.rSOANew.SerialFirst || old.ID == tc.rSOANew.ID {
			t.Fatalf("expected changed soa record, but did not happen, record %v", tc.rSOANew)
		}
	}
}

type typecounts map[string]int

// check for deletions and additions of records in our database.
func (tc testChange) checkRecordDelta(expDel, expAdd typecounts) {
	t := tc.t
	t.Helper()

	tc.checkNewSOA()

	del := typecounts{}
	add := typecounts{}
	for _, r := range tc.rDel {
		del[dns.Type(r.Type).String()]++
	}
	for _, r := range tc.rAdd {
		add[dns.Type(r.Type).String()]++
	}
	tcompare(t, del, expDel)
	tcompare(t, add, expAdd)
}

// get records in a map, except soa record, which is returned separately.
func ldrmap(l []libdns.Record) (map[libdns.Record]libdns.Record, libdns.Record) {
	var zsoa, soa libdns.Record
	m := map[libdns.Record]libdns.Record{}
	for _, r := range l {
		if r.Type == "SOA" {
			if soa != zsoa {
				panic("duplicate soa")
			}
			soa = r
			continue
		}
		if _, ok := m[r]; ok {
			panic(fmt.Sprintf("duplicate record %#v", r))
		}
		m[r] = r
	}
	return m, soa
}

// db records to map, with a separate map for soa records. Includes deleted records.
func rmap(l []Record) (map[int64]Record, map[int64]Record) {
	msoa := map[int64]Record{}
	m := map[int64]Record{}
	for _, r := range l {
		if r.Type == Type(dns.TypeSOA) {
			msoa[r.ID] = r
		} else {
			m[r.ID] = r
		}
	}
	return msoa, m
}

// Compare 2 maps, return deletions and additions.
func splitmap[T comparable, R any](o, n map[T]R) (del, add []R) {
	for k, v := range n {
		if _, ok := o[k]; !ok {
			add = append(add, v)
		}
	}
	for k, v := range o {
		if _, ok := n[k]; !ok {
			del = append(del, v)
		}
	}
	return
}

// Compare two maps with db records, split into deletions and additions. Record IDs
// that are in both, but deleted in the new map, are returned as deleted.
func splitrmap(o, n map[int64]Record) (del, add []Record) {
	for k, v := range n {
		if _, ok := o[k]; !ok {
			if v.Deleted == nil {
				add = append(add, v)
			} else {
				del = append(del, v)
			}
		}
	}
	for k, v := range o {
		if rn, ok := n[k]; ok && v.Deleted == nil && rn.Deleted != nil {
			del = append(del, v)
		}
	}
	return
}

// zoneChanged runs fn and checks that z0 has changed.
func (te testEnv) zoneChanged(fn func()) testChange {
	te.z0.n.drain()

	te.z0.z, _, _, _, _ = te.api.Zone(ctxbg, te.z0.z.Name)
	te.z0.records = te.api.ZoneRecords(ctxbg, te.z0.z.Name)
	te.zoneUnchanged(func() {
		te.api.ZoneRefresh(ctxbg, te.z0.z.Name)
	})

	oldrrl, ldrSOAOld := ldrmap(te.z0.p.Records)
	oldsoarl, oldrl := rmap(te.z0.records)

	fn()

	te.z0.n.wait()

	newrrl, ldrSOANew := ldrmap(te.z0.p.Records)
	ldrDel, ldrAdd := splitmap(oldrrl, newrrl)

	rl := te.api.ZoneRecords(ctxbg, te.z0.z.Name)
	newsoarl, newrl := rmap(rl)
	rSOADel, rSOAAdd := splitrmap(oldsoarl, newsoarl)
	var rSOANew *Record
	if len(rSOAAdd) > 1 {
		panic(fmt.Sprintf("multiple active soa records in zone, %#v", rSOAAdd))
	} else if len(rSOAAdd) == 1 {
		rSOANew = &rSOAAdd[0]
	}
	rDel, rAdd := splitrmap(oldrl, newrl)
	tc := testChange{te.t, ldrSOAOld, ldrSOANew, ldrDel, ldrAdd, rSOADel, rSOANew, rDel, rAdd}

	te.zoneUnchanged(func() {
		te.api.ZoneRefresh(ctxbg, te.z0.z.Name)
	})

	return tc
}

// DNS servers, single instance for all tests.
var testtcpconn net.Listener
var testtlsconn net.Listener
var testudpconn net.PacketConn

func TestMain(m *testing.M) {
	slogInit()
	if s := os.Getenv("DNSCLAY_TEST_LOGLEVEL"); s != "" {
		err := logLevel.UnmarshalText([]byte(s))
		if err != nil {
			panic(fmt.Sprintf("parsing log level %q: %v", s, err))
		}
		if logLevel.Level() < slog.LevelDebug {
			serveTraceDNS = []traceDNS{traceJSON}
		}
	}

	// We would query SOA straight from the authoritative name servers. We don't want
	// to cause outgoing traffic, so we mock the whole function.
	mockGetSOA = func(zone string) (*dns.SOA, error) {
		id := strings.TrimSuffix(zone, ".example.")
		p, err := getProvider(&fakeProvider{ID: id})
		if err != nil {
			return nil, err
		}
		r, ok := p.SOA()
		value := r.Value
		if !ok || p.FixedSerial {
			value = fmt.Sprintf("ns0.%s %s 1 3600 300 1209600 300", zone, zone)
		}
		text := fmt.Sprintf("%s %d IN SOA %s", zone, int(r.TTL/time.Second), value)
		rr, err := dns.NewRR(text)
		if err != nil {
			return nil, fmt.Errorf("new rr: %v", err)
		}
		return rr.(*dns.SOA), nil
	}

	propagationFirstWait = time.Second / 100

	adminpassword = genpassword()

	tlskeypemDefault = "testdata/dnsclay-test.ed25519.privkey-pkcs8.pem"
	os.Remove(tlskeypemDefault)
	tlsPrivKey = xprivatekey(tlskeypemDefault, true)
	tlsCert = xminimalCert(tlsPrivKey)
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: tlsCert.Leaf.Raw})
	err := os.WriteFile("testdata/dnsclay-test.ed25519.cert.pem", certPEM, 0666)
	xcheckf(err, "writing cert")
	tlsCert = xreadcert("testdata/dnsclay-test.ed25519.cert.pem", tlsPrivKey)

	tlsConfig = tlsServerConfig(tlsCert)

	testtcpconn, err = net.Listen("tcp", "127.0.0.1:0")
	xcheckf(err, "listen tcp")

	testtlsconn, err = net.Listen("tcp", "127.0.0.1:0")
	xcheckf(err, "listen tls")

	testudpconn, err = net.ListenPacket("udp", "127.0.0.1:0")
	xcheckf(err, "listen udp")

	testSyncNotify = true
	testSyncUpdate = true

	go func() {
		for {
			conn, err := testtcpconn.Accept()
			if err != nil {
				return
			}
			go serveDNS(conn, listener{false, false, true, true, true})
		}
	}()

	go func() {
		for {
			conn, err := testtlsconn.Accept()
			if err != nil {
				return
			}
			go serveDNS(conn, listener{true, false, true, true, true})
		}
	}()
	go func() {
		for {
			buf := make([]byte, 64*1024)
			n, raddr, err := testudpconn.ReadFrom(buf)
			if err != nil {
				return
			}
			cid := connID.Add(1)
			c := conn{
				cid:           cid,
				udpRemoteAddr: raddr,
				udpconn:       testudpconn,
				log:           slog.With("cid", cid),
				listener:      listener{notify: true},
				buf:           buf,
			}
			c.handleDNS(buf[:n])
		}
	}()

	os.Exit(m.Run())
}

package main

import (
	"crypto"
	"crypto/ed25519"
	cryptorand "crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"math/big"
	"os"
	"time"
)

// Read a pem-encoded pkcs#8 private key from path. If create is set, and the file
// does not exist, generate a new ed25519 key and write it to the path.
func xprivatekey(path string, create bool) crypto.Signer {
	keypem, err := os.ReadFile(path)
	if err != nil && errors.Is(err, fs.ErrNotExist) && create && path == tlskeypemDefault {
		var privKey crypto.Signer
		_, privKey, err = ed25519.GenerateKey(cryptorand.Reader)
		xcheckf(err, "generating ed25519 key")
		buf, err := x509.MarshalPKCS8PrivateKey(privKey)
		xcheckf(err, "marshal private key to pkcs8")
		b := pem.Block{Type: "PRIVATE KEY", Bytes: buf}
		keypem = pem.EncodeToMemory(&b)

		err = os.WriteFile(path, keypem, 0600)
		xcheckf(err, "writing new private key")
		slog.Info("wrote new server private key file", "path", path)
	}
	xcheckf(err, "reading tls key pem file")
	b, _ := pem.Decode(keypem)
	if b == nil {
		err = errors.New("no pem data found")
	} else if b.Type != "PRIVATE KEY" {
		err = fmt.Errorf("found unsupported type %q", b.Type)
	}
	xcheckf(err, "parsing tls key pem file")
	privKey, err := x509.ParsePKCS8PrivateKey(b.Bytes)
	xcheckf(err, "parsing pkcs#8 tls private key")
	return privKey.(crypto.Signer)
}

// xreadcert reads a certificate from path (and potential chain), and combines it
// with privKey into a tls.Certificate for use with tls (with client certificate
// authentication or as server). Checks if certificate is for the public key of the
// private key.
func xreadcert(path string, privKey crypto.Signer) tls.Certificate {
	certpem, err := os.ReadFile(path)
	xcheckf(err, "reading tls cert pem file")

	rest := certpem
	var certs [][]byte
	var leaf *x509.Certificate
	for {
		var b *pem.Block
		b, rest = pem.Decode(rest)
		if b == nil && len(rest) != 0 {
			xcheckf(errors.New("leftover data"), "parsing tls cert pem file")
		} else if b == nil {
			break
		}
		if b.Type != "CERTIFICATE" {
			xcheckf(fmt.Errorf("found supported type %q", b.Type), "looking for pem type \"CERTIFICATE\"")
		}
		cert, err := x509.ParseCertificate(b.Bytes)
		xcheckf(err, "parsing certificate")
		if leaf == nil {
			leaf = cert
		}
		certs = append(certs, b.Bytes)
	}

	type cryptoPublicKey interface {
		Equal(x crypto.PublicKey) bool
	}
	if !privKey.Public().(cryptoPublicKey).Equal(leaf.PublicKey) {
		xcheckf(errors.New("private key does not match public key of certificate"), "making tls certificate")
	}

	return tls.Certificate{
		Certificate: certs,
		PrivateKey:  privKey,
		Leaf:        leaf,
	}
}

// xminimalCert generates a minimal certificate for the private key, with no fields
// other than required certificate serial (so no expires, constraints, names).
func xminimalCert(privKey crypto.Signer) tls.Certificate {
	template := &x509.Certificate{
		// Required field.
		SerialNumber: big.NewInt(time.Now().Unix()),
	}
	certBuf, err := x509.CreateCertificate(cryptorand.Reader, template, template, privKey.Public(), privKey)
	xcheckf(err, "creating minimal certificate")
	cert, err := x509.ParseCertificate(certBuf)
	xcheckf(err, "parsing certificate")
	c := tls.Certificate{
		Certificate: [][]byte{certBuf},
		PrivateKey:  privKey,
		Leaf:        cert,
	}
	return c
}

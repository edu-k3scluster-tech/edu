package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
)

func GeneratePrivateKeyAndCSR(name string) ([]byte, *rsa.PrivateKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, nil, err
	}

	subject := pkix.Name{
		CommonName:         name,
		Country:            []string{},
		Locality:           []string{},
		Organization:       []string{},
		OrganizationalUnit: []string{},
		Province:           []string{},
	}

	asn1, err := asn1.Marshal(subject.ToRDNSequence())
	if err != nil {
		return nil, nil, err
	}

	csrReq := x509.CertificateRequest{
		RawSubject:         asn1,
		EmailAddresses:     []string{},
		SignatureAlgorithm: x509.SHA256WithRSA,
	}
	bytes, err := x509.CreateCertificateRequest(rand.Reader, &csrReq, key)
	if err != nil {
		return nil, nil, err
	}
	return bytes, key, nil
}

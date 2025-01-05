package cluster

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"

	certificates "k8s.io/api/certificates/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// create and approve CSR
func (c *Cluster) CertificateSigningRequest(ctx context.Context, username string, key *rsa.PrivateKey) error {
	subject := pkix.Name{
		CommonName:         username,
		Country:            []string{},
		Locality:           []string{},
		Organization:       []string{},
		OrganizationalUnit: []string{},
		Province:           []string{},
	}
	asn1, err := asn1.Marshal(subject.ToRDNSequence())
	if err != nil {
		return err
	}
	csrReq := x509.CertificateRequest{
		RawSubject:         asn1,
		EmailAddresses:     []string{},
		SignatureAlgorithm: x509.SHA256WithRSA,
	}
	bytes, err := x509.CreateCertificateRequest(rand.Reader, &csrReq, key)
	if err != nil {
		return err
	}
	csr := &certificates.CertificateSigningRequest{
		ObjectMeta: v1.ObjectMeta{
			Name: username,
		},
		Spec: certificates.CertificateSigningRequestSpec{
			Groups: []string{
				"system:authenticated",
			},
			Usages: []certificates.KeyUsage{
				"client auth",
			},
			SignerName: "kubernetes.io/kube-apiserver-client",
			Request:    pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: bytes}),
		},
	}
	_, err = c.clientset.CertificatesV1().CertificateSigningRequests().Create(context.TODO(), csr, v1.CreateOptions{})
	if err != nil {
		var a *k8serrors.StatusError

		if !errors.As(err, &a) {
			return fmt.Errorf("create csr: %v", err)
		} else {
			switch a.ErrStatus.Code {
			case 409:
			default:
				return err
			}
		}
	}
	csr.Status.Conditions = append(csr.Status.Conditions, certificates.CertificateSigningRequestCondition{
		Type:           certificates.CertificateApproved,
		Status:         "True",
		Reason:         "User activation",
		Message:        "This CSR was approved",
		LastUpdateTime: v1.Now(),
	})
	_, err = c.clientset.CertificatesV1().CertificateSigningRequests().UpdateApproval(context.TODO(), username, csr, v1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("approve csr: %v", err)
	}
	return nil
}

// return a certificate from CSR by its name
func (c *Cluster) Certificate(ctx context.Context, name string) ([]byte, bool, error) {
	csr, err := c.clientset.CertificatesV1().CertificateSigningRequests().Get(context.TODO(), name, v1.GetOptions{})
	if err != nil {
		return nil, false, fmt.Errorf("get csr after approval: %v", err)
	}
	if len(csr.Status.Certificate) == 0 {
		return nil, false, nil
	}
	return csr.Status.Certificate, true, nil
}

func (c *Cluster) GenerateUserCertificate(ctx context.Context, username string) ([]byte, []byte, error) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, nil, err
	}

	subject := pkix.Name{
		CommonName:         username,
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

	csr := &certificates.CertificateSigningRequest{
		ObjectMeta: v1.ObjectMeta{
			Name: username,
		},
		Spec: certificates.CertificateSigningRequestSpec{
			Groups: []string{
				"system:authenticated",
			},
			Usages: []certificates.KeyUsage{
				"client auth",
			},
			SignerName: "kubernetes.io/kube-apiserver-client",
			Request:    pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: bytes}),
		},
	}

	_, err = c.clientset.CertificatesV1().CertificateSigningRequests().Create(context.TODO(), csr, v1.CreateOptions{})

	if err != nil {
		var a *k8serrors.StatusError

		if !errors.As(err, &a) {
			return nil, nil, fmt.Errorf("create csr: %v", err)
		} else {
			switch a.ErrStatus.Code {
			case 409:
			default:
				return nil, nil, err
			}
		}
	}

	csr.Status.Conditions = append(csr.Status.Conditions, certificates.CertificateSigningRequestCondition{
		Type:           certificates.CertificateApproved,
		Status:         "True",
		Reason:         "User activation",
		Message:        "This CSR was approved",
		LastUpdateTime: v1.Now(),
	})

	csr, err = c.clientset.CertificatesV1().CertificateSigningRequests().UpdateApproval(context.TODO(), username, csr, v1.UpdateOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("approve csr: %v", err)
	}

	csr, err = c.clientset.CertificatesV1().CertificateSigningRequests().Get(context.TODO(), csr.GetName(), v1.GetOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("get csr after approval: %v", err)
	}

	if len(csr.Status.Certificate) == 0 {
		return nil, nil, fmt.Errorf("certificate is empty")
	}

	return csr.Status.Certificate, pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	), nil
}

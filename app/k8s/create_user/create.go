package createuser

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"edu-portal/app"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"

	certificates "k8s.io/api/certificates/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Creator struct {
	kubeConfigPath string
}

func New(kubeConfigPath string) *Creator {
	return &Creator{
		kubeConfigPath: kubeConfigPath,
	}
}

func (u *Creator) CreateCSR(ctx context.Context, user *app.User) (bool, error) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return false, err
	}

	username := fmt.Sprintf("edu-user-%d", user.Id)

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
		return false, err
	}

	csrReq := x509.CertificateRequest{
		RawSubject:         asn1,
		EmailAddresses:     []string{},
		SignatureAlgorithm: x509.SHA256WithRSA,
	}
	bytes, err := x509.CreateCertificateRequest(rand.Reader, &csrReq, key)
	if err != nil {
		return false, err
	}

	config, err := clientcmd.BuildConfigFromFlags("", u.kubeConfigPath)
	if err != nil {
		return false, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return false, err
	}

	pk64 := base64.StdEncoding.EncodeToString(
		pem.EncodeToMemory(
			&pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(key),
			},
		),
	)

	csr := &certificates.CertificateSigningRequest{
		ObjectMeta: v1.ObjectMeta{
			Name: username,
			Annotations: map[string]string{
				"pk": pk64,
			},
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

	_, err = clientset.CertificatesV1().CertificateSigningRequests().Create(context.TODO(), csr, v1.CreateOptions{})

	var a *k8serrors.StatusError
	if errors.As(err, &a) {
		switch a.ErrStatus.Code {
		case 409:
			return false, nil
		}
	}

	if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func (u *Creator) ApplyRoleBinding(ctx context.Context, user *app.User) (bool, error) {
	return false, nil
}

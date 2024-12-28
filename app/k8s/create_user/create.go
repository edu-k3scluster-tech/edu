package createuser

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"edu-portal/app"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"

	certificates "k8s.io/api/certificates/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type Creator struct{}

func New() *Creator {
	return &Creator{}
}

func (u *Creator) Create(ctx context.Context, user *app.User) (string, error) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return "", err
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
		return "", err
	}

	csrReq := x509.CertificateRequest{
		RawSubject:         asn1,
		EmailAddresses:     []string{},
		SignatureAlgorithm: x509.SHA256WithRSA,
	}
	bytes, err := x509.CreateCertificateRequest(rand.Reader, &csrReq, key)
	if err != nil {
		return "", err
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		return "", err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", err
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

	_, err = clientset.CertificatesV1().CertificateSigningRequests().Create(context.TODO(), csr, v1.CreateOptions{})

	var a *k8serrors.StatusError
	if errors.As(err, &a) {
		switch a.ErrStatus.Code {
		case 409:
			return "", nil
		}
	}

	if err != nil {
		return "", err
	}

	csr.Status.Conditions = append(csr.Status.Conditions, certificates.CertificateSigningRequestCondition{
		Type:           certificates.CertificateApproved,
		Reason:         "User activation",
		Message:        "This CSR was approved",
		LastUpdateTime: v1.Now(),
	})

	csr, err = clientset.CertificatesV1().CertificateSigningRequests().UpdateApproval(context.Background(), "kubernetes.io/kube-apiserver-client", csr, v1.UpdateOptions{})
	if err != nil {
		return "", err
	}

	csr, err = clientset.CertificatesV1().CertificateSigningRequests().Get(context.TODO(), csr.GetName(), v1.GetOptions{})
	if err != nil {
		return "", err
	}

	err = clientset.CertificatesV1().CertificateSigningRequests().Delete(context.TODO(), csr.GetName(), v1.DeleteOptions{})
	if err != nil {
		return "", err
	}

	_, err = clientset.RbacV1().ClusterRoles().Create(ctx, &rbacv1.ClusterRole{
		ObjectMeta: v1.ObjectMeta{
			Name: fmt.Sprintf("cluster-role-%s", username),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "namespaces", "services", "ingresses", "nodes", "endpoints"},
				Verbs:     []string{"get", "watch", "list"},
			},
			{
				APIGroups: []string{"apps"},
				Resources: []string{"deployments", "replicasets"},
				Verbs:     []string{"get", "watch", "list"},
			},
		},
	}, v1.CreateOptions{})
	if err != nil {
		return "", err
	}

	_, err = clientset.RbacV1().ClusterRoleBindings().Create(ctx, &rbacv1.ClusterRoleBinding{
		ObjectMeta: v1.ObjectMeta{
			Name: fmt.Sprintf("cluster-role-binding-%s", username),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:     "user",
				Name:     username,
				APIGroup: "rbac.authorization.k8s.io",
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     fmt.Sprintf("cluster-role-%s", username),
			APIGroup: "rbac.authorization.k8s.io",
		},
	}, v1.CreateOptions{})
	if err != nil {
		return "", err
	}

	kubeconfig := &clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			"k3scluster.tech": {
				Server:                   "https://109.106.138.127:6443",
				CertificateAuthorityData: config.CAData,
			},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			"k3scluster.tech/user": {
				ClientCertificateData: csr.Status.Certificate,
				ClientKeyData: pem.EncodeToMemory(
					&pem.Block{
						Type:  "RSA PRIVATE KEY",
						Bytes: x509.MarshalPKCS1PrivateKey(key),
					},
				),
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			"k3scluster.tech/user": {
				Cluster:  "k3scluster.tech",
				AuthInfo: "k3scluster.tech/user",
			},
		},
		CurrentContext: "k3scluster.tech/user",
	}

	data, err := clientcmd.Write(*kubeconfig)
	if err != nil {
		return "", err
	} else {
		return string(data), nil
	}
}

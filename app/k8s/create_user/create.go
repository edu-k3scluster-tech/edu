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

const clusterHost = "https://109.106.138.127:6443"

type Creator struct{}

func New() *Creator {
	return &Creator{}
}

func (c *Creator) Create(ctx context.Context, user *app.User) (string, error) {
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
		return "", fmt.Errorf("get incluster config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", fmt.Errorf("create config: %v", err)
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

	if err != nil {
		var a *k8serrors.StatusError

		if !errors.As(err, &a) {
			return "", fmt.Errorf("create csr: %v", err)
		} else {
			switch a.ErrStatus.Code {
			case 409:
			default:
				return "", err
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

	csr, err = clientset.CertificatesV1().CertificateSigningRequests().UpdateApproval(context.Background(), username, csr, v1.UpdateOptions{})
	if err != nil {
		return "", fmt.Errorf("approve csr: %v", err)
	}

	csr, err = clientset.CertificatesV1().CertificateSigningRequests().Get(context.TODO(), csr.GetName(), v1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("get csr after approval: %v", err)
	}

	err = clientset.CertificatesV1().CertificateSigningRequests().Delete(context.TODO(), csr.GetName(), v1.DeleteOptions{})
	if err != nil {
		return "", fmt.Errorf("delete csr: %v", err)
	}

	_, err = c.CreateClusterRoleBinding(ctx, clientset, username)
	if err != nil {
		return "", fmt.Errorf("create cluster role binding: %v", err)
	}

	kubeconfig := &clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			"k3scluster.tech": {
				Server:                   clusterHost,
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
		return "", fmt.Errorf("write kube config: %v", err)
	} else {
		return string(data), nil
	}
}

func (c *Creator) CreateClusterRole(ctx context.Context, client *kubernetes.Clientset, username string) (string, error) {
	name := fmt.Sprintf("cluster-role-%s", username)

	_, err := client.RbacV1().ClusterRoles().Create(ctx, &rbacv1.ClusterRole{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
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
	} else {
		return name, nil
	}
}

func (c *Creator) CreateClusterRoleBinding(ctx context.Context, client *kubernetes.Clientset, username string) (string, error) {
	name := fmt.Sprintf("cluster-role-binding-%s", username)

	_, err := client.RbacV1().ClusterRoleBindings().Create(ctx, &rbacv1.ClusterRoleBinding{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
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
			Name:     "ro-cluster-role",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}, v1.CreateOptions{})

	if err != nil {
		return "", err
	} else {
		return name, nil
	}
}

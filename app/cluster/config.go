package cluster

import (
	"context"

	clientcmd "k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func (c *Cluster) Config(ctx context.Context, certificate, privateKey []byte) (string, error) {
	clusterName := "k3scluster.tech"
	clusterServer := "https://109.106.138.127:6443"
	userName := "user"
	contextName := "k3scluster.tech/user"
	namespace := "default"

	config := clientcmdapi.NewConfig()

	config.Clusters[clusterName] = &clientcmdapi.Cluster{
		Server:                   clusterServer,
		CertificateAuthorityData: []byte(c.ca), // Укажите сертификат CA (если нужен)
	}

	config.AuthInfos[userName] = &clientcmdapi.AuthInfo{
		ClientCertificateData: []byte(certificate),
		ClientKeyData:         []byte(privateKey),
	}

	config.Contexts[contextName] = &clientcmdapi.Context{
		Cluster:   clusterName,
		AuthInfo:  userName,
		Namespace: namespace,
	}

	config.CurrentContext = contextName

	data, err := clientcmd.Write(*config)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

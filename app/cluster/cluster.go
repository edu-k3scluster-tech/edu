package cluster

import (
	"k8s.io/client-go/kubernetes"
)

type Cluster struct {
	clientset *kubernetes.Clientset
	ca        []byte
}

func New(clientset *kubernetes.Clientset, ca []byte) *Cluster {
	return &Cluster{
		clientset: clientset,
		ca:        ca,
	}
}

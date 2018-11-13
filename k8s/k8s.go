package k8s

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// NewClientSet will create a new InCluster config for k8s cluster
func NewClientSet() kubernetes.Interface {
	// creating in-cluster config.
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Error(err)
		return nil
	}

	// create clientset of kubernetes.
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error(err)
		return nil
	}

	return clientSet
}

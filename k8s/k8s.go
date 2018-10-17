package k8s

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	// ClientSet contains kubernetes client.
	ClientSet *kubernetes.Clientset
)

func init() {
	// creating in-cluster config.
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Error(err)
	}

	// create clientset of kubernetes.
	ClientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Error(err)
	}
}

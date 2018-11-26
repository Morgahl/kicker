package client

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/curlymon/kicker/pkg/conf"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // the following line to loads the gcp plugin (only required to authenticate against GKE clusters).
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// New instantiates a new kubernetes.ClientSet from the config file or the environment.
func New(c conf.Conf) (*kubernetes.Clientset, error) {
	if c.KubeConf != "" {
		config, err := clientcmd.BuildConfigFromFlags("", c.KubeConf)
		if err != nil {
			return nil, fmt.Errorf("error creating client from KubeConf: %s", err)
		}

		return kubernetes.NewForConfig(config)
	}

	if config, err := rest.InClusterConfig(); err == nil {
		return kubernetes.NewForConfig(config)
	} else if err != rest.ErrNotInCluster {
		return nil, fmt.Errorf("error creating client from InClusterConfig: %s", err)
	}

	if home := homeDir(); home != "" {
		path := filepath.Join(home, ".kube", "config")
		config, err := clientcmd.BuildConfigFromFlags("", path)
		if err != nil {
			return nil, fmt.Errorf("error creating client from '%s': %s", path, err)
		}

		return kubernetes.NewForConfig(config)
	}

	return nil, fmt.Errorf("No KubeConf provided and unable to resolve config from InClusterConfig or invoking user's home directory")
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

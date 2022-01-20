package kubernetes

import (
	"os"
	"path/filepath"

	"gopkg.in/alecthomas/kingpin.v2"
	client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"
)

type K8SInformer struct {
	Clientset *client.Clientset
	config    *rest.Config
}

func NewClient(apiserver *string, kubecfg *string) *K8SInformer {
	config, err := buildConfigFromFlags(*apiserver, *kubecfg)
	kingpin.FatalIfError(err, "cannot create Kubernetes client configuration")
	clientset, err := client.NewForConfig(config)
	kingpin.FatalIfError(err, "cannot create Kubernetes client")

	return &K8SInformer{config: config, Clientset: clientset}
}

func buildConfigFromFlags(apiserver, kubecfg string) (*rest.Config, error) {
	if home := homedir.HomeDir(); kubecfg == "" && home != "" {
		filePath := filepath.Join(home, ".kube", "config")
		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			kubecfg = filePath
		}
	}

	if kubecfg != "" || apiserver != "" {
		return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubecfg},
			&clientcmd.ConfigOverrides{ClusterInfo: api.Cluster{Server: apiserver}}).ClientConfig()
	}
	return rest.InClusterConfig()
}

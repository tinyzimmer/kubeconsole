package k8sutils

import (
	"context"
	"io"
	"os"
	"os/user"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

type KubernetesFactory interface {
	BuildConfigFromFlags(string, string) (*rest.Config, error)
	NewForConfig(*rest.Config) (*kubernetes.Clientset, error)

	CreateClientSet() error

	ApiHost() string

	ListNamespaces() ([]string, error)
	ListPods(string) ([]string, error)

	GetPod(string, string) (*corev1.Pod, error)
	GetLogStream(string, string, string, context.Context) (io.ReadCloser, error)
	GetExecutor(string, string, string) (remotecommand.Executor, error)
}

type CoreV1 interface {
	v1.PodsGetter
	v1.NamespacesGetter
}

type kubeFactory struct {
	KubernetesFactory
	CoreV1

	newClientFunc   func(*rest.Config) (*kubernetes.Clientset, error)
	buildConfigFunc func(string, string) (*rest.Config, error)
	podsFunc        func(CoreV1, string) v1.PodInterface
	namespacesFunc  func(CoreV1) v1.NamespaceInterface

	conf      *rest.Config
	clientset *kubernetes.Clientset
}

func New() KubernetesFactory {
	return &kubeFactory{
		buildConfigFunc: clientcmd.BuildConfigFromFlags,
		newClientFunc:   kubernetes.NewForConfig,
		podsFunc:        CoreV1.Pods,
		namespacesFunc:  CoreV1.Namespaces,
	}
}

func (k *kubeFactory) BuildConfigFromFlags(url string, kpath string) (*rest.Config, error) {
	return k.buildConfigFunc(url, kpath)
}

func (k *kubeFactory) NewForConfig(conf *rest.Config) (*kubernetes.Clientset, error) {
	return k.newClientFunc(conf)
}

func (k *kubeFactory) Pods(c CoreV1, s string) (iface v1.PodInterface) {
	return k.podsFunc(c, s)
}

func (k *kubeFactory) Namespaces(c CoreV1) (iface v1.NamespaceInterface) {
	return k.namespacesFunc(c)
}

func (k *kubeFactory) ApiHost() string {
	return k.conf.Host
}

func (k *kubeFactory) CreateClientSet() (err error) {
	k.conf, err = k.BuildConfigFromFlags("", getKubeConfig())
	if err != nil {
		return
	}
	k.clientset, err = k.NewForConfig(k.conf)
	return
}

func getKubeConfig() (config string) {
	usr, err := user.Current()
	if err == nil {
		config = filepath.Join(usr.HomeDir, ".kube", "config")
	}
	if env := os.Getenv("KUBECONFIG"); env != "" {
		config = env
	}
	return
}

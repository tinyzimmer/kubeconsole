package k8sutils

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v2"
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

	AvailableContexts() ([]string, error)
	SwitchContext(string) error
	CreateClientSet() error

	APIHost() string
	APIVersion() (string, error)

	ListNamespaces() ([]string, error)
	ListPods(string) ([]string, error)

	GetPod(string, string) (*corev1.Pod, error)
	GetLogStream(string, string, string, context.Context) (io.ReadCloser, error)
	GetExecutor(string, string, string) (remotecommand.Executor, error)
}

type kubeContexts struct {
	Contexts []kubeContext `yaml:"contexts"`
}

type kubeContext struct {
	Ctx  map[string]string `yaml:"context"`
	Name string            `yaml:"name"`
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

	incluster bool
	conf      *rest.Config
	clientset *kubernetes.Clientset
}

func New(incluster bool) KubernetesFactory {
	return &kubeFactory{
		buildConfigFunc: clientcmd.BuildConfigFromFlags,
		newClientFunc:   kubernetes.NewForConfig,
		podsFunc:        CoreV1.Pods,
		namespacesFunc:  CoreV1.Namespaces,
		incluster:       incluster,
	}
}

func (k *kubeFactory) BuildConfigFromFlags(url, kpath string) (*rest.Config, error) {
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

func (k *kubeFactory) APIHost() string {
	return k.conf.Host
}

func (k *kubeFactory) APIVersion() (version string, err error) {
	vers, err := k.clientset.Discovery().ServerVersion()
	if err != nil {
		return
	}
	version = vers.String()
	return
}

func (k *kubeFactory) CreateClientSet() (err error) {
	if k.incluster {
		k.conf, err = rest.InClusterConfig()
	} else {
		k.conf, err = k.BuildConfigFromFlags("", getKubeConfig())
	}
	if err != nil {
		return
	}
	k.clientset, err = k.NewForConfig(k.conf)
	return
}

func (k *kubeFactory) AvailableContexts() (contexts []string, err error) {
	if k.incluster {
		err = errors.New("Context switching not available with in-cluster config")
		return
	}
	var ctxs kubeContexts
	conf := getKubeConfig()
	if _, err := os.Stat(conf); err == nil {
		file, err := os.Open(conf)
		if err != nil {
			return nil, err
		}
		body, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(body, &ctxs)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}
	for _, c := range ctxs.Contexts {
		contexts = append(contexts, c.Name)
	}
	return
}

func (k *kubeFactory) SwitchContext(ctx string) (err error) {
	k.conf, err = buildConfigWithContext(ctx, getKubeConfig())
	if err != nil {
		return
	}
	k.clientset, err = k.NewForConfig(k.conf)
	return
}

func buildConfigWithContext(context, kubeconfig string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		}).ClientConfig()
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

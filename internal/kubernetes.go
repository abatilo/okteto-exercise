package internal

import (
	"context"
	"io/ioutil"

	"github.com/rs/zerolog"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type (
	ObjectMeta        = metav1.ObjectMeta
	Time              = metav1.Time
	Pod               = v1.Pod
	PodStatus         = v1.PodStatus
	ContainerStatuses = v1.ContainerStatus
	PodList           = v1.PodList
	Result            = rest.Result
)

type ControlPlaneClient interface {
	ListPods(ctx context.Context) (*v1.PodList, error)
	Healthz(ctx context.Context) Result
}

// MockKubernetesClient is a mock implementation of KubernetesClient. It's used
// for testing. Normally I'd just use like https://github.com/golang/mock
type MockKubernetesClient struct {
	PodList *PodList
	Error   error
}

func (m *MockKubernetesClient) ListPods(ctx context.Context) (*PodList, error) {
	return m.PodList, m.Error
}

func (m *MockKubernetesClient) Healthz(ctx context.Context) Result {
	return Result{}
}

// Real implementation of a Kubernetes client
type KubernetesClient struct {
	log       zerolog.Logger
	clientset *kubernetes.Clientset
}

func NewKubernetesClient(log zerolog.Logger) *KubernetesClient {
	cfg, _ := rest.InClusterConfig()
	clientset, _ := kubernetes.NewForConfig(cfg)

	return &KubernetesClient{
		log:       log,
		clientset: clientset,
	}
}

func (k *KubernetesClient) ListPods(ctx context.Context) (*PodList, error) {
	namespace, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		k.log.Fatal().Err(err).Msg("Failed to read namespace")
	}
	return k.clientset.CoreV1().Pods(string(namespace)).List(ctx, metav1.ListOptions{})
}

func (k *KubernetesClient) Healthz(ctx context.Context) Result {
	return k.clientset.Discovery().RESTClient().Get().AbsPath("/healthz").Do(ctx)
}

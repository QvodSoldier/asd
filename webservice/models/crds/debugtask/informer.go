package debugtask

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type DebugInterface interface {
	List(opts metav1.ListOptions) (*DebugTaskList, error)
	Get(name string, options metav1.GetOptions) (*DebugTask, error)
	Create(*DebugTask) (*DebugTask, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	Update(*DebugTask) (*DebugTask, error)
	Delete(name string, options *metav1.DeleteOptions) error
}

type DebugClient struct {
	restClient rest.Interface
	ns         string
}

type DebugTaskInterface interface {
	Debug(namespace string) DebugInterface
}

type DebugTaskClient struct {
	restClient rest.Interface
}

func (c *DebugTaskClient) Debug(namespace string) DebugInterface {
	return &DebugClient{
		restClient: c.restClient,
		ns:         namespace,
	}
}

func (c *DebugClient) List(opts metav1.ListOptions) (*DebugTaskList, error) {
	result := DebugTaskList{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("debugtasks").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(&result)

	return &result, err
}

func (c *DebugClient) Get(name string, opts metav1.GetOptions) (*DebugTask, error) {
	result := DebugTask{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("debugtasks").
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(&result)

	return &result, err
}

func (c *DebugClient) Create(debug *DebugTask) (*DebugTask, error) {
	result := DebugTask{}
	err := c.restClient.
		Post().
		Namespace(c.ns).
		Resource("debugtasks").
		Body(debug).
		Do().
		Into(&result)

	return &result, err
}

func (c *DebugClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.restClient.
		Get().
		Namespace(c.ns).
		Resource("debugtasks").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

func (c *DebugClient) Update(debug *DebugTask) (result *DebugTask, err error) {
	result = &DebugTask{}
	err = c.restClient.Put().
		Namespace(c.ns).
		Resource("debugtasks").
		Name(debug.Name).
		Body(debug).
		Do().
		Into(result)
	return
}

func (c *DebugClient) Delete(name string, options *metav1.DeleteOptions) error {
	return c.restClient.Delete().
		Namespace(c.ns).
		Resource("debugtasks").
		Name(name).
		Body(options).
		Do().
		Error()
}

func WatchResources(clientSet DebugTaskInterface) cache.Store {
	appStore, appController := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(lo metav1.ListOptions) (result runtime.Object, err error) {
				return clientSet.Debug("").List(lo)
			},
			WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
				return clientSet.Debug("").Watch(lo)
			},
		},
		&DebugTask{},
		5*time.Minute,
		cache.ResourceEventHandlerFuncs{},
	)
	stop := make(chan struct{})
	go appController.Run(stop)
	return appStore
}

func NewDebugClientForConfig(c *rest.Config) (*DebugTaskClient, error) {
	config := *c
	config.ContentConfig.GroupVersion = &schema.GroupVersion{Group: GroupName, Version: GroupVersion}
	config.APIPath = "/apis"
	// config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}
	config.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{CodecFactory: scheme.Codecs}
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &DebugTaskClient{restClient: client}, nil
}

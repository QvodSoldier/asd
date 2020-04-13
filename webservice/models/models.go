package models

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"ggstudy/asd/webservice/models/crds/debugtask"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	lcorev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

// Cache is the k8s's core resource
type Cache struct {
	Client          *kubernetes.Clientset
	NameSpaceLister lcorev1.NamespaceLister
	PodLister       lcorev1.PodLister
}

type Store struct {
	ClientSet      *debugtask.DebugTaskClient
	DebugTaskStore cache.Store
}

var Cf *rest.Config

func init() {
	var err error
	Cf, err = generateRestConfig()
	if err != nil {
		log.Println(err)
	}
}

func NewStore() *Store {
	cs, err := debugtask.NewDebugClientForConfig(Cf)
	if err != nil {
		log.Println(err)
	}
	s := &Store{
		ClientSet: cs,
	}
	s.resourceLoad()
	return s
}

func (s *Store) resourceLoad() {
	debugtask.AddToScheme(scheme.Scheme)
	s.DebugTaskStore = debugtask.WatchResources(s.ClientSet)
}

// NewCache is used to generate a cache for a kubernetes's resource
func NewCache() *Cache {
	cs, err := kubernetes.NewForConfig(Cf)
	if err != nil {
		log.Println(err)
	}

	c := &Cache{
		Client: cs,
	}
	c.resourceLoad()
	return c
}

func (c *Cache) resourceLoad() {
	ifs := make([]cache.SharedIndexInformer, 2)
	stopper := make(chan struct{})

	factory := informers.NewSharedInformerFactory(c.Client, 5*time.Minute)
	namespaceInformer := factory.Core().V1().Namespaces()
	podInformer := factory.Core().V1().Pods()
	c.NameSpaceLister = namespaceInformer.Lister()
	c.PodLister = podInformer.Lister()
	ifs[0] = namespaceInformer.Informer()
	ifs[1] = podInformer.Informer()

	defer runtime.HandleCrash()
	go factory.Start(stopper)
	for _, v := range ifs {
		if !cache.WaitForCacheSync(stopper, v.HasSynced) {
			runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
			return
		}
	}
}

func generateRestConfig() (*rest.Config, error) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfigs", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		log.Println(*kubeconfig)
	} else {
		kubeconfig = flag.String("kubeconfigs", "", "(optional) absolute path to the kubeconfig file")
	}
	flag.Parse()

	// 在 kubeconfig 中使用当前上下文环境，config 获取支持 url 和 path 方式
	// var tmp = make([]*rest.Config, 2)
	// for i := 0; i < 2; i++ {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("in cluster config failed: %v", err)
		config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			return nil, err
		}
	}
	// tmp[i] = config
	// }

	return config, nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

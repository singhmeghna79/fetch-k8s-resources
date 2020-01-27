package dynamicwatcher

import (
	"time"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

// Watcher ...
type Watcher struct {
	WatcherInterface
	DynamicClient             dynamic.Interface
	GroupVersionResource      schema.GroupVersionResource
	Informer                  informers.GenericInformer
	ResourceEventHandlerFuncs cache.ResourceEventHandlerFuncs
	StopChannel               chan bool
}

// New ...
func New(dc dynamic.Interface, gvr schema.GroupVersionResource,
	i informers.GenericInformer, rehf cache.ResourceEventHandlerFuncs) *Watcher {
	return &Watcher{
		DynamicClient:             dc,
		GroupVersionResource:      gvr,
		StopChannel:               make(chan bool),
		Informer:                  i,
		ResourceEventHandlerFuncs: rehf,
	}
}

// Run maintains full lifecycle of a dynamic watcher. If any error occurs it
// calls schedule run and tries to re-run . For now it waits constant time and
// do a re run. We can put more intelligence with wait time later. Main intention
// to have it - you call run and it will do self heal for any error.
func (w *Watcher) Run() {
	err := w.Verify()
	if err != nil {
		w.ScheduleRun(120 * time.Second)
	}
	w.Watch()
}

// ScheduleRun calls run after given duration.
func (w *Watcher) ScheduleRun(d time.Duration) {
	time.Sleep(d)
	w.Run()
}

// Verify checks if we can start watcher for a given details.
func (w *Watcher) Verify() error {
	_, err := w.DynamicClient.
		Resource(w.GroupVersionResource).
		Namespace("").
		List(metav1.ListOptions{})
	return err
}

// Watch contains start and stop watcher functionality.
func (w *Watcher) Watch() {
	ch := make(chan struct{})
	go func(stopCh <-chan struct{}) {
		w.Informer.Informer().AddEventHandler(w.ResourceEventHandlerFuncs)
		w.Informer.Informer().Run(stopCh)
	}(ch)
	<-w.StopChannel
	close(ch)
	logrus.Info("stoping watcher for ", w.GroupVersionResource)
}

// Stop is the only way to stop a watcher. If you called stop it will not be
// re launched automatically.
func (w *Watcher) Stop() {
	w.StopChannel <- true
}

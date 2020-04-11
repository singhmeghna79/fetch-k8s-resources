package main

import (
	"os"
	"os/signal"

	"encoding/json"

	"github.com/sirupsen/logrus"

	"k8s.io/client-go/dynamic/dynamicinformer"

	"github.com/shovanmaity/fetch-k8s-resource/pkg/dynamicwatcher"
	"github.com/shovanmaity/fetch-k8s-resource/pkg/memdb"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

var tables = make(map[string]*memdb.Table)

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	dci := dynamicinformer.NewFilteredDynamicSharedInformerFactory(client, 0, "", nil)

	gvrs := []schema.GroupVersionResource{
		{Group: "openebs.io", Version: "v1alpha1", Resource: "cstorpoolclusters"},
		{Group: "openebs.io", Version: "v1alpha1", Resource: "blockdeviceclaims"},
		{Group: "openebs.io", Version: "v1alpha1", Resource: "blockdevices"},
		{Group: "openebs.io", Version: "v1alpha1", Resource: "cstorpoolinstances"},
		{Group: "openebs.io", Version: "v1alpha1", Resource: "disks"},
		{Group: "openebs.io", Version: "v1alpha1", Resource: "castemplates"},
		{Group: "openebs.io", Version: "v1alpha1", Resource: "csivolumes"},
		{Group: "openebs.io", Version: "v1alpha1", Resource: "cstorbackups"},
		{Group: "openebs.io", Version: "v1alpha1", Resource: "cstorcompletedbackups"},
		{Group: "openebs.io", Version: "v1alpha1", Resource: "cstorpools"},
		{Group: "openebs.io", Version: "v1alpha1", Resource: "cstorrestores"},
		{Group: "openebs.io", Version: "v1alpha1", Resource: "cstorvolumeclaims"},
		{Group: "openebs.io", Version: "v1alpha1", Resource: "cstorvolumereplicas"},
		{Group: "openebs.io", Version: "v1alpha1", Resource: "cstorvolumes"},
		{Group: "openebs.io", Version: "v1alpha1", Resource: "runtasks"},
		{Group: "openebs.io", Version: "v1alpha1", Resource: "storagepoolclaims"},
		{Group: "openebs.io", Version: "v1alpha1", Resource: "storagepools"},
		{Group: "openebs.io", Version: "v1alpha1", Resource: "upgradetasks"},
		{Group: "", Version: "v1", Resource: "persistentvolumeclaims"},
		{Group: "", Version: "v1", Resource: "persistentvolumes"},
		{Group: "", Version: "v1", Resource: "pods"},
		{Group: "", Version: "v1", Resource: "replicationcontrollers"},
		{Group: "apps", Version: "v1", Resource: "daemonsets"},
		{Group: "apps", Version: "v1", Resource: "deployments"},
		{Group: "apps", Version: "v1", Resource: "replicasets"},
	}

	handlers := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			us := unstructured.Unstructured{}
			inInterface := make(map[string]interface{})
			inrec, err := json.Marshal(obj)
			if err != nil {
				logrus.Error(err)
				return
			}
			err = json.Unmarshal(inrec, &inInterface)
			if err != nil {
				logrus.Error(err)
				return
			}
			us.Object = inInterface

			t, ok := tables[us.GetKind()]
			if !ok {
				schema := memdb.NewSchemaForTable(us.GetKind())
				t, err = schema.Apply()
				if err != nil {
					logrus.Error(err)
					return
				}
				logrus.Infof("Creating in memory database for %s", us.GetKind())
				tables[us.GetKind()] = t
			}
			r := memdb.K8sResource{
				us.GetName(),
				us.GetNamespace(),
				us.GetKind(),
				us.GetAPIVersion(),
				us.GetLabels(),
				us.GetAnnotations(),
				us.GetResourceVersion(),
				string(us.GetUID()),
				us.UnstructuredContent(),
			}
			err = t.Save(r)
			if err != nil {

			}
			//logrus.Infof("received add event [%s/%s] in [%s] namespace", us.GetKind(), us.GetName(), us.GetNamespace())

		},
		UpdateFunc: func(oldObj, obj interface{}) {
			usNew := unstructured.Unstructured{}
			inInterfaceNew := make(map[string]interface{})
			inrecNew, err := json.Marshal(obj)
			if err != nil {
				logrus.Error(err)
				return
			}
			err = json.Unmarshal(inrecNew, &inInterfaceNew)
			if err != nil {
				logrus.Error(err)
				return
			}
			usNew.Object = inInterfaceNew

			usOld := unstructured.Unstructured{}
			inInterfaceOld := make(map[string]interface{})
			inrecOld, err := json.Marshal(obj)
			if err != nil {
				logrus.Error(err)
				return
			}
			err = json.Unmarshal(inrecOld, &inInterfaceOld)
			if err != nil {
				logrus.Error(err)
				return
			}
			usOld.Object = inInterfaceOld
			if usOld.GetResourceVersion() != usNew.GetResourceVersion() {
				t, ok := tables[usNew.GetKind()]
				if !ok {
					schema := memdb.NewSchemaForTable(usNew.GetKind())
					t, err = schema.Apply()
					if err != nil {
						logrus.Error(err)
						return
					}
					logrus.Infof("Creating in memory database for %s", usNew.GetKind())
					tables[usNew.GetKind()] = t
				}
				r := memdb.K8sResource{
					usNew.GetName(),
					usNew.GetNamespace(),
					usNew.GetKind(),
					usNew.GetAPIVersion(),
					usNew.GetLabels(),
					usNew.GetAnnotations(),
					usNew.GetResourceVersion(),
					string(usNew.GetUID()),
					usNew.UnstructuredContent(),
				}
				err = t.Save(r)
				if err != nil {
					logrus.Error(err)
				}
				logrus.Infof("received update event [%s/%s] in [%s] namespace", usNew.GetKind(), usNew.GetName(), usNew.GetNamespace())
			}
		},
		DeleteFunc: func(obj interface{}) {
			us := unstructured.Unstructured{}
			inInterface := make(map[string]interface{})
			inrec, err := json.Marshal(obj)
			if err != nil {
				logrus.Error(err)
				return
			}
			err = json.Unmarshal(inrec, &inInterface)
			if err != nil {
				logrus.Error(err)
				return
			}
			us.Object = inInterface
			logrus.Infof("received delete event [%s/%s] in [%s] namespace", us.GetKind(), us.GetName(), us.GetNamespace())
		},
	}

	ws := make([]*dynamicwatcher.Watcher, 0)
	for _, gvr := range gvrs {
		w := dynamicwatcher.New(client, gvr, dci.ForResource(gvr), handlers)
		ws = append(ws, w)
		go w.Run()
	}

	sigCh := make(chan os.Signal, 0)
	signal.Notify(sigCh, os.Kill, os.Interrupt)

	<-sigCh
	for _, w := range ws {
		w.Stop()
	}
	return
}

/*
Copyright 2020 mahuang.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	debugv1alpha1 "ggstudy/asd/api/v1alpha1"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apiErr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	myFinalizerName = "debug.finalizers.name"
)

// DebugTaskReconciler ctrls a DebugTask object
type DebugTaskReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=debug.mahuang.cn,resources=debugtasks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=debug.mahuang.cn,resources=debugtasks/status,verbs=get;update;patch

func (r *DebugTaskReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	// ctx := context.Background()
	log := r.Log.WithValues("debugtask", req.NamespacedName)
	log.Info("debug task reconcile", "request", req)

	instance := &debugv1alpha1.DebugTask{}
	err := r.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if apiErr.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			// TODO: add cleanup finalizers
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}
	// examine DeletionTimestamp to determine if object is under deletion
	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !containsString(instance.ObjectMeta.Finalizers, myFinalizerName) {
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, myFinalizerName)
			if err := r.Update(context.Background(), instance); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if containsString(instance.ObjectMeta.Finalizers, myFinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.Delete(context.Background(), instance); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			instance.ObjectMeta.Finalizers = removeString(instance.ObjectMeta.Finalizers, myFinalizerName)
			if err := r.Update(context.Background(), instance); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	if err := r.updateStatus(instance); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *DebugTaskReconciler) updateStatus(instance *debugv1alpha1.DebugTask) error {
	log := r.Log.WithValues("debugtask", "status-flow")

	name := "debug-pod-" + instance.Name
	namespace := instance.Spec.TargetObjectInfo.TargetPodNamespace
	image := instance.Spec.DebugObjectInfo.DebugPodImage
	switch instance.Status.Phase {
	case "":
		node, err := r.getNodeName(types.NamespacedName{
			Name:      instance.Spec.TargetObjectInfo.TargetPodName,
			Namespace: namespace})
		if err != nil {
			return err
		}

		pod := getPod(name, namespace, image, node)
		err = r.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, pod)
		if err != nil && apiErr.IsNotFound(err) {
			log.Info("Creating Pod", "namespace", namespace, "name", name)
			err = controllerutil.SetOwnerReference(instance, pod, r.Scheme)
			if err != nil {
				return err
			}
			err = r.Create(context.TODO(), pod)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		instance.Spec.DebugObjectInfo.DebugPodName = pod.Name
		instance.Status.Phase = debugv1alpha1.DebugPending
		if err := r.Update(context.TODO(), instance); err != nil {
			return err
		}
		return nil
	case debugv1alpha1.DebugSucceeded:
		pod := &corev1.Pod{}
		err := r.Get(context.TODO(), types.NamespacedName{
			Name:      instance.Spec.DebugObjectInfo.DebugPodName,
			Namespace: namespace}, pod)
		if err != nil {
			return err
		}
		err = r.Delete(context.TODO(), pod)
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}

// get node that debug target on
func (r *DebugTaskReconciler) getNodeName(nn types.NamespacedName) (string, error) {
	pod := &corev1.Pod{}
	err := r.Get(context.TODO(), nn, pod)
	if err != nil {
		return "", err
	}

	return pod.Spec.NodeName, nil
}

func (r *DebugTaskReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&debugv1alpha1.DebugTask{}).
		Complete(r)
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

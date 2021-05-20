/*


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
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metallbv1beta1 "community-operators/community-operators/metallb-operator/apis/metallb/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	uns "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
)

const (
	addrpools_configMapFile string = "./template/addrpool_config.yaml"
)

// AddressPoolReconciler reconciles a AddressPool object
type AddressPoolReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=metallb.metallb.io,resources=addresspools,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=metallb.metallb.io,resources=addresspools/status,verbs=get;update;patch

func (r *AddressPoolReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("addresspool", req.NamespacedName)
	log.Info("Reconciling AddressPools resource")

	instance := &metallbv1beta1.AddressPool{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Create config map based on CR
	if err := r.applyConfigMap(ctx, instance); err != nil {
		errors.Wrap(err, "Failed to create config map")
		return ctrl.Result{}, err
	}

	log.Info("Reconcile complete")
	return ctrl.Result{}, nil
}

func (r *AddressPoolReconciler) applyConfigMap(ctx context.Context, instance *metallbv1beta1.AddressPool) error {
	data := make(map[string]interface{})
	data["AddressPool"] = instance.Spec.AddressPool

	obj, err := renderConfig(ctx, addrpools_configMapFile, data)
	if err != nil {
		return errors.Wrapf(err, "Failed to render config to ConfigMap object")
	}

	name := obj.GetName()
	namespace := obj.GetNamespace()
	if name == "" {
		return errors.Errorf("Object %s has no name", obj.GroupVersionKind().String())
	}
	gvk := obj.GroupVersionKind()
	// used for logging and errors
	objDesc := fmt.Sprintf("(%s) %s/%s", gvk.String(), namespace, name)

	// Get existing
	existing := &uns.Unstructured{}
	existing.SetGroupVersionKind(gvk)
	err = r.Client.Get(ctx, types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}, existing)

	if err != nil && apierrors.IsNotFound(err) {
		r.Log.Info("Object not found create it")
		err = r.Client.Create(ctx, &obj)
		if err != nil {
			return errors.Wrapf(err, "could not create %s", objDesc)
		}
		return nil
	}
	if err != nil {
		return errors.Wrapf(err, "could not retrieve existing %s", objDesc)
	}

	if !reflect.DeepEqual(&obj, existing) {
		objPtr, err := mergeObjects(&obj, existing)
		if objPtr == nil || err != nil {
			return errors.Wrapf(err, "failed to merge configmaps")
		}
		if err := r.Client.Update(ctx, objPtr); err != nil {
			return errors.Wrapf(err, "could not update object %s", objDesc)
		}
	}

	return nil
}

func (r *AddressPoolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&metallbv1beta1.AddressPool{}).
		Complete(r)
}

/*
Copyright 2021 Luis Rascao.

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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"context"
	"path"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	relupv1alpha1 "github.com/lrascao/erlang-relup-k8s-operator/api/v1alpha1"
)

// ReleaseUpgradeReconciler reconciles a ReleaseUpgrade object
type ReleaseUpgradeReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=relup.lrascao.github.io,resources=releaseupgrades,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=relup.lrascao.github.io,resources=releaseupgrades/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=relup.lrascao.github.io,resources=releaseupgrades/finalizers,verbs=update

// +kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;create

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ReleaseUpgrade object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *ReleaseUpgradeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("releaseupgrade", req.NamespacedName)

	// Fetch the ReleaseUpgrade instance
	releaseupgrade := &relupv1alpha1.ReleaseUpgrade{}
	err := r.Get(ctx, req.NamespacedName, releaseupgrade)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("ReleaseUpgrade resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get ReleaseUpgrade")
		return ctrl.Result{}, err
	}

	log.Info("ReleaseUpgrade spec params", "Name: ", releaseupgrade.Name, ", relup: ", releaseupgrade.Spec.ImageSpec, "deployment: ", releaseupgrade.Spec.DeploymentSpec)

	// Check if the daemonset already exists, if not create a new one
	found := &appsv1.DaemonSet{}
	err = r.Get(ctx, types.NamespacedName{Name: releaseupgrade.Name, Namespace: releaseupgrade.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new daemonset
		daemonset := r.daemonSetForReleaseUpgrade(releaseupgrade)
		log.Info("Creating a new DaemonSet", "DaemonSet.Namespace", daemonset.Namespace, "DaemonSet.Name", daemonset.Name)
		err = r.Create(ctx, daemonset)
		if err != nil {
			log.Error(err, "Failed to create new DaemonSet", "DaemonSet.Namespace", daemonset.Namespace, "DaemonSet.Name", daemonset.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue - next time around the daemonset should already be there
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReleaseUpgradeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&relupv1alpha1.ReleaseUpgrade{}).
		Complete(r)
}

//
// private
//

// daemonSetForReleaseUpgrade returns a ReleaseUpgrade DaemonSet object
func (r ReleaseUpgradeReconciler) daemonSetForReleaseUpgrade(releaseupgrade *relupv1alpha1.ReleaseUpgrade) *appsv1.DaemonSet {
	labels := map[string]string{"app": "relup", "relup": releaseupgrade.Name}
	pathType := corev1.HostPathDirectoryOrCreate
	mountPath := path.Join("/tmp", releaseupgrade.Name)

	daemonset := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      releaseupgrade.Name,
			Namespace: releaseupgrade.Namespace,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:    releaseupgrade.Spec.ImageSpec.Name,
						Image:   releaseupgrade.Spec.ImageSpec.Image,
						Command: []string{"/bin/sh", "-c"},
						Args:    []string{"cp " + releaseupgrade.Spec.ImageSpec.Tarball + " " + mountPath + ";" + "sleep 3600"},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      releaseupgrade.Name,
							MountPath: mountPath,
						}},
					}},
					Volumes: []corev1.Volume{{
						Name: releaseupgrade.Name,
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: releaseupgrade.Spec.VolumeSpec.HostPath,
								Type: &pathType,
							},
						},
					}},
				},
			},
		},
	}
	// Set ReleaseUpgrade instance as the owner and controller of the DaemonSet
	ctrl.SetControllerReference(releaseupgrade, daemonset, r.Scheme)
	return daemonset
}

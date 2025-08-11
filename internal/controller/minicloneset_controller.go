/*
Copyright 2025.

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	appsexamplecomv1beta1 "k8s.openkruise.com/v1/api/v1beta1"
).k8s.io/controller-runtime/pkg/log"

	appsexamplecomv1beta1 "k8s.openkruise.com/v1/api/v1beta1" under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	appsexamplecomv1alpha1 "k8s.openkruise.com/v1/api/v1alpha1"
)

// MiniCloneSetReconciler reconciles a MiniCloneSet object
type MiniCloneSetReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps.example.com.my.domain,resources=miniclonesets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.example.com.my.domain,resources=miniclonesets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.example.com.my.domain,resources=miniclonesets/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MiniCloneSet object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *MiniCloneSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Fetch the MiniCloneSet instance
	var myCR appsexamplecomv1alpha1.MiniCloneSet
	if err := r.Get(ctx, req.NamespacedName, &myCR); err != nil {
		log.Error(err, "unable to fetch MiniCloneSet")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Log Replicas and Image as requested
	log.Info("Reconciling MiniCloneSet",
		"replicas", myCR.Spec.Replicas,
		"image", myCR.Spec.Image) // Fake status update: set AvailableReplicas = Replicas
	myCR.Status.AvailableReplicas = myCR.Spec.Replicas
	if err := r.Status().Update(ctx, &myCR); err != nil {
		log.Error(err, "failed to update MiniCloneSet status")
		return ctrl.Result{}, err
	}

	// Return empty result as requested
	return ctrl.Result{}, nil
}

// handleRollingUpdate implements rolling update strategy
func (r *MiniCloneSetReconciler) handleRollingUpdate(ctx context.Context, myCR *appsexamplecomv1alpha1.MiniCloneSet, podList *corev1.PodList, desiredReplicas int, desiredImage string) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Count pods by status
	readyPods := 0
	outdatedPods := []corev1.Pod{}

	for _, pod := range podList.Items {
		if isPodReady(&pod) {
			readyPods++
		}
		if !isPodUpToDate(&pod, desiredImage) {
			outdatedPods = append(outdatedPods, pod)
		}
	}

	// Scale up if we need more pods
	currentPods := len(podList.Items)
	if currentPods < desiredReplicas {
		for i := currentPods; i < desiredReplicas; i++ {
			pod := r.createPodForMiniCloneSet(myCR, i)
			if err := r.Create(ctx, pod); err != nil {
				log.Error(err, "failed to create pod", "pod", pod.Name)
				return ctrl.Result{}, err
			}
			log.Info("Created new pod", "pod", pod.Name)
		}
		// Requeue to check status
		return ctrl.Result{RequeueAfter: time.Second * 10}, nil
	}

	// Rolling update: replace outdated pods one by one
	if len(outdatedPods) > 0 {
		// Only update one pod at a time for rolling update
		podToUpdate := outdatedPods[0]

		// Create new pod first
		newPod := r.createPodForMiniCloneSet(myCR, len(podList.Items))
		if err := r.Create(ctx, newPod); err != nil {
			log.Error(err, "failed to create replacement pod", "pod", newPod.Name)
			return ctrl.Result{}, err
		}

		log.Info("Created replacement pod for rolling update", "newPod", newPod.Name, "oldPod", podToUpdate.Name)

		// Wait for new pod to be ready before deleting old one
		return ctrl.Result{RequeueAfter: time.Second * 5}, nil
	}

	// Scale down if we have too many pods
	if currentPods > desiredReplicas {
		podsToDelete := currentPods - desiredReplicas
		for i := 0; i < podsToDelete && i < len(podList.Items); i++ {
			pod := &podList.Items[i]
			if err := r.Delete(ctx, pod); err != nil {
				log.Error(err, "failed to delete pod", "pod", pod.Name)
				return ctrl.Result{}, err
			}
			log.Info("Deleted excess pod", "pod", pod.Name)
		}
		return ctrl.Result{RequeueAfter: time.Second * 5}, nil
	}

	// Update status
	myCR.Status.AvailableReplicas = readyPods
	if err := r.Status().Update(ctx, myCR); err != nil {
		log.Error(err, "failed to update MiniCloneSet status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// handleRecreateUpdate implements recreate update strategy
func (r *MiniCloneSetReconciler) handleRecreateUpdate(ctx context.Context, myCR *appsexamplecomv1alpha1.MiniCloneSet, podList *corev1.PodList, desiredReplicas int, desiredImage string) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// Check if any pods need updating
	needsUpdate := false
	for _, pod := range podList.Items {
		if !isPodUpToDate(&pod, desiredImage) {
			needsUpdate = true
			break
		}
	}

	if needsUpdate {
		// Delete all existing pods first
		for _, pod := range podList.Items {
			if err := r.Delete(ctx, &pod); err != nil {
				log.Error(err, "failed to delete pod during recreate", "pod", pod.Name)
				return ctrl.Result{}, err
			}
			log.Info("Deleted pod for recreate update", "pod", pod.Name)
		}

		// Wait for pods to be deleted before creating new ones
		return ctrl.Result{RequeueAfter: time.Second * 5}, nil
	}

	// Create new pods if needed
	currentPods := len(podList.Items)
	if currentPods < desiredReplicas {
		for i := currentPods; i < desiredReplicas; i++ {
			pod := r.createPodForMiniCloneSet(myCR, i)
			if err := r.Create(ctx, pod); err != nil {
				log.Error(err, "failed to create pod", "pod", pod.Name)
				return ctrl.Result{}, err
			}
			log.Info("Created new pod", "pod", pod.Name)
		}
		return ctrl.Result{RequeueAfter: time.Second * 10}, nil
	}

	// Update status
	readyPods := 0
	for _, pod := range podList.Items {
		if isPodReady(&pod) {
			readyPods++
		}
	}

	myCR.Status.AvailableReplicas = readyPods
	if err := r.Status().Update(ctx, myCR); err != nil {
		log.Error(err, "failed to update MiniCloneSet status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// createPodForMiniCloneSet creates a new pod based on the MiniCloneSet spec
func (r *MiniCloneSetReconciler) createPodForMiniCloneSet(myCR *appsexamplecomv1alpha1.MiniCloneSet, index int) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%d", myCR.Name, index),
			Namespace: myCR.Namespace,
			Labels: map[string]string{
				"app": myCR.Name,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "main",
					Image: myCR.Spec.Image,
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: 80,
							Protocol:      corev1.ProtocolTCP,
						},
					},
				},
			},
		},
	}

	// Set owner reference
	ctrl.SetControllerReference(myCR, pod, r.Scheme)
	return pod
}

// isPodReady checks if a pod is ready
func isPodReady(pod *corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

// isPodUpToDate checks if a pod is using the desired image
func isPodUpToDate(pod *corev1.Pod, desiredImage string) bool {
	for _, container := range pod.Spec.Containers {
		if container.Name == "main" && container.Image != desiredImage {
			return false
		}
	}
	return true
}

// SetupWithManager sets up the controller with the Manager.
func (r *MiniCloneSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsexamplecomv1alpha1.MiniCloneSet{}).
		Named("minicloneset").
		Complete(r)
}

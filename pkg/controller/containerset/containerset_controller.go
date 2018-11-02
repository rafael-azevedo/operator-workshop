package containerset

import (
	"context"
	"fmt"
	"log"

	csv1alpha1 "github.com/rafael-azevedo/operator-workshop/containerset/pkg/apis/cs/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Containerset Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileContainerset{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("containerset-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Containerset
	err = c.Watch(&source.Kind{Type: &csv1alpha1.Containerset{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Containerset
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &csv1alpha1.Containerset{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileContainerset{}

// ReconcileContainerset reconciles a Containerset object
type ReconcileContainerset struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Containerset object and makes changes based on the state read
// and what is in the Containerset.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileContainerset) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Printf("Reconciling Containerset %s/%s\n", request.Namespace, request.Name)

	// Fetch the Containerset instance
	instance := &csv1alpha1.Containerset{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Define a new Pod object

	pod := newPodForCR(instance)

	// Set Containerset instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	podList := &corev1.PodList{}
	lbs := map[string]string{
		"app":     instance.Name,
		"version": "v0.1",
		"env":     "dev",
	}
	labelSelector := labels.SelectorFromSet(lbs)

	listOps := &client.ListOptions{Namespace: instance.Namespace, LabelSelector: labelSelector}
	if err = r.client.List(context.TODO(), listOps, podList); err != nil {
		return reconcile.Result{}, err
	}

	lpods := len(podList.Items)
	for _, pod := range podList.Items {
		if pod.ObjectMeta.DeletionTimestamp != nil {
			lpods = lpods - 1
		}
		switch pod.Status.Phase {
		case corev1.PodSucceeded:
			lpods = lpods - 1
		case corev1.PodFailed:
			lpods = lpods - 1
		}

	}

	instance.Status.AvailableReplicas = lpods
	PodNames := []string{}
	for _, pod := range podList.Items {
		PodNames = append(PodNames, pod.ObjectMeta.Name)
	}
	instance.Status.PodNames = PodNames

	log.Printf("# of pods %v, # of replicas required %v\n", lpods, instance.Spec.Replicas)

	if lpods > instance.Spec.Replicas {
		diff := lpods - instance.Spec.Replicas
		dpods := podList.Items[:diff]
		for _, dpod := range dpods {
			err = r.client.Delete(context.TODO(), &dpod)
			if err != nil {
				log.Println(err)
				return reconcile.Result{}, err
			}
		}
		_ = r.client.Update(context.TODO(), instance)
		return reconcile.Result{Requeue: true}, nil
	}

	if lpods < instance.Spec.Replicas {
		log.Print("You need more replicas")
		_ = r.client.Update(context.TODO(), instance)
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			log.Println(err)
			return reconcile.Result{}, err
		}
	}

	// Pod already exists - don't requeue
	log.Printf("Skip reconcile: Pod replica %v matches spec %v already exists\n", lpods, instance.Spec.Replicas)
	return reconcile.Result{}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *csv1alpha1.Containerset) *corev1.Pod {
	labels := map[string]string{
		"app":     cr.Name,
		"version": "v0.1",
		"env":     "dev",
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", cr.Name),
			Namespace:    cr.Namespace,
			Labels:       labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}

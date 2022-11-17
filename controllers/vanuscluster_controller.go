/*
Copyright 2022.

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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/linkall-labs/vanus-operator/internal/resource"
	"github.com/linkall-labs/vanus-operator/internal/status"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	clientretry "k8s.io/client-go/util/retry"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	// "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	vanusv1 "github.com/linkall-labs/vanus-operator/api/v1"
)

// VanusClusterReconciler reconciles a VanusCluster object
type VanusClusterReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	Namespace     string
	Recorder      record.EventRecorder
	ClusterConfig *rest.Config
	Clientset     *kubernetes.Clientset
	// PodExecutor   PodExecutor
}

//+kubebuilder:rbac:groups=vanus.vanus.io,resources=vanusclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=vanus.vanus.io,resources=vanusclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=vanus.vanus.io,resources=vanusclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the VanusCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *VanusClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	logger := ctrl.LoggerFrom(ctx)

	vanusCluster, err := r.getVanusCluster(ctx, req.NamespacedName)

	if client.IgnoreNotFound(err) != nil {
		return ctrl.Result{}, err
	} else if k8serrors.IsNotFound(err) {
		// No need to requeue if the resource no longer exists
		return ctrl.Result{}, nil
	}

	instanceSpec, err := json.Marshal(vanusCluster.Spec)
	if err != nil {
		logger.Error(err, "Failed to marshal cluster spec")
	}

	logger.Info("Start reconciling",
		"spec", string(instanceSpec))

	resourceBuilder := resource.VanusResourceBuilder{
		Instance: vanusCluster,
		Scheme:   r.Scheme,
	}

	builders := resourceBuilder.ResourceBuilders()

	for _, builder := range builders {
		resource, err := builder.Build()
		if err != nil {
			return ctrl.Result{}, err
		}

		// only StatefulSetBuilder returns true
		if builder.UpdateMayRequireRecreate() {
			sts := resource.(*appsv1.StatefulSet)

			// current, err := r.statefulset(ctx, vanusCluster, "controller")
			// if client.IgnoreNotFound(err) != nil {
			// 	return ctrl.Result{}, err
			// }

			// only checks for scale down if statefulSet is created
			// else continue to CreateOrUpdate()
			if !k8serrors.IsNotFound(err) {
				if err := builder.Update(sts); err != nil {
					return ctrl.Result{}, err
				}
				// if r.scaleDown(ctx, vanusCluster, current, sts) {
				// 	// return when cluster scale down detected; unsupported operation
				// 	return ctrl.Result{}, nil
				// }
			}

			// The PVCs for the StatefulSet may require expanding
			// if err = r.reconcilePVC(ctx, vanusCluster, sts); err != nil {
			// 	r.setReconcileSuccess(ctx, vanusCluster, corev1.ConditionFalse, "FailedReconcilePVC", err.Error())
			// 	return ctrl.Result{}, err
			// }
		}

		var operationResult controllerutil.OperationResult
		err = clientretry.RetryOnConflict(clientretry.DefaultRetry, func() error {
			var apiError error
			operationResult, apiError = controllerutil.CreateOrUpdate(ctx, r.Client, resource, func() error {
				return builder.Update(resource)
			})
			return apiError
		})
		r.logAndRecordOperationResult(logger, vanusCluster, resource, operationResult, err)
		if err != nil {
			r.setReconcileSuccess(ctx, vanusCluster, corev1.ConditionFalse, "Error", err.Error())
			return ctrl.Result{}, err
		}

		// if err = r.annotateIfNeeded(ctx, logger, builder, operationResult, vanusCluster); err != nil {
		// 	return ctrl.Result{}, err
		// }
	}

	// if requeueAfter, err := r.restartStatefulSetIfNeeded(ctx, logger, vanusCluster); err != nil || requeueAfter > 0 {
	// 	return ctrl.Result{RequeueAfter: requeueAfter}, err
	// }

	// if err := r.reconcileStatus(ctx, vanusCluster); err != nil {
	// 	return ctrl.Result{}, err
	// }

	// By this point the StatefulSet may have finished deploying. Run any
	// post-deploy steps if so, or requeue until the deployment is finished.
	// if requeueAfter, err := r.runRabbitmqCLICommandsIfAnnotated(ctx, vanusCluster); err != nil || requeueAfter > 0 {
	// 	if err != nil {
	// 		r.setReconcileSuccess(ctx, vanusCluster, corev1.ConditionFalse, "FailedCLICommand", err.Error())
	// 	}
	// 	return ctrl.Result{RequeueAfter: requeueAfter}, err
	// }

	// Set ReconcileSuccess to true and update observedGeneration after all reconciliation steps have finished with no error
	// vanusCluster.Status.ObservedGeneration = vanusCluster.GetGeneration()
	// r.setReconcileSuccess(ctx, vanusCluster, corev1.ConditionTrue, "Success", "Finish reconciling")

	logger.Info("Finished reconciling")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VanusClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vanusv1.VanusCluster{}).
		Complete(r)
}

func (r *VanusClusterReconciler) getVanusCluster(ctx context.Context, namespacedName types.NamespacedName) (*vanusv1.VanusCluster, error) {
	vanusClusterInstance := &vanusv1.VanusCluster{}
	err := r.Get(ctx, namespacedName, vanusClusterInstance)
	return vanusClusterInstance, err
}

// logAndRecordOperationResult - helper function to log and record events with message and error
// it logs and records 'updated' and 'created' OperationResult, and ignores OperationResult 'unchanged'
func (r *VanusClusterReconciler) logAndRecordOperationResult(logger logr.Logger, rmq runtime.Object, resource runtime.Object, operationResult controllerutil.OperationResult, err error) {
	if operationResult == controllerutil.OperationResultNone && err == nil {
		return
	}

	var operation string
	if operationResult == controllerutil.OperationResultCreated {
		operation = "create"
	}
	if operationResult == controllerutil.OperationResultUpdated {
		operation = "update"
	}

	if err == nil {
		msg := fmt.Sprintf("%sd resource %s of Type %T", operation, resource.(metav1.Object).GetName(), resource.(metav1.Object))
		logger.Info(msg)
		r.Recorder.Event(rmq, corev1.EventTypeNormal, fmt.Sprintf("Successful%s", strings.Title(operation)), msg)
	}

	if err != nil {
		msg := fmt.Sprintf("failed to %s resource %s of Type %T", operation, resource.(metav1.Object).GetName(), resource.(metav1.Object))
		logger.Error(err, msg)
		r.Recorder.Event(rmq, corev1.EventTypeWarning, fmt.Sprintf("Failed%s", strings.Title(operation)), msg)
	}
}

// func (r *RabbitmqClusterReconciler) updateStatusConditions(ctx context.Context, rmq *rabbitmqv1beta1.RabbitmqCluster) (time.Duration, error) {
// 	logger := ctrl.LoggerFrom(ctx)
// 	childResources, err := r.getChildResources(ctx, rmq)
// 	if err != nil {
// 		return 0, err
// 	}

// 	oldConditions := make([]status.RabbitmqClusterCondition, len(rmq.Status.Conditions))
// 	copy(oldConditions, rmq.Status.Conditions)
// 	rmq.Status.SetConditions(childResources)

// 	if !reflect.DeepEqual(rmq.Status.Conditions, oldConditions) {
// 		if err = r.Status().Update(ctx, rmq); err != nil {
// 			if k8serrors.IsConflict(err) {
// 				logger.Info("failed to update status because of conflict; requeueing...",
// 					"namespace", rmq.Namespace,
// 					"name", rmq.Name)
// 				return 2 * time.Second, nil
// 			}
// 			return 0, err
// 		}
// 	}
// 	return 0, nil
// }

// func (r *RabbitmqClusterReconciler) getChildResources(ctx context.Context, rmq *rabbitmqv1beta1.RabbitmqCluster) ([]runtime.Object, error) {
// 	sts := &appsv1.StatefulSet{}
// 	endPoints := &corev1.Endpoints{}

// 	if err := r.Client.Get(ctx,
// 		types.NamespacedName{Name: rmq.ChildResourceName("server"), Namespace: rmq.Namespace},
// 		sts); err != nil && !k8serrors.IsNotFound(err) {
// 		return nil, err
// 	} else if k8serrors.IsNotFound(err) {
// 		sts = nil
// 	}

// 	if err := r.Client.Get(ctx,
// 		types.NamespacedName{Name: rmq.ChildResourceName(resource.ServiceSuffix), Namespace: rmq.Namespace},
// 		endPoints); err != nil && !k8serrors.IsNotFound(err) {
// 		return nil, err
// 	} else if k8serrors.IsNotFound(err) {
// 		endPoints = nil
// 	}

// 	return []runtime.Object{sts, endPoints}, nil
// }

func (r *VanusClusterReconciler) setReconcileSuccess(ctx context.Context, vanusCluster *vanusv1.VanusCluster, condition corev1.ConditionStatus, reason, msg string) {
	vanusCluster.Status.SetCondition(status.ReconcileSuccess, condition, reason, msg)
	if writerErr := r.Status().Update(ctx, vanusCluster); writerErr != nil {
		ctrl.LoggerFrom(ctx).Error(writerErr, "Failed to update Custom Resource status",
			"namespace", vanusCluster.Namespace,
			"name", vanusCluster.Name)
	}
}

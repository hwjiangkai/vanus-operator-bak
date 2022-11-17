package controllers

import (
	"context"

	vanusv1 "github.com/linkall-labs/vanus-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *VanusClusterReconciler) statefulset(ctx context.Context, rmq *vanusv1.VanusCluster, resourceName string) (*appsv1.StatefulSet, error) {
	sts := &appsv1.StatefulSet{}
	if err := r.Get(ctx, types.NamespacedName{Name: rmq.ChildResourceName(resourceName), Namespace: rmq.Namespace}, sts); err != nil {
		return nil, err
	}
	return sts, nil
}

func (r *VanusClusterReconciler) deployment(ctx context.Context, rmq *vanusv1.VanusCluster, resourceName string) (*appsv1.Deployment, error) {
	dep := &appsv1.Deployment{}
	if err := r.Get(ctx, types.NamespacedName{Name: rmq.ChildResourceName(resourceName), Namespace: rmq.Namespace}, dep); err != nil {
		return nil, err
	}
	return dep, nil
}

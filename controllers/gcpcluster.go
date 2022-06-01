package controllers

import (
	"context"

	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	capg "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/annotations"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const FinalizerDNS = "dns-operator-gcp.finalizers.giantswarm.io"

//counterfeiter:generate . GCPClusterClient
type GCPClusterClient interface {
	Get(context.Context, types.NamespacedName) (*capg.GCPCluster, error)
	GetOwner(context.Context, *capg.GCPCluster) (*capi.Cluster, error)
	AddFinalizer(context.Context, *capg.GCPCluster, string) error
	RemoveFinalizer(context.Context, *capg.GCPCluster, string) error
}

//counterfeiter:generate . Registrar
type Registrar interface {
	Register(context.Context, *capg.GCPCluster) error
	Unregister(context.Context, *capg.GCPCluster) error
}

type GCPClusterReconciler struct {
	client     GCPClusterClient
	registrars []Registrar
}

func NewGCPClusterReconciler(client GCPClusterClient, registrars []Registrar) *GCPClusterReconciler {
	return &GCPClusterReconciler{
		client:     client,
		registrars: registrars,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *GCPClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capg.GCPCluster{}).
		Complete(r)
}

func (r *GCPClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Reconciling")
	defer logger.Info("Done reconciling")

	gcpCluster, err := r.client.Get(ctx, req.NamespacedName)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("GCP Cluster no longer exists")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, microerror.Mask(err)
	}

	cluster, err := r.client.GetOwner(ctx, gcpCluster)
	if err != nil {
		return ctrl.Result{}, microerror.Mask(err)
	}

	if cluster == nil {
		logger.Info("GCP Cluster does not have an owner cluster yet")
		return ctrl.Result{}, nil
	}

	if annotations.IsPaused(cluster, gcpCluster) {
		logger.Info("Infrastructure or core cluster is marked as paused. Won't reconcile")
		return ctrl.Result{}, nil
	}

	if !gcpCluster.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, gcpCluster)
	}

	return r.reconcileNormal(ctx, gcpCluster)
}

func (r *GCPClusterReconciler) reconcileNormal(ctx context.Context, gcpCluster *capg.GCPCluster) (ctrl.Result, error) {
	err := r.client.AddFinalizer(ctx, gcpCluster, FinalizerDNS)
	if err != nil {
		return ctrl.Result{}, microerror.Mask(err)
	}

	for _, registrar := range r.registrars {
		err = registrar.Register(ctx, gcpCluster)
		if err != nil {
			return ctrl.Result{}, microerror.Mask(err)
		}
	}

	return ctrl.Result{}, nil
}

func (r *GCPClusterReconciler) reconcileDelete(ctx context.Context, gcpCluster *capg.GCPCluster) (ctrl.Result, error) {
	for i := range r.registrars {
		registrar := r.registrars[len(r.registrars)-1-i]

		err := registrar.Unregister(ctx, gcpCluster)
		if err != nil {
			return ctrl.Result{}, microerror.Mask(err)
		}
	}

	err := r.client.RemoveFinalizer(ctx, gcpCluster, FinalizerDNS)
	if err != nil {
		return ctrl.Result{}, microerror.Mask(err)
	}

	return ctrl.Result{}, nil
}

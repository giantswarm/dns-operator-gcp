package controllers

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	capg "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/annotations"
	ctrl "sigs.k8s.io/controller-runtime"
)

const FinalizerDNS = "dns-operator-gcp.finalizers.giantswarm.io"

//counterfeiter:generate . GCPClusterClient
type GCPClusterClient interface {
	Get(context.Context, types.NamespacedName) (*capg.GCPCluster, error)
	GetOwner(context.Context, *capg.GCPCluster) (*capi.Cluster, error)
	AddFinalizer(context.Context, *capg.GCPCluster, string) error
	RemoveFinalizer(context.Context, *capg.GCPCluster, string) error
}

//counterfeiter:generate . CloudDNSClient
type CloudDNSClient interface {
	CreateZone(context.Context, *capg.GCPCluster) error
	CreateARecords(context.Context, *capg.GCPCluster) error
	DeleteARecords(context.Context, *capg.GCPCluster) error
	DeleteZone(context.Context, *capg.GCPCluster) error
}

type GCPClusterReconciler struct {
	logger    logr.Logger
	client    GCPClusterClient
	dnsClient CloudDNSClient
}

func NewGCPClusterReconciler(logger logr.Logger, client GCPClusterClient, dnsClient CloudDNSClient) *GCPClusterReconciler {
	return &GCPClusterReconciler{
		logger:    logger,
		client:    client,
		dnsClient: dnsClient,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *GCPClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capg.GCPCluster{}).
		Complete(r)
}

func (r *GCPClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	gcpCluster, err := r.client.Get(ctx, req.NamespacedName)
	if err != nil {
		if errors.IsNotFound(err) {
			r.logger.Info("GCP Cluster no longer exists")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, microerror.Mask(err)
	}
	cluster, err := r.client.GetOwner(ctx, gcpCluster)
	if err != nil {
		return ctrl.Result{}, microerror.Mask(err)
	}

	if cluster == nil {
		r.logger.Info("GCP Cluster does not have an owner cluster yet")
		return ctrl.Result{}, nil
	}

	if annotations.IsPaused(cluster, gcpCluster) {
		r.logger.Info("infrastructure or core cluster is marked as paused. Won't reconcile")
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

	err = r.dnsClient.CreateZone(ctx, gcpCluster)
	if err != nil {
		return ctrl.Result{}, microerror.Mask(err)
	}

	err = r.dnsClient.CreateARecords(ctx, gcpCluster)
	if err != nil {
		return ctrl.Result{}, microerror.Mask(err)
	}

	return ctrl.Result{}, nil
}

func (r *GCPClusterReconciler) reconcileDelete(ctx context.Context, gcpCluster *capg.GCPCluster) (ctrl.Result, error) {
	err := r.dnsClient.DeleteARecords(ctx, gcpCluster)
	if err != nil {
		return ctrl.Result{}, microerror.Mask(err)
	}

	err = r.dnsClient.DeleteZone(ctx, gcpCluster)
	if err != nil {
		return ctrl.Result{}, microerror.Mask(err)
	}

	err = r.client.RemoveFinalizer(ctx, gcpCluster, FinalizerDNS)
	if err != nil {
		return ctrl.Result{}, microerror.Mask(err)
	}

	return ctrl.Result{}, nil
}

package k8sclient

import (
	"context"
	"errors"
	"fmt"

	"github.com/giantswarm/microerror"
	capg "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const BastionLabelKey = "cluster.x-k8s.io/deployment-name"

type Bastions struct {
	client    client.Client
	finalizer string
}

func NewBastions(client client.Client, finalizer string) *Bastions {
	return &Bastions{
		client:    client,
		finalizer: finalizer,
	}
}

func (b *Bastions) GetBastionIPList(ctx context.Context, cluster *capg.GCPCluster) ([]string, error) {
	machineList, err := b.getBastionMachineList(ctx, cluster)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var bastionPublicIPList []string

	for _, machine := range machineList.Items {
		// no ip address yet
		if len(machine.Status.Addresses) == 0 {
			return nil, errors.New("bastion IP is not yet available")
		}

		// get the public IP
		for _, addr := range machine.Status.Addresses {
			if addr.Type == "ExternalIP" {
				bastionPublicIPList = append(bastionPublicIPList, addr.Address)
				break
			}
		}
	}

	return bastionPublicIPList, nil
}

func (b *Bastions) getBastionMachineList(ctx context.Context, cluster *capg.GCPCluster) (*capg.GCPMachineList, error) {
	machineList := &capg.GCPMachineList{}
	err := b.client.List(
		ctx,
		machineList,
		client.InNamespace(cluster.Namespace),
		client.MatchingLabels{BastionLabelKey: bastionLabel(cluster.Name)},
	)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return machineList, nil
}

func bastionLabel(clusterName string) string {
	return fmt.Sprintf("%s-bastion", clusterName)
}

func (b *Bastions) AddFinalizerToBastions(ctx context.Context, cluster *capg.GCPCluster) error {
	machineList, err := b.getBastionMachineList(ctx, cluster)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, bastion := range machineList.Items {
		if !bastion.DeletionTimestamp.IsZero() {
			continue
		}

		original := bastion.DeepCopy()
		controllerutil.AddFinalizer(&bastion, b.finalizer)
		err := b.client.Patch(ctx, &bastion, client.MergeFrom(original))

		if err != nil {
			return microerror.Mask(err)
		}
	}
	return nil
}

func (b *Bastions) RemoveFinalizerFromBastions(ctx context.Context, cluster *capg.GCPCluster) error {
	machineList, err := b.getBastionMachineList(ctx, cluster)
	if err != nil {
		return microerror.Mask(err)
	}
	for _, bastion := range machineList.Items {
		original := bastion.DeepCopy()
		controllerutil.RemoveFinalizer(&bastion, b.finalizer)
		err := b.client.Patch(ctx, &bastion, client.MergeFrom(original))

		if err != nil {
			return microerror.Mask(err)
		}
	}
	return nil
}

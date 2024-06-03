package set

import (
	"context"
	"fmt"

	ocsv1 "github.com/red-hat-storage/ocs-operator/api/v4/v1"
	"github.com/red-hat-storage/odf-cli/cmd/odf/root"
	"github.com/rook/kubectl-rook-ceph/pkg/logging"
	rookcephv1 "github.com/rook/rook/pkg/apis/ceph.rook.io/v1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	mcsv1a1 "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"
)

var multiClusterServiceCmd = &cobra.Command{
	Args:               cobra.ExactArgs(1),
	Use:                "multiclusterservice",
	Short:              "Enable MultiClusterService for StorageCluster",
	DisableFlagParsing: true,
	Example:            "odf set multiclusterservice <ClusterID> --namespace <StorageClusterNamespace>",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		if !isMultiClusterServiceAvailable(root.ClientSets.Kube) {
			logging.Fatal(fmt.Errorf("MultiClusterService API is not available on this cluster. Please ensure that Submariner is installed and Globalnet is enabled."))
		}

		clusterID := args[0]
		err := enableMultiClusterService(ctx, root.CtrlClient, root.StorageClusterNamespace, clusterID)
		if err != nil {
			logging.Fatal(err)
		}
	},
}

func isMultiClusterServiceAvailable(k8sclient kubernetes.Interface) bool {
	_, apis, err := k8sclient.Discovery().ServerGroupsAndResources()
	if err != nil && !discovery.IsGroupDiscoveryFailedError(err) {
		logging.Fatal(err)
	}
	if discovery.IsGroupDiscoveryFailedError(err) {
		e := err.(*discovery.ErrGroupDiscoveryFailed)
		if _, exists := e.Groups[mcsv1a1.SchemeGroupVersion]; exists {
			return true
		}
	}
	for _, api := range apis {
		if api.GroupVersion == mcsv1a1.GroupVersion.String() {
			for _, resource := range api.APIResources {
				if resource.Name == "serviceexports" {
					return true
				}
			}
		}
	}
	return false
}

func enableMultiClusterService(ctx context.Context, client ctrl.Client, storageClusterNamespace, clusterID string) error {
	var storageClusterList ocsv1.StorageClusterList
	err := client.List(ctx, &storageClusterList, &ctrl.ListOptions{Namespace: storageClusterNamespace})
	if err != nil {
		return err
	}
	if len(storageClusterList.Items) == 0 {
		return fmt.Errorf("No StorageCluster found in namespace %q.", storageClusterNamespace)
	}
	if len(storageClusterList.Items) > 1 {
		return fmt.Errorf("Multiple StorageClusters found in namespace %q. Expected 1 but found %d.", storageClusterNamespace, len(storageClusterList.Items))
	}
	storageClusterName := storageClusterList.Items[0].Name

	var storageCluster ocsv1.StorageCluster
	err = client.Get(ctx, types.NamespacedName{Namespace: storageClusterNamespace, Name: storageClusterName}, &storageCluster)
	if err != nil {
		return err
	}
	if storageCluster.Spec.Network != nil {
		clusterID := storageCluster.Spec.Network.MultiClusterService.ClusterID
		if clusterID != "" {
			return fmt.Errorf("ClusterID for MultiClusterService is already set to %q.", clusterID)
		}
	}
	if storageCluster.Spec.Network == nil {
		storageCluster.Spec.Network = &rookcephv1.NetworkSpec{}
	}
	storageCluster.Spec.Network.MultiClusterService.Enabled = true
	storageCluster.Spec.Network.MultiClusterService.ClusterID = clusterID
	err = client.Update(ctx, &storageCluster)
	if err != nil {
		return err
	}
	logging.Info("Successfully set 'multiClusterService' on StorageCluster %q in namespace %q.", storageClusterName, storageClusterNamespace)
	return nil
}

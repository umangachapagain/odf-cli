package set

import (
	"context"
	"testing"

	ocsv1 "github.com/red-hat-storage/ocs-operator/api/v4/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgo "k8s.io/client-go/discovery/fake"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client/fake"
	mcsv1a1 "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"
)

func Test_enableMultiClusterService(t *testing.T) {
	var scheme = runtime.NewScheme()
	err := ocsv1.AddToScheme(scheme)
	assert.Nil(t, err)

	client := ctrl.NewClientBuilder().WithScheme(scheme).Build()

	err = enableMultiClusterService(context.TODO(), client, "openshift-storage", "test")
	assert.NotNil(t, err)

	storagecluster1 := ocsv1.StorageCluster{ObjectMeta: metav1.ObjectMeta{Name: "ocs-storagecluster", Namespace: "openshift-storage"}, Spec: ocsv1.StorageClusterSpec{}}
	err = client.Create(context.TODO(), &storagecluster1)
	assert.Nil(t, err)
	err = enableMultiClusterService(context.TODO(), client, "openshift-storage", "test")
	assert.Nil(t, err)

	err = enableMultiClusterService(context.TODO(), client, "openshift-storage", "new-test")
	assert.NotNil(t, err)

	storagecluster2 := ocsv1.StorageCluster{ObjectMeta: metav1.ObjectMeta{Name: "another-ocs-storagecluster", Namespace: "openshift-storage"}, Spec: ocsv1.StorageClusterSpec{}}
	err = client.Create(context.TODO(), &storagecluster2)
	assert.Nil(t, err)
	err = enableMultiClusterService(context.TODO(), client, "openshift-storage", "test")
	assert.NotNil(t, err)
}

func Test_isMultiClusterServiceAvailable(t *testing.T) {
	client := fakeclientset.NewSimpleClientset()
	fakeDiscovery, ok := client.Discovery().(*clientgo.FakeDiscovery)
	assert.True(t, ok)

	isAvailable := isMultiClusterServiceAvailable(client)
	assert.False(t, isAvailable)

	fakeDiscovery.Resources = []*metav1.APIResourceList{}

	isAvailable = isMultiClusterServiceAvailable(client)
	assert.False(t, isAvailable)

	fakeDiscovery.Resources = []*metav1.APIResourceList{
		{
			GroupVersion: mcsv1a1.GroupVersion.String(),
			APIResources: []metav1.APIResource{
				{
					Name: "serviceexports",
				},
			},
		},
	}

	isAvailable = isMultiClusterServiceAvailable(client)
	assert.True(t, isAvailable)
}

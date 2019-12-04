package test

import (
	"context"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	reconcilertesting "github.com/knative/eventing/pkg/reconciler/testing"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

// Tests Must Set Client Before These Custom Verifications Will Work
var MockClient *reconcilertesting.MockClient

// VerifyWantDeploymentPresent verifies that the client contains the deployment resource specified in OtherTestData
// it utilizes custom comparator options to avoid issues comparing resource quantities.
func VerifyWantDeploymentPresent(otherTestDataKey string) func(t *testing.T, tc *reconcilertesting.TestCase) {
	return func(t *testing.T, tc *reconcilertesting.TestCase) {
		verifyWantDeploymentPresent(t, tc, otherTestDataKey)
	}
}

// VerifyWantDeploymentPresent verifies that the client contains the deployment resource specified in OtherTestData
// it utilizes custom comparator options to avoid issues comparing resource quantities.
func verifyWantDeploymentPresent(t *testing.T, tc *reconcilertesting.TestCase, testDataName string) {
	expectedDeployment := tc.OtherTestData[testDataName].(*v1.Deployment)
	o, err := scheme.Scheme.New(expectedDeployment.GetObjectKind().GroupVersionKind())
	if err != nil {
		t.Errorf("error creating a copy of %T: %v", expectedDeployment, err)
	}
	acc, err := meta.Accessor(expectedDeployment)
	if err != nil {
		t.Errorf("error getting accessor for %#v %v", expectedDeployment, err)
	}
	err = MockClient.Get(context.TODO(), client.ObjectKey{Namespace: acc.GetNamespace(), Name: acc.GetName()}, o)
	if err != nil {
		t.Errorf("error getting %T %s/%s: %v", expectedDeployment, acc.GetNamespace(), acc.GetName(), err)
	}

	diffOpts := cmp.Options{
		// Ignore TypeMeta, since the objects created by the controller won't have
		// it
		cmpopts.IgnoreTypes(metav1.TypeMeta{}),
		cmpopts.IgnoreUnexported(resource.Quantity{}),
	}

	if tc.IgnoreTimes {
		// Ignore VolatileTime fields, since they rarely compare correctly.
		diffOpts = append(diffOpts, cmpopts.IgnoreTypes(apis.VolatileTime{}))
	}

	if diff := cmp.Diff(expectedDeployment, o, diffOpts...); diff != "" {
		t.Errorf("Unexpected present %T %s/%s (-want +got):\n%v", expectedDeployment, acc.GetNamespace(), acc.GetName(), diff)
	}
}

package operator

import (
	"context"
	"testing"

	appsv1alpha1 "github.com/3scale/3scale-operator/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/reconcilers"

	appsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func TestSystemPostgreSQLReconcilerCreate(t *testing.T) {
	var (
		appLabel       = "someLabel"
		name           = "example-apimanager"
		namespace      = "operator-unittest"
		trueValue      = true
		imageURL       = "postgresql:test"
		wildcardDomain = "test.3scale.net"
		tenantName     = "someTenant"
		log            = logf.Log.WithName("operator_test")
	)

	ctx := context.TODO()

	apimanager := &appsv1alpha1.APIManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1alpha1.APIManagerSpec{
			APIManagerCommonSpec: appsv1alpha1.APIManagerCommonSpec{
				AppLabel:                     &appLabel,
				ImageStreamTagImportInsecure: &trueValue,
				ResourceRequirementsEnabled:  &trueValue,
				WildcardDomain:               wildcardDomain,
				TenantName:                   &tenantName,
			},
			System: &appsv1alpha1.SystemSpec{
				DatabaseSpec: &appsv1alpha1.SystemDatabaseSpec{
					PostgreSQL: &appsv1alpha1.SystemPostgreSQLSpec{
						Image: &imageURL,
					},
				},
			},
		},
	}
	s := scheme.Scheme
	s.AddKnownTypes(appsv1alpha1.GroupVersion, apimanager)
	err := imagev1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}

	// Objects to track in the fake client.
	objs := []runtime.Object{}

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)
	clientAPIReader := fake.NewFakeClient(objs...)
	clientset := fakeclientset.NewSimpleClientset()
	recorder := record.NewFakeRecorder(10000)

	baseReconciler := reconcilers.NewBaseReconciler(ctx, cl, s, clientAPIReader, log, clientset.Discovery(), recorder)
	baseAPIManagerLogicReconciler := NewBaseAPIManagerLogicReconciler(baseReconciler, apimanager)

	reconciler := NewSystemPostgreSQLReconciler(baseAPIManagerLogicReconciler)
	_, err = reconciler.Reconcile()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		testName string
		objName  string
		obj      client.Object
	}{
		{"systemPostgreSQL_DC", "system-postgresql", &appsv1.DeploymentConfig{}},
		{"systemPostgreSQL_Service", "system-postgresql", &v1.Service{}},
		{"systemPostgreSQL_PVC", "postgresql-data", &v1.PersistentVolumeClaim{}},
		{"systemDatabaseSecret", component.SystemSecretSystemDatabaseSecretName, &v1.Secret{}},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			obj := tc.obj
			namespacedName := types.NamespacedName{
				Name:      tc.objName,
				Namespace: namespace,
			}
			err = cl.Get(context.TODO(), namespacedName, obj)
			// object must exist, that is all required to be tested
			if err != nil {
				subT.Errorf("error fetching object %s: %v", tc.objName, err)
			}
		})
	}
}

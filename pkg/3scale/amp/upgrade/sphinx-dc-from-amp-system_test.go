package upgrade

import (
	"context"
	"testing"
	"time"

	appsv1 "github.com/openshift/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
)

func TestSphinxFromAMPSystemImage(t *testing.T) {
	var (
		log         = logf.Log.WithName("upgrade_test")
		namespace   = "operator-unittest"
		ctx         = context.TODO()
		s           = scheme.Scheme
		sphinxDCKey = client.ObjectKey{Name: "system-sphinx", Namespace: namespace}
	)

	err := appsv1.AddToScheme(s)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Sphinx DC not found", func(subT *testing.T) {
		objs := []runtime.Object{}
		cl := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
		_, err := SphinxFromAMPSystemImage(ctx, cl, sphinxDCKey, log)
		if err != nil {
			subT.Fatal(err)
		}
	})

	t.Run("Searchd DC found", func(subT *testing.T) {
		opts := &component.SystemSphinxOptions{}
		sphinx := component.NewSystemSphinx(opts)

		newSphinxDC := sphinx.DeploymentConfig()
		newSphinxDC.Namespace = namespace
		sphinxDCKey := client.ObjectKeyFromObject(newSphinxDC)

		objs := []runtime.Object{newSphinxDC}
		cl := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
		res, err := SphinxFromAMPSystemImage(ctx, cl, sphinxDCKey, log)
		if err != nil {
			subT.Fatal(err)
		}

		if res.Requeue {
			subT.Fatal("requeue scheduled and not expected")
		}

		newSearchdDC := &appsv1.DeploymentConfig{}
		err = cl.Get(ctx, sphinxDCKey, newSearchdDC)
		// object must exist, that is all required to be tested
		if err != nil {
			subT.Fatalf("error fetching object %s: %v", sphinxDCKey, err)
		}
	})

	t.Run("Old Sphinx DC found", func(subT *testing.T) {
		opts := &component.SystemSphinxOptions{}
		sphinx := component.NewSystemSphinx(opts)

		newSphinxDC := sphinx.DeploymentConfig()
		newSphinxDC.Namespace = namespace
		sphinxDCKey := client.ObjectKeyFromObject(newSphinxDC)

		oldSphinxDC := basicOldSphinxDC(sphinxDCKey.Name, sphinxDCKey.Namespace)

		objs := []runtime.Object{oldSphinxDC}
		cl := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
		res, err := SphinxFromAMPSystemImage(ctx, cl, sphinxDCKey, log)
		if err != nil {
			subT.Fatal(err)
		}

		if !res.Requeue {
			subT.Fatal("requeue not scheduled and it was expected")
		}

		// object should have been deleted
		newSearchdDC := &appsv1.DeploymentConfig{}
		err = cl.Get(ctx, sphinxDCKey, newSearchdDC)
		if err == nil {
			subT.Fatalf("reading an object expected to be deleted: %s", sphinxDCKey)
		}

		if !errors.IsNotFound(err) {
			subT.Fatalf("unexpected error reading object %s: %v", sphinxDCKey, err)
		}
	})

	t.Run("Old Sphinx DC in deleting state found", func(subT *testing.T) {
		opts := &component.SystemSphinxOptions{}
		sphinx := component.NewSystemSphinx(opts)

		newSphinxDC := sphinx.DeploymentConfig()
		newSphinxDC.Namespace = namespace
		sphinxDCKey := client.ObjectKeyFromObject(newSphinxDC)

		oldSphinxDC := basicOldSphinxDC(sphinxDCKey.Name, sphinxDCKey.Namespace)
		now := metav1.NewTime(time.Now())
		oldSphinxDC.SetDeletionTimestamp(&now)

		objs := []runtime.Object{oldSphinxDC}
		cl := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
		res, err := SphinxFromAMPSystemImage(ctx, cl, sphinxDCKey, log)
		if err != nil {
			subT.Fatal(err)
		}

		if !res.Requeue {
			subT.Fatal("requeue not scheduled and it was expected")
		}

		// object should have not been deleted
		newSearchdDC := &appsv1.DeploymentConfig{}
		err = cl.Get(ctx, sphinxDCKey, newSearchdDC)
		if err != nil {
			subT.Fatalf("error fetching object %s: %v", sphinxDCKey, err)
		}
	})
}

func basicOldSphinxDC(name, namespace string) *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentConfigSpec{
			Triggers: appsv1.DeploymentTriggerPolicies{
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerOnConfigChange,
				},
				appsv1.DeploymentTriggerPolicy{
					Type: appsv1.DeploymentTriggerOnImageChange,
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic: true,
						ContainerNames: []string{
							"system-sphinx",
						},
						From: v1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: "system-system:2.X",
						},
					},
				},
			},
		},
	}
}

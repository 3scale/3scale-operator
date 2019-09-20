package apimanager

import (
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type BaseUpgrade struct {
	fromVersion     string
	toVersion       string
	cr              *appsv1alpha1.APIManager
	client          client.Client
	logger          logr.Logger
	apiClientReader client.Reader
	scheme          *runtime.Scheme
}

type Upgrade interface {
	Upgrade() (reconcile.Result, error)
}

package operator

import (
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type BaseReconciler struct {
	// client should be a split client that reads objects from
	// the cache and writes to the Kubernetes APIServer
	client client.Client
	// apiClientReader should be a client that directly reads objects
	// from the Kubernetes APIServer
	apiClientReader client.Reader
	scheme          *runtime.Scheme
	logger          logr.Logger
	discoveryClient discovery.DiscoveryInterface
}

func NewBaseReconciler(client client.Client, apiClientReader client.Reader, scheme *runtime.Scheme, logger logr.Logger, dicoveryClient discovery.DiscoveryInterface) BaseReconciler {
	return BaseReconciler{
		client:          client,
		apiClientReader: apiClientReader,
		scheme:          scheme,
		logger:          logger,
		discoveryClient: dicoveryClient,
	}
}

func (b *BaseReconciler) Client() client.Client {
	return b.client
}

func (b *BaseReconciler) APIClientReader() client.Reader {
	return b.apiClientReader
}

func (b *BaseReconciler) Scheme() *runtime.Scheme {
	return b.scheme
}

func (b *BaseReconciler) Logger() logr.Logger {
	return b.logger
}

func (b *BaseReconciler) DiscoveryClient() discovery.DiscoveryInterface {
	return b.discoveryClient
}

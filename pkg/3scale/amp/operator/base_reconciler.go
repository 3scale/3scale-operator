package operator

import (
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	restclient "k8s.io/client-go/rest"
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
	cfg             *restclient.Config
}

func NewBaseReconciler(client client.Client, apiClientReader client.Reader, scheme *runtime.Scheme, logger logr.Logger, cfg *restclient.Config) BaseReconciler {
	return BaseReconciler{
		client:          client,
		apiClientReader: apiClientReader,
		scheme:          scheme,
		logger:          logger,
		cfg:             cfg,
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

func (b *BaseReconciler) Config() *restclient.Config {
	return b.cfg
}

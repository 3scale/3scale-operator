package upgrade

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	appsv1 "github.com/openshift/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/3scale/3scale-operator/pkg/helper"
)

func SphinxFromAMPSystemImage(ctx context.Context, cl client.Client, dcKey client.ObjectKey, logger logr.Logger) (reconcile.Result, error) {
	logger.V(1).Info("Upgrade SphinxFromAMPSystemImage", "dc obj", dcKey)
	existingDC := &appsv1.DeploymentConfig{}
	if err := cl.Get(ctx, dcKey, existingDC); err != nil {
		if errors.IsNotFound(err) {
			// Nothing to upgrade
			logger.V(1).Info("Upgrade SphinxFromAMPSystemImage: object not found")
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Look at the image in the triggers
	// triggers:
	//	- type: ConfigChange
	//	- type: ImageChange
	//	  imageChangeParams:
	//		automatic: true
	//		containerNames:
	//		- system-master-svc
	//		- system-sphinx
	//		from:
	//		  kind: ImageStreamTag
	//		  name: amp-system:2.X <-- HERE
	//	      namespace: NAMESPACE
	//		lastTriggeredImage: quay.io/3scale/porta@sha256:3269ea78d162f33e50492445a893869cb386e5f80febca2b30b7899c24e9145b

	triggerImageChangePos, err := helper.FindDeploymentTriggerOnImageChange(existingDC.Spec.Triggers)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("unexpected: '%s' in DeploymentConfig '%s'", err, existingDC.Name)
	}

	imageChangeParams := existingDC.Spec.Triggers[triggerImageChangePos].ImageChangeParams

	if strings.HasPrefix(imageChangeParams.From.Name, "system-searchd") {
		// Nothing to upgrade
		logger.V(1).Info("Upgrade SphinxFromAMPSystemImage: new image found, nothing to upgrade")
		return reconcile.Result{}, nil
	}

	if helper.IsDeploymentConfigDeleting(existingDC) {
		// Already deleted, requeue
		logger.V(1).Info("Upgrade SphinxFromAMPSystemImage: DC still deleting, requeue")
		return reconcile.Result{Requeue: true}, nil
	}

	err = cl.Delete(ctx, existingDC)
	logger.Info("Upgrade SphinxFromAMPSystemImage", "delete dc obj", dcKey, "err", err)

	return reconcile.Result{Requeue: true}, err
}

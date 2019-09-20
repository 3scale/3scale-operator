package apimanager

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v2"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/operator"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Upgrade26_to_27 struct {
	BaseUpgrade
}

func (u *Upgrade26_to_27) Upgrade() (reconcile.Result, error) {
	res, err := u.upgradeAMPImageStreams()
	if res.Requeue || err != nil {
		return res, err
	}

	res, err = u.upgradeSystemRollingUpdatesCfgMap()
	if res.Requeue || err != nil {
		return res, err
	}

	return reconcile.Result{}, nil
}

func (u *Upgrade26_to_27) upgradeAMPImageStreams() (reconcile.Result, error) {
	// implement upgrade procedure by reconcile procedure
	baseReconciler := operator.NewBaseReconciler(u.client, u.apiClientReader, u.scheme, u.logger)
	baseLogicReconciler := operator.NewBaseLogicReconciler(baseReconciler)
	reconciler := operator.NewAMPImagesReconciler(operator.NewBaseAPIManagerLogicReconciler(baseLogicReconciler, u.cr))
	return reconciler.Reconcile()
}

func (u *Upgrade26_to_27) upgradeSystemRollingUpdatesCfgMap() (reconcile.Result, error) {
	cfgMapName := "system"
	cfgMapEntry := "rolling_updates.yml"
	apiAsProductKey := "api_as_product"

	systemCfgMap := &v1.ConfigMap{}
	err := u.client.Get(context.TODO(), types.NamespacedName{Name: cfgMapName, Namespace: u.cr.Namespace}, systemCfgMap)
	if err != nil {
		return reconcile.Result{}, err
	}

	rollingUpdatesTxt, ok := systemCfgMap.Data[cfgMapEntry]
	if !ok {
		return reconcile.Result{}, fmt.Errorf("Expected '%s' key in '%s' ConfigMap not found", cfgMapEntry, cfgMapName)
	}

	var rollingUpdatesYAML map[string]map[string]bool
	err = yaml.Unmarshal([]byte(rollingUpdatesTxt), &rollingUpdatesYAML)
	if err != nil {
		return reconcile.Result{}, err
	}

	productionKey := "production"
	productionRollingUpdatesYAML, ok := rollingUpdatesYAML[productionKey]
	if !ok {
		u.logger.Info(fmt.Sprintf("'%s' key not found in '%s' ConfigMap key's value\n", productionKey, cfgMapEntry))
	}

	var changed bool
	if enabled, ok := productionRollingUpdatesYAML[apiAsProductKey]; ok {
		if !enabled {
			productionRollingUpdatesYAML[apiAsProductKey] = true
			changed = true
		}
	} else {
		productionRollingUpdatesYAML[apiAsProductKey] = true
		changed = true
	}

	if changed {
		marshaledRollingUpdates, err := yaml.Marshal(rollingUpdatesYAML)
		if err != nil {
			return reconcile.Result{}, err
		}
		systemCfgMap.Data[cfgMapEntry] = string(marshaledRollingUpdates)
		err = u.client.Update(context.TODO(), systemCfgMap)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{Requeue: changed}, nil
}

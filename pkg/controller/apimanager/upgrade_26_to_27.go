package apimanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	imagev1 "github.com/openshift/api/image/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Upgrade26_to_27 struct {
	BaseUpgrade
}

func (u *Upgrade26_to_27) Upgrade() (reconcile.Result, error) {
	res, err := u.upgradeImageStreams()
	if res.Requeue || err != nil {
		return res, err
	}

	return reconcile.Result{}, nil
}

func (u *Upgrade26_to_27) upgradeImageStream(imageStream *imagev1.ImageStream, newImageURL string) (bool, error) {
	changed := false
	tagToRemove := "2.6"
	desiredTag := product.ThreescaleRelease

	// TODO there are three options to approach this:
	// 1 - Replace the name of the the old version-tag with the new one. This is
	//     the approach that has been implemented here
	// 2 - Delete the version-tag and create a new version tag
	// 3 - Add the new version tag
	imageStreamSpec := &imageStream.Spec
	for idx := range imageStreamSpec.Tags {
		tag := &imageStreamSpec.Tags[idx]
		if tag.Name == "latest" && tag.From.Name == tagToRemove {
			tag.From.Name = desiredTag
			changed = true
		}

		if tag.Name == desiredTag {
			if tag.From.Name != newImageURL {
				tag.From.Name = newImageURL
				changed = true
			}
			if strings.Contains(tag.Annotations["openshift.io/display-name"], tagToRemove) {
				tag.Annotations["openshift.io/display-name"] = strings.ReplaceAll(tag.Annotations["openshift.io/display-name"], tagToRemove, desiredTag)
				changed = true
			}
		}

		if tag.Name == tagToRemove {
			tag.Name = desiredTag
			tag.From.Name = newImageURL
			tag.Annotations["openshift.io/display-name"] = strings.ReplaceAll(tag.Annotations["openshift.io/display-name"], tagToRemove, desiredTag)
			changed = true
		}
	}

	if changed {
		err := u.client.Update(context.TODO(), imageStream)
		if err != nil {
			return changed, err
		}
	}

	return changed, nil
}

func (u *Upgrade26_to_27) upgradeAPIcastImageStream() (bool, error) {
	apicastImageStream := &imagev1.ImageStream{}
	err := u.client.Get(context.TODO(), types.NamespacedName{Name: "amp-apicast", Namespace: u.cr.Namespace}, apicastImageStream)
	if err != nil {
		return false, err
	}

	newImageURL := product.CurrentImageProvider().GetApicastImage()
	if u.cr.Spec.Apicast != nil && u.cr.Spec.Apicast.Image != nil {
		newImageURL = *u.cr.Spec.Apicast.Image
	}

	return u.upgradeImageStream(apicastImageStream, newImageURL)
}

func (u *Upgrade26_to_27) upgradeBackendImageStream() (bool, error) {
	backendImageStream := &imagev1.ImageStream{}
	err := u.client.Get(context.TODO(), types.NamespacedName{Name: "amp-backend", Namespace: u.cr.Namespace}, backendImageStream)
	if err != nil {
		return false, err
	}

	newImageURL := product.CurrentImageProvider().GetBackendImage()
	if u.cr.Spec.Backend != nil && u.cr.Spec.Backend.Image != nil {
		newImageURL = *u.cr.Spec.Backend.Image
	}

	return u.upgradeImageStream(backendImageStream, newImageURL)

}

func (u *Upgrade26_to_27) upgradeBackendRedisImageStream() (bool, error) {
	backendRedisImageStream := &imagev1.ImageStream{}
	err := u.client.Get(context.TODO(), types.NamespacedName{Name: "backend-redis", Namespace: u.cr.Namespace}, backendRedisImageStream)
	if err != nil {
		return false, err
	}

	newImageURL := product.CurrentImageProvider().GetBackendRedisImage()
	if u.cr.Spec.Backend != nil && u.cr.Spec.Backend.RedisImage != nil {
		newImageURL = *u.cr.Spec.Backend.RedisImage
	}

	return u.upgradeImageStream(backendRedisImageStream, newImageURL)
}

func (u *Upgrade26_to_27) upgradeSystemRedisImageStream() (bool, error) {
	systemRedisImageStream := &imagev1.ImageStream{}
	err := u.client.Get(context.TODO(), types.NamespacedName{Name: "system-redis", Namespace: u.cr.Namespace}, systemRedisImageStream)
	if err != nil {
		return false, err
	}

	newImageURL := product.CurrentImageProvider().GetSystemRedisImage()
	if u.cr.Spec.System != nil && u.cr.Spec.System.RedisImage != nil {
		newImageURL = *u.cr.Spec.System.RedisImage
	}

	return u.upgradeImageStream(systemRedisImageStream, newImageURL)
}

func (u *Upgrade26_to_27) upgradeSystemImageStream() (bool, error) {
	systemImageStream := &imagev1.ImageStream{}
	err := u.client.Get(context.TODO(), types.NamespacedName{Name: "amp-system", Namespace: u.cr.Namespace}, systemImageStream)
	if err != nil {
		return false, err
	}

	newImageURL := product.CurrentImageProvider().GetSystemImage()
	if u.cr.Spec.System != nil && u.cr.Spec.System.Image != nil {
		newImageURL = *u.cr.Spec.System.Image
	}

	return u.upgradeImageStream(systemImageStream, newImageURL)
}

func (u *Upgrade26_to_27) upgradeSystemMySQLImageStream() (bool, error) {
	systemMySQLImageStream := &imagev1.ImageStream{}
	err := u.client.Get(context.TODO(), types.NamespacedName{Name: "system-mysql", Namespace: u.cr.Namespace}, systemMySQLImageStream)
	if err != nil {
		return false, err
	}

	newImageURL := product.CurrentImageProvider().GetSystemMySQLImage()
	if u.cr.Spec.System != nil && u.cr.Spec.System.DatabaseSpec != nil &&
		u.cr.Spec.System.DatabaseSpec.MySQL != nil &&
		u.cr.Spec.System.DatabaseSpec.MySQL.Image != nil {
		newImageURL = *u.cr.Spec.System.DatabaseSpec.MySQL.Image
	}

	return u.upgradeImageStream(systemMySQLImageStream, newImageURL)
}

func (u *Upgrade26_to_27) upgradeSystemPostgreSQLImageStream() (bool, error) {
	systemPostgreSQLImageStream := &imagev1.ImageStream{}
	err := u.client.Get(context.TODO(), types.NamespacedName{Name: "system-postgresql", Namespace: u.cr.Namespace}, systemPostgreSQLImageStream)
	if err != nil {
		return false, err
	}

	newImageURL := product.CurrentImageProvider().GetSystemPostgreSQLImage()
	if u.cr.Spec.System != nil && u.cr.Spec.System.DatabaseSpec != nil &&
		u.cr.Spec.System.DatabaseSpec.PostgreSQL != nil &&
		u.cr.Spec.System.DatabaseSpec.PostgreSQL.Image != nil {
		newImageURL = *u.cr.Spec.System.DatabaseSpec.PostgreSQL.Image
	}

	return u.upgradeImageStream(systemPostgreSQLImageStream, newImageURL)
}

func (u *Upgrade26_to_27) upgradeSystemDatabaseImageStream() (bool, error) {
	if u.cr.Spec.System.DatabaseSpec.MySQL != nil {
		return u.upgradeSystemMySQLImageStream()
	}

	if u.cr.Spec.System.DatabaseSpec.PostgreSQL != nil {
		return u.upgradeSystemPostgreSQLImageStream()
	}

	return false, fmt.Errorf("System database is not set")
}

func (u *Upgrade26_to_27) upgradeSystemMemcachedImageStream() (bool, error) {
	systemMemcachedImageStream := &imagev1.ImageStream{}
	err := u.client.Get(context.TODO(), types.NamespacedName{Name: "system-memcached", Namespace: u.cr.Namespace}, systemMemcachedImageStream)
	if err != nil {
		return false, err
	}

	newImageURL := product.CurrentImageProvider().GetSystemMemcachedImage()
	if u.cr.Spec.System != nil && u.cr.Spec.System.MemcachedImage != nil {
		newImageURL = *u.cr.Spec.System.MemcachedImage
	}

	return u.upgradeImageStream(systemMemcachedImageStream, newImageURL)
}

func (u *Upgrade26_to_27) upgradeZyncImageStream() (bool, error) {
	zyncImageStream := &imagev1.ImageStream{}
	err := u.client.Get(context.TODO(), types.NamespacedName{Name: "amp-zync", Namespace: u.cr.Namespace}, zyncImageStream)
	if err != nil {
		return false, err
	}

	newImageURL := product.CurrentImageProvider().GetZyncImage()
	if u.cr.Spec.Zync != nil && u.cr.Spec.Zync.Image != nil {
		newImageURL = *u.cr.Spec.Zync.Image
	}

	return u.upgradeImageStream(zyncImageStream, newImageURL)
}

func (u *Upgrade26_to_27) upgradeZyncDatabaseImageStream() (bool, error) {
	zyncDatabaseImageStream := &imagev1.ImageStream{}
	err := u.client.Get(context.TODO(), types.NamespacedName{Name: "zync-database-postgresql", Namespace: u.cr.Namespace}, zyncDatabaseImageStream)
	if err != nil {
		return false, err
	}

	newImageURL := product.CurrentImageProvider().GetZyncPostgreSQLImage()
	if u.cr.Spec.Zync != nil && u.cr.Spec.Zync.PostgreSQLImage != nil {
		newImageURL = *u.cr.Spec.Zync.PostgreSQLImage
	}

	return u.upgradeImageStream(zyncDatabaseImageStream, newImageURL)
}

func (u *Upgrade26_to_27) highAvailabilityModeEnabled() bool {
	return u.cr.Spec.HighAvailability != nil && u.cr.Spec.HighAvailability.Enabled
}

func (u *Upgrade26_to_27) upgradeImageStreams() (reconcile.Result, error) {
	anImageStreamChanged := false
	changed, err := u.upgradeAPIcastImageStream()
	if err != nil {
		return reconcile.Result{}, err
	}
	if changed {
		u.logger.Info("APIcast ImageStream upgraded")
		anImageStreamChanged = true
	}

	changed, err = u.upgradeBackendImageStream()
	if err != nil {
		return reconcile.Result{}, err
	}
	if changed {
		u.logger.Info("Backend ImageStream upgraded")
		anImageStreamChanged = true
	}

	changed, err = u.upgradeSystemImageStream()
	if err != nil {
		return reconcile.Result{}, err
	}
	if changed {
		u.logger.Info("System ImageStream upgraded")
		anImageStreamChanged = true
	}

	if !u.highAvailabilityModeEnabled() {
		changed, err = u.upgradeBackendRedisImageStream()
		if err != nil {
			return reconcile.Result{}, err
		}
		if changed {
			u.logger.Info("Backend Redis ImageStream upgraded")
			anImageStreamChanged = true
		}

		changed, err = u.upgradeSystemRedisImageStream()
		if err != nil {
			return reconcile.Result{}, err
		}
		if changed {
			u.logger.Info("System Redis ImageStream upgraded")
			anImageStreamChanged = true
		}

		changed, err = u.upgradeSystemDatabaseImageStream()
		if err != nil {
			return reconcile.Result{}, err
		}
		if changed {
			u.logger.Info("System Database ImageStream upgraded")
			anImageStreamChanged = true
		}
	}

	changed, err = u.upgradeSystemMemcachedImageStream()
	if err != nil {
		return reconcile.Result{}, err
	}
	if changed {
		u.logger.Info("System Memcached ImageStream upgraded")
		anImageStreamChanged = true
	}

	changed, err = u.upgradeZyncImageStream()
	if err != nil {
		return reconcile.Result{}, err
	}
	if changed {
		u.logger.Info("Zync ImageStream upgraded")
		anImageStreamChanged = true
	}

	changed, err = u.upgradeZyncDatabaseImageStream()
	if err != nil {
		return reconcile.Result{}, err
	}
	if changed {
		u.logger.Info("Zync Database ImageStream upgraded")
		anImageStreamChanged = true
	}

	return reconcile.Result{Requeue: anImageStreamChanged}, nil
}

package upgrader

import (
	"context"
	"fmt"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/operator"
	"github.com/3scale/3scale-operator/pkg/3scale/amp/product"
	"github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	imagev1 "github.com/openshift/api/image/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func UpgradeImages(cl client.Client, ns string) error {
	err := upgradeAmpImages(cl, ns)
	if err != nil {
		return err
	}

	err = upgradeBackendRedis(cl, ns)
	if err != nil {
		return err
	}

	err = upgradeSystemRedis(cl, ns)
	if err != nil {
		return err
	}

	err = upgradeSystemMysql(cl, ns)
	if err != nil {
		return err
	}

	err = upgradeSystemPostgresql(cl, ns)
	if err != nil {
		return err
	}

	return nil
}

func upgradeSystemMysql(cl client.Client, ns string) error {
	mysqlImage, err := getMysqlImage()
	if err != nil {
		return err
	}

	desired := mysqlImage.ImageStream()

	existing := &imagev1.ImageStream{}
	err = cl.Get(
		context.TODO(),
		types.NamespacedName{Name: desired.Name, Namespace: ns},
		existing)
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	} else if err != nil && k8serrors.IsNotFound(err) {
		// imagestream not deployed, no need to upgrade
		return nil
	}

	reconciler := operator.NewImageStreamGenericReconciler()
	if reconciler.IsUpdateNeeded(desired, existing) {
		err := cl.Update(context.TODO(), existing)
		if err != nil {
			return err
		}
		fmt.Printf("Update object %s\n", operator.ObjectInfo(existing))
	}
	return nil
}

func upgradeSystemPostgresql(cl client.Client, ns string) error {
	postgresqlImage, err := getPostgresqlImage()
	if err != nil {
		return err
	}

	desired := postgresqlImage.ImageStream()

	existing := &imagev1.ImageStream{}
	err = cl.Get(
		context.TODO(),
		types.NamespacedName{Name: desired.Name, Namespace: ns},
		existing)
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	} else if err != nil && k8serrors.IsNotFound(err) {
		// imagestream not deployed, no need to upgrade
		return nil
	}

	reconciler := operator.NewImageStreamGenericReconciler()
	if reconciler.IsUpdateNeeded(desired, existing) {
		err := cl.Update(context.TODO(), existing)
		if err != nil {
			return err
		}
		fmt.Printf("Update object %s\n", operator.ObjectInfo(existing))
	}
	return nil
}

func upgradeSystemRedis(cl client.Client, ns string) error {
	redis, err := getRedisComponent()
	if err != nil {
		return err
	}

	desired := redis.SystemImageStream()

	existing := &imagev1.ImageStream{}
	err = cl.Get(
		context.TODO(),
		types.NamespacedName{Name: desired.Name, Namespace: ns},
		existing)
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	} else if err != nil && k8serrors.IsNotFound(err) {
		// imagestream not deployed, no need to upgrade
		return nil
	}

	reconciler := operator.NewImageStreamGenericReconciler()
	if reconciler.IsUpdateNeeded(desired, existing) {
		err := cl.Update(context.TODO(), existing)
		if err != nil {
			return err
		}
		fmt.Printf("Update object %s\n", operator.ObjectInfo(existing))
	}
	return nil
}

func upgradeBackendRedis(cl client.Client, ns string) error {
	redis, err := getRedisComponent()
	if err != nil {
		return err
	}

	desired := redis.BackendImageStream()

	existing := &imagev1.ImageStream{}
	err = cl.Get(
		context.TODO(),
		types.NamespacedName{Name: desired.Name, Namespace: ns},
		existing)
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	} else if err != nil && k8serrors.IsNotFound(err) {
		// imagestream not deployed, no need to upgrade
		return nil
	}

	reconciler := operator.NewImageStreamGenericReconciler()
	if reconciler.IsUpdateNeeded(desired, existing) {
		err := cl.Update(context.TODO(), existing)
		if err != nil {
			return err
		}
		fmt.Printf("Update object %s\n", operator.ObjectInfo(existing))
	}
	return nil
}

func upgradeAmpImages(cl client.Client, ns string) error {
	ampImages, err := getAMPImagesComponent()
	if err != nil {
		return err
	}

	imagestreams := []*imagev1.ImageStream{
		ampImages.SystemImageStream(),
		ampImages.APICastImageStream(),
		ampImages.BackendImageStream(),
		ampImages.ZyncImageStream(),
		ampImages.SystemMemcachedImageStream(),
		ampImages.ZyncDatabasePostgreSQLImageStream(),
	}
	reconciler := operator.NewImageStreamGenericReconciler()

	for _, desired := range imagestreams {
		existing := &imagev1.ImageStream{}
		err := cl.Get(
			context.TODO(),
			types.NamespacedName{Name: desired.Name, Namespace: ns},
			existing)
		if err != nil {
			return err
		}

		if reconciler.IsUpdateNeeded(desired, existing) {
			err := cl.Update(context.TODO(), existing)
			if err != nil {
				return err
			}
			fmt.Printf("Update object %s\n", operator.ObjectInfo(existing))
		}
	}
	return nil
}

func getAMPImagesComponent() (*component.AmpImages, error) {
	optProv := component.AmpImagesOptionsBuilder{}
	// should be read from installed 3scale?
	optProv.AppLabel(appsv1alpha1.DefaultAppLabel)
	optProv.AMPRelease(product.ThreescaleRelease)
	optProv.ApicastImage(component.ApicastImageURL())
	optProv.BackendImage(component.BackendImageURL())
	optProv.SystemImage(component.SystemImageURL())
	optProv.ZyncImage(component.ZyncImageURL())
	optProv.ZyncDatabasePostgreSQLImage(component.ZyncPostgreSQLImageURL())
	optProv.SystemMemcachedImage(component.SystemMemcachedImageURL())
	// should be read from installed 3scale?
	optProv.InsecureImportPolicy(v1alpha1.DefaultImageStreamImportInsecure)

	otions, err := optProv.Build()
	if err != nil {
		return nil, err
	}
	return component.NewAmpImages(otions), nil
}

func getRedisComponent() (*component.Redis, error) {
	optProv := component.RedisOptionsBuilder{}
	// should be read from installed 3scale?
	optProv.AppLabel(appsv1alpha1.DefaultAppLabel)
	optProv.AMPRelease(product.ThreescaleRelease)
	optProv.BackendImage(component.BackendRedisImageURL())
	optProv.SystemImage(component.SystemRedisImageURL())
	// should be read from installed 3scale?
	optProv.InsecureImportPolicy(v1alpha1.DefaultImageStreamImportInsecure)
	// resource requirements only required for deployment config, not required for image change

	options, err := optProv.Build()
	if err != nil {
		return nil, err
	}
	return component.NewRedis(options), nil
}

func getPostgresqlImage() (*component.SystemPostgreSQLImage, error) {
	optProv := component.SystemPostgreSQLImageOptionsBuilder{}
	// should be read from installed 3scale?
	optProv.AppLabel(appsv1alpha1.DefaultAppLabel)
	optProv.AmpRelease(product.ThreescaleRelease)
	optProv.Image(component.SystemPostgreSQLImageURL())
	// should be read from installed 3scale?
	optProv.InsecureImportPolicy(v1alpha1.DefaultImageStreamImportInsecure)

	options, err := optProv.Build()
	if err != nil {
		return nil, err
	}

	return component.NewSystemPostgreSQLImage(options), nil
}

func getMysqlImage() (*component.SystemMySQLImage, error) {
	optProv := component.SystemMySQLImageOptionsBuilder{}
	// should be read from installed 3scale?
	optProv.AppLabel(appsv1alpha1.DefaultAppLabel)
	optProv.AmpRelease(product.ThreescaleRelease)
	optProv.Image(component.SystemPostgreSQLImageURL())
	// should be read from installed 3scale?
	optProv.InsecureImportPolicy(v1alpha1.DefaultImageStreamImportInsecure)

	options, err := optProv.Build()
	if err != nil {
		return nil, err
	}

	return component.NewSystemMySQLImage(options), nil
}

package backup

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/3scale/3scale-operator/pkg/3scale/amp/component"
	appsv1alpha1 "github.com/3scale/3scale-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/3scale-operator/pkg/helper"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type APIManagerBackupOptionsProvider struct {
	APIManagerBackupCR *appsv1alpha1.APIManagerBackup
	Client             client.Client
}

func NewAPIManagerBackupOptionsProvider(cr *appsv1alpha1.APIManagerBackup, client client.Client) *APIManagerBackupOptionsProvider {
	return &APIManagerBackupOptionsProvider{
		APIManagerBackupCR: cr,
		Client:             client,
	}
}

func (a *APIManagerBackupOptionsProvider) Options() (*APIManagerBackupOptions, error) {
	res := NewAPIManagerBackupOptions()
	res.APIManagerBackupName = a.APIManagerBackupCR.Name
	res.Namespace = a.APIManagerBackupCR.Namespace

	// Should we rely on always having the APIManager existing before doing something?
	// In restores for example it is desirable to not mandate it at all times so it
	// won't be able to properly obtained at option retrieval time. We'll only be able
	// to use the name and Get it when appropriate
	apiManager, err := a.apiManager()
	if err != nil {
		return nil, err
	}
	res.APIManager = apiManager
	res.APIManagerName = apiManager.Name

	pvcOptions, err := a.pvcBackupOptions()
	if err != nil {
		return nil, err
	}

	s3Options, err := a.s3BackupOptions()
	if err != nil {
		return nil, err
	}

	// TODO can this checks be omitted and just rely on the validator package in the APIManagerBackup struct?
	if pvcOptions == nil && s3Options == nil {
		return nil, fmt.Errorf("At least one backup destination has to be specified")
	}

	if pvcOptions != nil && s3Options != nil {
		return nil, fmt.Errorf("Only one backup destination can be specified")
	}

	res.APIManagerBackupPVCOptions = pvcOptions
	res.APIManagerBackupS3Options = s3Options

	return res, res.Validate()
}

func (a *APIManagerBackupOptionsProvider) pvcBackupOptions() (*APIManagerBackupPVCOptions, error) {
	if a.APIManagerBackupCR.Spec.BackupSource.PersistentVolumeClaim == nil {
		return nil, nil
	}

	res := NewAPIManagerBackupPVCOptions()
	res.BackupDestinationPVC.Name = fmt.Sprintf("apimanager-backup-%s", a.APIManagerBackupCR.Name)
	res.BackupDestinationPVC.StorageClass = a.APIManagerBackupCR.Spec.BackupSource.PersistentVolumeClaim.StorageClass
	res.BackupDestinationPVC.VolumeName = a.APIManagerBackupCR.Spec.BackupSource.PersistentVolumeClaim.VolumeName
	if a.APIManagerBackupCR.Spec.BackupSource.PersistentVolumeClaim.Resources != nil {
		res.BackupDestinationPVC.StorageRequests = &a.APIManagerBackupCR.Spec.BackupSource.PersistentVolumeClaim.Resources.Requests
	}

	return res, res.Validate()
}

func (a *APIManagerBackupOptionsProvider) s3BackupOptions() (*APIManagerBackupS3Options, error) {
	if a.APIManagerBackupCR.Spec.BackupSource.SimpleStorageService == nil {
		return nil, nil
	}

	res := NewAPIManagerBackupS3Options()

	s3BackupCredentials, err := a.s3BackupDestinationCredentials()
	if err != nil {
		return nil, err
	}
	res.BackupDestinationS3.Credentials = s3BackupCredentials

	res.BackupDestinationS3.Bucket = a.APIManagerBackupCR.Spec.BackupSource.SimpleStorageService.Bucket
	res.BackupDestinationS3.Region = a.APIManagerBackupCR.Spec.BackupSource.SimpleStorageService.Region
	res.BackupDestinationS3.Endpoint = a.APIManagerBackupCR.Spec.BackupSource.SimpleStorageService.Endpoint
	res.BackupDestinationS3.Path = a.APIManagerBackupCR.Spec.BackupSource.SimpleStorageService.Path
	res.BackupDestinationS3.ForcePathStyle = a.APIManagerBackupCR.Spec.BackupSource.SimpleStorageService.ForcePathStyle

	return res, res.Validate()
}

func (a *APIManagerBackupOptionsProvider) autodiscoveredAPIManager() (*appsv1alpha1.APIManager, error) {
	resList := &appsv1alpha1.APIManagerList{}
	err := a.Client.List(context.TODO(), resList, client.InNamespace(a.APIManagerBackupCR.Namespace))
	if err != nil {
		return nil, err
	}

	var res *appsv1alpha1.APIManager
	if len(resList.Items) == 0 {
		return nil, fmt.Errorf("No APIManagers found in namespace '%s'", a.APIManagerBackupCR.Namespace)
	}
	if len(resList.Items) > 1 {
		return nil, fmt.Errorf("Multiple APIManagers found in namespace '%s'. Unsupported scenario", a.APIManagerBackupCR.Namespace)
	}

	res = &resList.Items[0]
	return res, nil

}

func (a *APIManagerBackupOptionsProvider) apiManager() (*appsv1alpha1.APIManager, error) {
	var apiManager *appsv1alpha1.APIManager
	var err error
	if a.APIManagerBackupCR.Spec.APIManagerName != nil {
		apiManager, err = a.apiManagerFromName(*a.APIManagerBackupCR.Spec.APIManagerName)
		return apiManager, err
	}
	apiManager, err = a.autodiscoveredAPIManager()
	return apiManager, err
}

func (a *APIManagerBackupOptionsProvider) apiManagerFromName(name string) (*appsv1alpha1.APIManager, error) {
	res := &appsv1alpha1.APIManager{}
	err := a.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: a.APIManagerBackupCR.Namespace}, res)
	return res, err
}

func (a *APIManagerBackupOptionsProvider) s3BackupDestinationCredentials() (BackupDestinationS3Credentials, error) {
	res := BackupDestinationS3Credentials{}

	s3CredsSecretName := a.APIManagerBackupCR.Spec.BackupSource.SimpleStorageService.CredentialsSecretRef.Name
	s3CredsSecretNamespace := a.APIManagerBackupCR.Namespace
	if s3CredsSecretName == "" {
		return res, fmt.Errorf("Field 'Name' not specified for S3 backup destination secret reference")
	}

	s3CredsNamespacedName := types.NamespacedName{
		Name:      s3CredsSecretName,
		Namespace: s3CredsSecretNamespace,
	}

	s3CredsSecret := &v1.Secret{}
	err := a.Client.Get(context.TODO(), s3CredsNamespacedName, s3CredsSecret)
	if err != nil {
		return res, err
	}

	secretStringData := helper.GetSecretStringDataFromData(s3CredsSecret.Data)
	accessKeyID, ok := secretStringData[BackupDestinationS3CredentialsSecretAccessKeyFieldName]
	if !ok {
		return res, fmt.Errorf("Required key '%s' not found in secret '%s'", BackupDestinationS3CredentialsSecretAccessKeyFieldName, s3CredsSecretName)
	}

	secretAcessKey, ok := secretStringData[BackupDestinationS3CredentialsSecretSecretAccessKeyFieldName]
	if !ok {
		return res, fmt.Errorf("Required key '%s' not found in secret '%s'", BackupDestinationS3CredentialsSecretSecretAccessKeyFieldName, s3CredsSecretName)
	}

	res.AccessKeyID = accessKeyID
	res.SecretAccessKey = secretAcessKey

	return res, nil
}

func (a *APIManagerBackupOptionsProvider) systemS3FileStorage(apimanager *appsv1alpha1.APIManager) (SystemS3FileStorage, error) {
	res := SystemS3FileStorage{}

	// TODO perform nil checks in to verify that S3 is indeed used in APIManager

	s3ConfigurationSecretName := apimanager.Spec.System.FileStorageSpec.S3.ConfigurationSecretRef.Name
	s3ConfigurationSecretNamespace := apimanager.Namespace

	if s3ConfigurationSecretName == "" {
		return res, fmt.Errorf("Field 'Name' not specified for S3 System FileStorage secret reference")
	}

	s3ConfigurationSecretNamespacedName := types.NamespacedName{
		Name:      s3ConfigurationSecretName,
		Namespace: s3ConfigurationSecretNamespace,
	}

	s3ConfigurationSecret := &v1.Secret{}
	err := a.Client.Get(context.TODO(), s3ConfigurationSecretNamespacedName, s3ConfigurationSecret)
	if err != nil {
		return res, err
	}

	secretStringData := helper.GetSecretStringDataFromData(s3ConfigurationSecret.Data)

	accessKeyID, ok := secretStringData[component.AwsAccessKeyID]
	if !ok {
		return res, fmt.Errorf("Required key '%s' not found in secret '%s'", component.AwsAccessKeyID, s3ConfigurationSecretName)
	}

	secretAccessKeyID, ok := secretStringData[component.AwsSecretAccessKey]
	if !ok {
		return res, fmt.Errorf("Required key '%s' not found in secret '%s'", component.AwsSecretAccessKey, s3ConfigurationSecretName)
	}

	bucket, ok := secretStringData[component.AwsBucket]
	if !ok {
		return res, fmt.Errorf("Required key '%s' not found in secret '%s'", component.AwsBucket, s3ConfigurationSecretName)
	}

	region, ok := secretStringData[component.AwsRegion]
	if !ok {
		return res, fmt.Errorf("Required key '%s' not found in secret '%s'", component.AwsRegion, s3ConfigurationSecretName)
	}

	forcePathStyle, ok := secretStringData[component.AwsPathStyle]
	if ok && forcePathStyle != "" {
		boolVal, err := strconv.ParseBool(forcePathStyle)
		if err != nil {
			return res, err
		}

		res.ForcePathStyle = &boolVal
	}

	protocol, protocolOK := secretStringData[component.AwsProtocol]
	var disableSSL *bool
	if protocolOK && protocol == strings.ToLower(string(v1.URISchemeHTTP)) {
		tmpDisableSSL := true
		res.DisableSSL = &tmpDisableSSL
	}

	var endpoint *string
	hostname, ok := secretStringData[component.AwsHostname]
	if ok && hostname != "" {
		endpointTmp := hostname
		if protocolOK && protocol != "" {
			endpointTmp = fmt.Sprintf("%s://%s", protocol, endpointTmp)
		}
		url, err := url.Parse(endpointTmp)
		if err != nil {
			return res, err
		}
		tmpURLString := url.String()
		endpoint = &tmpURLString
	}

	res.Credentials.AccessKeyID = accessKeyID
	res.Credentials.SecretAccessKey = secretAccessKeyID
	res.Bucket = bucket
	res.Region = region
	res.Endpoint = endpoint
	res.DisableSSL = disableSSL

	return res, nil

}

// TODO decide where to instantiate aws client with config
// func (b *APIManagerBackupBuilder) awsConfig(backupS3Credentials *BackupS3Credentials) *nativeaws.Config {
// 	awsConfig := &nativeaws.Config{}

// 	credentials := nativeawscredentials.NewStaticCredentials(backupS3Credentials.AccessKeyID, backupS3Credentials.SecretAccessKey, "")
// 	awsConfig = awsConfig.WithCredentials(credentials)

// 	if b.cr.Spec.BackupSource.SimpleStorageService.Region != "" {
// 		awsConfig = awsConfig.WithRegion(b.cr.Spec.BackupSource.SimpleStorageService.Region)
// 	}

// 	if b.cr.Spec.BackupSource.SimpleStorageService.Endpoint != "" {
// 		// It seems AWS accepts endpoint as a hostname only or fully qualified URI
// 		awsConfig = awsConfig.WithEndpoint(b.cr.Spec.BackupSource.SimpleStorageService.Endpoint)
// 	}

// 	if b.cr.Spec.BackupSource.SimpleStorageService.ForcePathStyle {
// 		awsConfig = awsConfig.WithS3ForcePathStyle(b.cr.Spec.BackupSource.SimpleStorageService.ForcePathStyle)
// 	}

// 	// TODO add disable SSL based on some condition??? parse it from the endpoint? maybe when setting endpoint itself, if it is a fully qualified
// 	// URI it is able to set the SSL disablement automatically?
// 	// For what it has been seen in the endpoints package from aws, disableSSL only has effect when it is set to true
// 	// when the endpoint URI does NOT contain the scheme part. In this case we can decide whether to provide the DisableSSL scheme or not as part
// 	// of configurable field. If we decide to do so we should document that Endpoint has preference if set in that case
// 	return awsConfig
// }

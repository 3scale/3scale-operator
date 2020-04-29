package backup

// type S3BackupSource struct {
// 	BackupS3Credentials        BackupS3Credentials
// 	Endpoint                   string
// 	ForcePathStyle             bool
// 	Bucket                     string
// 	Region                     string
// 	Path                       string
// 	S3Client                   aws.S3Client
// 	KubernetesObjectsNamespace string
// 	Client                     client.Client
// }

// func (b *S3BackupSource) Validate() error {
// 	return nil
// }

// func (b *S3BackupSource) backupKubernetesObjects() error {
// 	err := b.backupSecrets()
// 	if err != nil {
// 		return err
// 	}

// 	err = b.backupConfigMaps()
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (b *S3BackupSource) cleanObjectMeta(original *metav1.ObjectMeta) metav1.ObjectMeta {
// 	return metav1.ObjectMeta{
// 		Name:        original.Name,
// 		Labels:      original.Labels,
// 		Annotations: original.Annotations,
// 		// backup Some other fields?
// 	}
// }

// func (b *S3BackupSource) cleanSecret(originalSecret *v1.Secret) *v1.Secret {
// 	newSecret := originalSecret.DeepCopy()
// 	newSecret.ObjectMeta = b.cleanObjectMeta(&newSecret.ObjectMeta)
// 	return newSecret
// }

// func (b *S3BackupSource) cleanConfigMap(originalConfigMap *v1.ConfigMap) *v1.ConfigMap {
// 	newConfigMap := originalConfigMap.DeepCopy()
// 	newConfigMap.ObjectMeta = b.cleanObjectMeta(&newConfigMap.ObjectMeta)
// 	return newConfigMap
// }

// func (b *S3BackupSource) backupSecrets() error {
// 	sortedSecretNames := helper.SortedMapStringStringValues(secretsToBackup)
// 	for _, v := range sortedSecretNames {
// 		err := b.backupSecret(v)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func (b *S3BackupSource) backupConfigMaps() error {
// 	sortedConfigMaps := helper.SortedMapStringStringValues(configMapsToBackup)
// 	for _, v := range sortedConfigMaps {
// 		err := b.backupConfigMap(v)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// // TODO try to refactor method to reuse between backupConfigMap and backupSecret
// func (b *S3BackupSource) backupConfigMap(name string) error {
// 	objectKey := fmt.Sprintf("%s.yaml", name)
// 	// S3 API is eventually consistent. How do we make sure that we will obtain
// 	// a Get after having done a PUT?? if we come into that case what will happen
// 	//
// 	err := b.objectExists(objectKey)
// 	if err != nil {
// 		return err
// 	}

// 	configMap := &v1.ConfigMap{}
// 	err = b.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: b.KubernetesObjectsNamespace}, configMap)
// 	if err != nil {
// 		return err
// 	}
// 	configMapToSerialize := b.cleanConfigMap(configMap)

// 	res, err := helper.MarshalObjectToYAML(configMapToSerialize)
// 	if err != nil {
// 		return err
// 	}
// 	err = b.PutObject(objectKey, res)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (b *S3BackupSource) backupSecret(secretName string) error {
// 	objectKey := fmt.Sprintf("%s.yaml", secretName)
// 	// S3 API is eventually consistent. How do we make sure that we will obtain
// 	// a Get after having done a PUT?? if we come into that case what will happen
// 	//
// 	err := b.objectExists(objectKey)
// 	if err != nil {
// 		return err
// 	}

// 	secret := &v1.Secret{}
// 	err = b.Client.Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: b.KubernetesObjectsNamespace}, secret)
// 	if err != nil {
// 		return err
// 	}
// 	secretToSerialize := b.cleanSecret(secret)

// 	res, err := helper.MarshalObjectToYAML(secretToSerialize)
// 	if err != nil {
// 		return err
// 	}
// 	err = b.PutObject(objectKey, res)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

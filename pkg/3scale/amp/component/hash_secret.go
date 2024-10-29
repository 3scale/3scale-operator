package component

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"

	"github.com/3scale/3scale-operator/pkg/helper"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	HashedSecretName = "hashed-secret-data"
)

func HashedSecret(ctx context.Context, k8sclient client.Client, secretRefs []*v1.LocalObjectReference, ns string, hashSecretLabels map[string]string) (*v1.Secret, error) {
	hashedSecretData, err := computeHashedSecretData(ctx, k8sclient, secretRefs, ns)
	if err != nil {
		return nil, err
	}

	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      HashedSecretName,
			Namespace: ns,
			Labels:    hashSecretLabels,
		},
		StringData: hashedSecretData,
		Type:       v1.SecretTypeOpaque,
	}, nil
}

func computeHashedSecretData(ctx context.Context, k8sclient client.Client, secretRefs []*v1.LocalObjectReference, ns string) (map[string]string, error) {
	data := make(map[string]string)

	for _, secretRef := range secretRefs {
		secret := &v1.Secret{}
		key := client.ObjectKey{
			Name:      secretRef.Name,
			Namespace: ns,
		}
		err := k8sclient.Get(ctx, key, secret)
		if err != nil {
			return nil, err
		}

		if helper.IsSecretWatchedBy3scale(secret) {
			data[secretRef.Name] = HashSecret(secret.Data)
		}
	}

	return data, nil
}

func HashSecret(data map[string][]byte) string {
	hash := sha256.New()

	sortedKeys := make([]string, 0, len(data))
	for k := range data {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for _, key := range sortedKeys {
		value := data[key]
		combinedKeyValue := append([]byte(key), value...)
		hash.Write(combinedKeyValue)
	}

	hashBytes := hash.Sum(nil)

	return hex.EncodeToString(hashBytes)
}

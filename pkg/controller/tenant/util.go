package tenant

import (
	"net/url"

	apiv1alpha1 "github.com/3scale/3scale-operator/apis/capabilities/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// addOwnerRefToObject appends the desired OwnerReference to the object
func addOwnerRefToObject(o metav1.Object, r metav1.OwnerReference) {
	o.SetOwnerReferences(append(o.GetOwnerReferences(), r))
}

// asOwner returns an owner reference set as the tenant CR
func asOwner(t *apiv1alpha1.Tenant) metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion: apiv1alpha1.GroupVersion.String(),
		Kind:       apiv1alpha1.TenantKind,
		Name:       t.Name,
		UID:        t.UID,
		Controller: &trueVar,
	}
}

func URLFromDomain(domain string) (*url.URL, error) {
	u, err := url.Parse(domain)
	if err != nil {
		return nil, err
	}
	u.Scheme = "https"
	return u, nil
}

package read

import (
	"context"
	"github.com/RHsyseng/operator-utils/pkg/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	clientv1 "sigs.k8s.io/controller-runtime/pkg/client"
)

type resourceReader struct {
	reader      clientv1.Reader
	namespace   string
	ownerObject metav1.Object
}

// New creates a resourceReader object that can be used to load/list kubernetes resources
// the provided reader object will be used for the underlying operations
func New(reader clientv1.Reader) *resourceReader {
	return &resourceReader{reader: reader}
}

// WithNamespace filters list operations to the provided namespace
func (this *resourceReader) WithNamespace(namespace string) *resourceReader {
	this.namespace = namespace
	return this
}

// WithOwnerObject filters list operations to items that have ownerObject as an owner reference
func (this *resourceReader) WithOwnerObject(ownerObject metav1.Object) *resourceReader {
	this.ownerObject = ownerObject
	return this
}

// List returns a list of Kubernetes resources based on provided List object and configuration
// any error from underlying calls is directly returned as well
func (this *resourceReader) List(listObject runtime.Object) ([]resource.KubernetesResource, error) {
	var resources []resource.KubernetesResource
	err := this.reader.List(context.TODO(), listObject, &clientv1.ListOptions{Namespace: this.namespace})
	if err != nil {
		return nil, err
	}
	itemsValue := reflect.Indirect(reflect.ValueOf(listObject)).FieldByName("Items")
	for index := 0; index < itemsValue.Len(); index++ {
		item := addr(itemsValue.Index(index)).Interface().(resource.KubernetesResource)
		if this.ownerObject == nil || isOwner(this.ownerObject, item) {
			resources = append(resources, item)
		}
	}
	return resources, nil
}

func addr(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		return v
	}

	return v.Addr()
}

func isOwner(owner metav1.Object, res resource.KubernetesResource) bool {
	for _, ownerRef := range res.GetOwnerReferences() {
		if ownerRef.UID == owner.GetUID() {
			return true
		}
	}
	return false
}

// ListAll returns a map of Kubernetes resources organized by type, based on provided List objects and configuration
// any error from underlying calls is directly returned as well
func (this *resourceReader) ListAll(listObjects ...runtime.Object) (map[reflect.Type][]resource.KubernetesResource, error) {
	objectMap := make(map[reflect.Type][]resource.KubernetesResource)
	for _, listObject := range listObjects {
		resources, err := this.List(listObject)
		if err != nil {
			return nil, err
		}
		if len(resources) > 0 {
			itemType := reflect.ValueOf(resources[0]).Elem().Type()
			objectMap[itemType] = resources
		}
	}
	return objectMap, nil
}

// Load returns an object of the specified type with the given name, in the previously configured namespace
// any error from the underlying call, including a not-found error, is directly returned as well
func (this *resourceReader) Load(resourceType reflect.Type, name string) (resource.KubernetesResource, error) {
	deployed := reflect.New(resourceType).Interface().(resource.KubernetesResource)
	err := this.reader.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: this.namespace}, deployed)
	return deployed, err
}

package reconcilers

import (
	"testing"

	imagev1 "github.com/openshift/api/image/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGenericImageStreamMutator(t *testing.T) {
	existing := &imagev1.ImageStream{
		TypeMeta:   metav1.TypeMeta{APIVersion: "image.openshift.io/v1", Kind: "ImageStream"},
		ObjectMeta: metav1.ObjectMeta{Name: "myIS", Namespace: "MyNS"},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				imagev1.TagReference{
					Name:         "tag0",
					ImportPolicy: imagev1.TagImportPolicy{Insecure: false, Scheduled: true},
					From:         &v1.ObjectReference{Kind: "ImageStreamTag", Name: "3scale-0"},
				},
				imagev1.TagReference{
					Name:         "tag1",
					ImportPolicy: imagev1.TagImportPolicy{Insecure: false, Scheduled: true},
					From:         &v1.ObjectReference{Kind: "ImageStreamTag", Name: "3scale-1"},
				},
			},
		},
	}

	desired := &imagev1.ImageStream{
		TypeMeta:   metav1.TypeMeta{APIVersion: "image.openshift.io/v1", Kind: "ImageStream"},
		ObjectMeta: metav1.ObjectMeta{Name: "myIS", Namespace: "MyNS"},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				imagev1.TagReference{
					// tag that should be updated
					Name:         "tag1",
					ImportPolicy: imagev1.TagImportPolicy{Insecure: true, Scheduled: false},
					From:         &v1.ObjectReference{Kind: "ImageStreamTag", Name: "3scale-1other"},
				},
				imagev1.TagReference{
					// tag that should be added
					Name:         "tag2",
					ImportPolicy: imagev1.TagImportPolicy{Insecure: false, Scheduled: true},
					From:         &v1.ObjectReference{Kind: "ImageStreamTag", Name: "3scale-2"},
				},
			},
		},
	}

	update, err := GenericImageStreamMutator(existing, desired)
	if err != nil {
		t.Fatal(err)
	}
	if !update {
		t.Fatal("when defaults can be applied, reconciler reported no update needed")
	}

	if len(existing.Spec.Tags) != 3 {
		t.Fatalf("reconciled obj does not have expected number of tags. Expected: 3, got: %d", len(existing.Spec.Tags))
	}

	findTagReference := func(tagRefName string, tagRefS []imagev1.TagReference) int {
		for i := range tagRefS {
			if tagRefS[i].Name == tagRefName {
				return i
			}
		}
		return -1
	}

	// tag0 existed previously in obj, should be left untouched
	tag0Index := findTagReference("tag0", existing.Spec.Tags)
	if tag0Index < 0 {
		t.Fatal("reconciled obj does not have tag0")
	}

	// tag1 existed previously in obj, should be updated
	tag1Index := findTagReference("tag1", existing.Spec.Tags)
	if tag1Index < 0 {
		t.Fatal("reconciled obj does not have tag1")
	}
	tag1 := existing.Spec.Tags[1]
	// From and ImportPolicy fields should have been reconciled
	if tag1.From.Name != "3scale-1other" {
		t.Fatal("reconciled obj tag1 'from' was not reconciled")
	}

	if !tag1.ImportPolicy.Insecure {
		t.Fatal("reconciled obj tag1 'impoortpolicy.insecure' was not reconciled")
	}

	if tag1.ImportPolicy.Scheduled {
		t.Fatal("reconciled obj tag1 'impoortpolicy.scheduled' was not reconciled")
	}

	// tag2 did not exist previously in obj, should be appended
	tag2Index := findTagReference("tag2", existing.Spec.Tags)
	if tag2Index < 0 {
		t.Fatal("reconciled obj does not have tag2")
	}
}

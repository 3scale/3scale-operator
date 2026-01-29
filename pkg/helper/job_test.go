package helper

import "testing"

// TestSystemAppRevisionAnnotationValue ensures the annotation key doesn't change
// unexpectedly, as existing Jobs in customer clusters depend on this value.
// If you need to change this value, you'll need a migration strategy for existing deployments.
func TestSystemAppRevisionAnnotationValue(t *testing.T) {
	expected := "apimanager.apps.3scale.net/system-app-deployment-revision"
	if SystemAppRevisionAnnotation != expected {
		t.Errorf("SystemAppRevisionAnnotation changed from %q to %q - this is a breaking change!",
			expected, SystemAppRevisionAnnotation)
	}
}

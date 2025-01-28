package helper

func ManagedByOperatorAnnotation() map[string]string {
	return map[string]string{
		"annotations[managed_by]": "operator",
	}
}

func ManagedByOperatorAnnotationExists(annotations map[string]string) bool {
	if val, exists := annotations["managed_by"]; exists && val == "operator" {
		return true
	}
	return false
}

func ManagedByOperatorDeveloperAccountAnnotation() map[string]string {
	return map[string]string{
		"managed_by": "operator",
	}
}

package helper

func Any(arr []bool) bool {
	result := false
	for _, v := range arr {
		result = result || v
	}
	return result
}

func All(arr []bool) bool {
	result := true
	for _, v := range arr {
		result = result && v
	}
	return result
}

package helper

func ArrayStringDifference(a, b []string) []string {
	target := map[string]bool{}
	for _, x := range b {
		target[x] = true
	}

	result := []string{}
	for _, x := range a {
		if _, ok := target[x]; !ok {
			result = append(result, x)
		}
	}

	return result
}

func ArrayStringIntersection(a, b []string) []string {
	target := map[string]bool{}
	for _, x := range b {
		target[x] = true
	}

	result := []string{}
	for _, x := range a {
		if _, ok := target[x]; ok {
			result = append(result, x)
		}
	}

	return result
}

// ArrayFind returns the smallest index i at which x == a[i],
// or len(a) if there is no such index.
func ArrayFind(a []string, x string) int {
	for i, n := range a {
		if x == n {
			return i
		}
	}
	return len(a)
}

// ArrayContains tells whether a contains x.
func ArrayContains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

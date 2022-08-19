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

// https://newbedev.com/check-for-equality-on-slices-without-order
func StringSliceEqualWithoutOrder(x, y []string) bool {
	if len(x) != len(y) {
		return false
	}
	// create a map of string -> int
	diff := make(map[string]int, len(x))
	for _, _x := range x {
		// 0 value for int is 0, so just increment a counter for the string
		diff[_x]++
	}
	for _, _y := range y {
		// If the string _y is not in diff bail out early
		if _, ok := diff[_y]; !ok {
			return false
		}
		diff[_y]--
		if diff[_y] == 0 {
			delete(diff, _y)
		}
	}

	return len(diff) == 0
}

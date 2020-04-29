package helper

import "sort"

func SortedMapStringStringValues(input map[string]string) []string {
	var sortedValues []string
	for _, v := range input {
		sortedValues = append(sortedValues, v)
	}
	sort.Slice(sortedValues, func(i, j int) bool { return sortedValues[i] < sortedValues[j] })
	return sortedValues
}

func SortedMapStringStringKeys(input map[string]string) []string {
	var sortedKeys []string
	for k := range input {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Slice(sortedKeys, func(i, j int) bool { return sortedKeys[i] < sortedKeys[j] })
	return sortedKeys
}

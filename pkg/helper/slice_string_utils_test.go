package helper

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestArrayStringDifference(t *testing.T) {
	cases := []struct {
		name         string
		a            []string
		b            []string
		expectedDiff []string
	}{
		{"test01", []string{"A", "B", "C"}, []string{"A", "D", "E"}, []string{"B", "C"}},
		{"test02", []string{"A", "D", "E"}, []string{"A", "B", "C"}, []string{"D", "E"}},
		{"test03", []string{"A", "B", "C"}, []string{}, []string{"A", "B", "C"}},
		{"test04", []string{}, []string{"A", "B", "C"}, []string{}},
		{"test05", []string{}, []string{}, []string{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(subT *testing.T) {
			diff := ArrayStringDifference(tc.a, tc.b)
			if !reflect.DeepEqual(diff, tc.expectedDiff) {
				diff := cmp.Diff(diff, tc.expectedDiff)
				subT.Errorf("diff %s", diff)
			}
		})
	}
}

func TestArrayStringIntersection(t *testing.T) {
	cases := []struct {
		name         string
		a            []string
		b            []string
		expectedDiff []string
	}{
		{"test01", []string{"A", "B", "C"}, []string{"A", "D", "E"}, []string{"A"}},
		{"test02", []string{"A", "D", "E"}, []string{"A", "B", "C"}, []string{"A"}},
		{"test03", []string{"A", "B", "C"}, []string{}, []string{}},
		{"test04", []string{}, []string{"A", "B", "C"}, []string{}},
		{"test05", []string{}, []string{}, []string{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(subT *testing.T) {
			diff := ArrayStringIntersection(tc.a, tc.b)
			if !reflect.DeepEqual(diff, tc.expectedDiff) {
				diff := cmp.Diff(diff, tc.expectedDiff)
				subT.Errorf("diff %s", diff)
			}
		})
	}
}

func TestStringSliceEqualWithoutOrder(t *testing.T) {
	cases := []struct {
		name        string
		a           []string
		b           []string
		expectedRes bool
	}{
		{"test01", []string{"A", "B", "C"}, []string{"A", "D", "E"}, false},
		{"test02", []string{"A", "B", "C"}, []string{"C", "B", "A"}, true},
		{"test03", []string{"A", "B", "C"}, []string{}, false},
		{"test04", []string{}, []string{"A", "B", "C"}, false},
		{"test05", []string{}, []string{}, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(subT *testing.T) {
			res := StringSliceEqualWithoutOrder(tc.a, tc.b)
			if res != tc.expectedRes {
				subT.Errorf("Got %t, expected %t", res, tc.expectedRes)
			}
		})
	}
}

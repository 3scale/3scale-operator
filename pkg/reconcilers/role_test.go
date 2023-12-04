package reconcilers

import (
	"github.com/google/go-cmp/cmp"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
)

func TestRoleRuleMutator(t *testing.T) {

	testRules1 := []rbacv1.PolicyRule{
		{
			APIGroups: []string{"apps"},
			Resources: []string{
				"deployment",
			},
			Verbs: []string{
				"get",
				"list",
			},
		},
	}

	testRules2 := []rbacv1.PolicyRule{
		{
			APIGroups: []string{"apps"},
			Resources: []string{
				"deployment",
			},
			Verbs: []string{
				"get",
				"list",
				"create",
				"delete",
			},
		},
	}

	roleFactory := func(rules []rbacv1.PolicyRule) *rbacv1.Role {
		return &rbacv1.Role{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Role",
				APIVersion: "rbac.authorization.k8s.io/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "testRole",
				Namespace: "testNamespace",
			},
			Rules: rules,
		}
	}

	cases := []struct {
		testName       string
		existingRules  []rbacv1.PolicyRule
		desiredRules   []rbacv1.PolicyRule
		expectedResult bool
	}{
		{"NothingToReconcile", nil, nil, false},
		{"EqualRules", testRules1, testRules1, false},
		{"DifferentRules", testRules1, testRules2, true},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(subT *testing.T) {
			existing := roleFactory(tc.existingRules)
			desired := roleFactory(tc.desiredRules)
			update, err := RoleRuleMutator(desired, existing)
			if err != nil {
				subT.Fatal(err)
			}
			if update != tc.expectedResult {
				subT.Fatalf("result failed, expected: %t, got: %t", tc.expectedResult, update)
			}
			if !reflect.DeepEqual(existing.Rules, desired.Rules) {
				subT.Fatal(cmp.Diff(existing.Rules, desired.Rules))
			}
		})
	}
}

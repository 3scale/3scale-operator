package v1beta1

import (
	"strings"
	"testing"

	"github.com/go-logr/logr"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func getv1beta1TestLogger() logr.Logger {
	return logf.Log.WithName("v1beta1")
}

func defaultTestingProduct() Product {
	product := Product{
		Spec: ProductSpec{
			Name: "productA",
		},
	}

	product.SetDefaults(getv1beta1TestLogger())
	return product
}

func TestDefaultProductValid(t *testing.T) {
	product := defaultTestingProduct()

	errors := product.Validate()
	if len(errors) > 0 {
		t.Errorf("product default is invalid: %s", errors.ToAggregate().Error())
	}
}

func TestValidateProductHitsMetric(t *testing.T) {
	productName := "productA"

	product := Product{
		Spec: ProductSpec{
			Name: productName,
			Metrics: map[string]MetricSpec{
				"hits": MetricSpec{
					Name:        "Hits",
					Unit:        "hit",
					Description: "Number of API hits",
				},
			},
		},
	}

	errors := product.Validate()
	if len(errors) > 0 && strings.Contains(errors.ToAggregate().Error(), "'hits' metric must exist") {
		t.Error("product hits validation fails when hits exists")
	}
}

func TestValidateProductMappingRuleRefs(t *testing.T) {
	product := defaultTestingProduct()

	product.Spec.MappingRules = []MappingRuleSpec{
		{
			HTTPMethod:      "GET",
			Pattern:         "/pets",
			MetricMethodRef: "notExistingRef",
		},
	}

	errors := product.Validate()
	if len(errors) == 0 || !strings.Contains(errors.ToAggregate().Error(), "mappingrule does not have valid metric or method reference") {
		t.Error("product mappingrule validation fails when not existing refs exist")
	}
}

func TestValidateProductNotUniqueLimitPeriods(t *testing.T) {
	product := defaultTestingProduct()

	product.Spec.ApplicationPlans = map[string]ApplicationPlanSpec{
		"plan01": ApplicationPlanSpec{
			Limits: []LimitSpec{
				{
					Period: "year",
					Value:  23,
					MetricMethodRef: MetricMethodRefSpec{
						SystemName: "hits",
					},
				},
				{
					Period: "year",
					Value:  1,
					MetricMethodRef: MetricMethodRefSpec{
						SystemName: "hits",
					},
				},
			},
		},
	}

	errors := product.Validate()
	if len(errors) == 0 || !strings.Contains(errors.ToAggregate().Error(), "limit period is not unique") {
		t.Error("product plan validation fails when limit period is not unique")
	}
}

func TestValidateProductPlanLimitUnkonwnRef(t *testing.T) {
	product := defaultTestingProduct()

	product.Spec.ApplicationPlans = map[string]ApplicationPlanSpec{
		"plan01": ApplicationPlanSpec{
			Limits: []LimitSpec{
				{
					Period: "year",
					Value:  23,
					MetricMethodRef: MetricMethodRefSpec{
						SystemName: "unknownRef",
					},
				},
			},
		},
	}

	errors := product.Validate()
	if len(errors) == 0 || !strings.Contains(errors.ToAggregate().Error(), "limit does not have valid local metric or method reference.") {
		t.Error("valition passes and limit does not have valid local metric or method reference.")
	}
}

func TestValidateProductPlanPricingRuleUnkonwnRef(t *testing.T) {
	product := defaultTestingProduct()

	product.Spec.ApplicationPlans = map[string]ApplicationPlanSpec{
		"plan01": ApplicationPlanSpec{
			PricingRules: []PricingRuleSpec{
				{
					From: 0,
					To:   10,
					MetricMethodRef: MetricMethodRefSpec{
						SystemName: "unknownRef",
					},
				},
			},
		},
	}

	errors := product.Validate()
	if len(errors) == 0 || !strings.Contains(errors.ToAggregate().Error(), "Pricing rule does not have valid local metric or method reference.") {
		t.Error("valition passes and pricing rule does not have valid local metric or method reference.")
	}
}

func TestValidateProductPlanPricingRuleInvalidRange(t *testing.T) {
	product := defaultTestingProduct()

	product.Spec.ApplicationPlans = map[string]ApplicationPlanSpec{
		"plan01": ApplicationPlanSpec{
			PricingRules: []PricingRuleSpec{
				{
					From: 10,
					To:   0,
					MetricMethodRef: MetricMethodRefSpec{
						SystemName: "hits",
					},
				},
			},
		},
	}

	errors := product.Validate()
	if len(errors) == 0 || !strings.Contains(errors.ToAggregate().Error(), "'To' value cannot be less than your 'From' value.") {
		t.Error("valition passes and pricing rule does not have valid range.")
	}
}

func TestValidateProductPlanPricingRuleOverlappingRange(t *testing.T) {
	product := defaultTestingProduct()

	product.Spec.ApplicationPlans = map[string]ApplicationPlanSpec{
		"plan01": ApplicationPlanSpec{
			PricingRules: []PricingRuleSpec{
				{
					From: 15,
					To:   20,
					MetricMethodRef: MetricMethodRefSpec{
						SystemName: "hits",
					},
				},
				{
					From: 0,
					To:   10,
					MetricMethodRef: MetricMethodRefSpec{
						SystemName: "hits",
					},
				},
				{
					From: 10,
					To:   50,
					MetricMethodRef: MetricMethodRefSpec{
						SystemName: "hits",
					},
				},
			},
		},
	}

	errors := product.Validate()
	if len(errors) == 0 || !strings.Contains(errors.ToAggregate().Error(), "'From' value cannot be less than 'To' values of current rules") {
		t.Error("valition passes and pricing rule ranges overlap.")
	}
}

func TestValidateProductHappyPath(t *testing.T) {
	product := defaultTestingProduct()

	product.Spec.MappingRules = []MappingRuleSpec{
		{
			HTTPMethod:      "GET",
			Pattern:         "/pets",
			MetricMethodRef: "hits",
		},
	}

	product.Spec.ApplicationPlans = map[string]ApplicationPlanSpec{
		"plan01": ApplicationPlanSpec{
			Limits: []LimitSpec{
				{
					Period: "year",
					Value:  100000,
					MetricMethodRef: MetricMethodRefSpec{
						SystemName: "hits",
					},
				},
			},
			PricingRules: []PricingRuleSpec{
				{
					From: 0,
					To:   10,
					MetricMethodRef: MetricMethodRefSpec{
						SystemName: "hits",
					},
				},
				{
					From: 10,
					To:   20,
					MetricMethodRef: MetricMethodRefSpec{
						SystemName: "hits",
					},
				},
				{
					From: 20,
					To:   100,
					MetricMethodRef: MetricMethodRefSpec{
						SystemName: "hits",
					},
				},
			},
		},
	}

	errors := product.Validate()
	if len(errors) > 0 {
		t.Errorf("product validation fails: %s", errors.ToAggregate().Error())
	}
}

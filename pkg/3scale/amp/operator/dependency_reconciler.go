package operator

import (
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// DependencyReconciler is an object that reconciles a dependency for a component
type DependencyReconciler interface {
	Reconcile() (reconcile.Result, error)
}

// DependencyReconcilerConstructor is the standard function to instantiate
// a DependencyReconciler
type DependencyReconcilerConstructor func(*BaseAPIManagerLogicReconciler) DependencyReconciler

// CompositeDependecyReconcilerConstructor creates a single DependencyReconcilerConstructor
// from multiple ones that instantiates a CompositeDependencyReconciler with
// the instantiated DependencyReconcilers
func CompositeDependencyReconcilerConstructor(constructors ...DependencyReconcilerConstructor) DependencyReconcilerConstructor {
	return func(b *BaseAPIManagerLogicReconciler) DependencyReconciler {
		reconcilers := make([]DependencyReconciler, len(constructors))

		for i, constructor := range constructors {
			reconcilers[i] = constructor(b)
		}

		return &CompositeDependencyReconciler{
			Reconcilers: reconcilers,
		}
	}
}

type CompositeDependencyReconciler struct {
	Reconcilers []DependencyReconciler
}

var _ DependencyReconciler = &CompositeDependencyReconciler{}

func (r *CompositeDependencyReconciler) Reconcile() (reconcile.Result, error) {
	for _, ir := range r.Reconcilers {
		result, err := ir.Reconcile()
		if result.Requeue || err != nil {
			return result, err
		}
	}

	return reconcile.Result{}, nil
}

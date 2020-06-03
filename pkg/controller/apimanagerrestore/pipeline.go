package apimanagerrestore

import (
	"fmt"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Step interface {
	Execute() (reconcile.Result, error)
	Completed() (bool, error)
	Identifier() string
}

type APIManagerRestoreBaseStep struct {
	*APIManagerRestoreLogicReconciler
}

type Pipeline interface {
	Execute() (reconcile.Result, error)
	Completed() (bool, error)
	Step(identifier string) Step
}

type PipelineBuilder interface {
	Build() Pipeline
	AddStep(step Step) error
}

type RestorePipeline struct {
	Steps []Step
}

type RestorePipelineBuilder struct {
	steps []Step
}

func NewRestorePipelineBuilder() RestorePipelineBuilder {
	return RestorePipelineBuilder{}
}

func (r *RestorePipelineBuilder) Build() (*RestorePipeline, error) {
	return &RestorePipeline{
		Steps: r.steps,
	}, nil
}

func (r *RestorePipelineBuilder) AddStep(step Step) error {
	stepID := step.Identifier()
	for _, step := range r.steps {
		if step.Identifier() == stepID {
			return fmt.Errorf("Step '%s' already added in pipeline", stepID)
		}
	}
	r.steps = append(r.steps, step)
	return nil
}

func (p *RestorePipeline) Execute() (reconcile.Result, error) {
	for _, step := range p.Steps {
		stepCompleted, err := step.Completed()
		if err != nil {
			return reconcile.Result{}, err
		}
		if stepCompleted {
			continue
		}

		res, err := step.Execute()
		if res.Requeue || err != nil {
			return res, err
		}

		// We check again after performing the step execution to verify we are in
		// the desired state. // TODO should we change this? this has been added
		// due to the Execute step might not return an error or a requeue request
		// because everything has gone ok but we might not be in a completed state
		// just after that
		stepCompleted, err = step.Completed()
		if err != nil {
			return reconcile.Result{}, err
		}
		if !stepCompleted {
			return reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
		}
	}
	return reconcile.Result{}, nil
}

func (p *RestorePipeline) Completed() (bool, error) {
	for _, step := range p.Steps {
		stepCompleted, err := step.Completed()
		if err != nil {
			return false, err
		}
		if !stepCompleted {
			return false, nil
		}
	}
	return true, nil
}

func (p *RestorePipeline) Step(identifier string) Step {
	for _, step := range p.Steps {
		if step.Identifier() == identifier {
			return step
		}
	}
	return nil
}

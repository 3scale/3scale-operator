package helper

import (
	"fmt"
	"time"

	"github.com/go-logr/logr"
)

type task struct {
	Name string
	Run  func(interface{}) (error, []string)
}

// TaskRunner abstracts task running engine
type TaskRunner interface {
	Run() (error, []string)
	// AddTask register tasks to be executed sequentially
	// Tasks will be executed in order. First in, first to be executed.
	AddTask(string, func(interface{}) (error, []string))
}

type taskRunnerImpl struct {
	ctx      interface{}
	taskList []task
	logger   logr.Logger
}

// NewTaskRunner TaskRunner Constructor
func NewTaskRunner(ctx interface{}, logger logr.Logger) TaskRunner {
	return &taskRunnerImpl{
		ctx:      ctx,
		taskList: []task{},
		logger:   logger,
	}
}

func (t *taskRunnerImpl) Run() (error, []string) {
	var warning []string
	for _, task := range t.taskList {
		start := time.Now()
		err, warningMessages := task.Run(t.ctx)
		warning = append(warning, warningMessages...)
		if err != nil {
			return fmt.Errorf("Task failed %s: %w", task.Name, err), warning
		}

		elapsed := time.Since(start)
		t.logger.V(1).Info("Measure", task.Name, elapsed)
	}
	return nil, warning
}

func (t *taskRunnerImpl) AddTask(name string, f func(interface{}) (error, []string)) {
	t.taskList = append(t.taskList, task{Name: name, Run: f})
}

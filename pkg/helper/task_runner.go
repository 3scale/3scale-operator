package helper

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
)

type task struct {
	Name string
	Run  func(interface{}) error
}

// TaskRunner abstracts task running engine
type TaskRunner interface {
	Run() error
	// AddTask register tasks to be executed sequentially
	// Tasks will be executed in order. First in, first to be executed.
	AddTask(string, func(interface{}) error)
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

func (t *taskRunnerImpl) Run() error {
	reqerr := false
	for _, task := range t.taskList {
		start := time.Now()
		if err := task.Run(t.ctx); err != nil {
			if strings.Contains(err.Error(), "Method is used by the latest gateway configuration and cannot be deleted") {
				reqerr = true
				continue
			}
			return fmt.Errorf("Task failed %s: %w", task.Name, err)
		}
		elapsed := time.Since(start)
		t.logger.V(1).Info("Measure", task.Name, elapsed)
	}
	if reqerr {
		return fmt.Errorf("Method is used by the latest gateway configuration, it will be removed from 3scale only once the promotion without the mapping rule that uses the deleted method is done.")
	}
	return nil
}

func (t *taskRunnerImpl) AddTask(name string, f func(interface{}) error) {
	t.taskList = append(t.taskList, task{Name: name, Run: f})
}

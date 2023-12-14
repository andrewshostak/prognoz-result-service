package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/procyon-projects/chrono"
)

type Task struct {
	scheduler   chrono.TaskScheduler
	activeTasks map[string]chrono.ScheduledTask
}

func NewTaskScheduler(scheduler chrono.TaskScheduler) *Task {
	return &Task{scheduler: scheduler, activeTasks: map[string]chrono.ScheduledTask{}}
}

func (s *Task) Schedule(key string, task func(ctx context.Context), period time.Duration, startTime time.Time) error {
	scheduledTask, err := s.scheduler.ScheduleAtFixedRate(task, period, chrono.WithTime(startTime))

	if err != nil {
		return fmt.Errorf("failed to schedule a task: %w", err)
	}

	s.activeTasks[key] = scheduledTask

	return nil
}

func (s *Task) Cancel(key string) {
	scheduledTask, ok := s.activeTasks[key]
	if ok {
		scheduledTask.Cancel()
		delete(s.activeTasks, key)
	}
}

package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/procyon-projects/chrono"
)

type Task struct {
	scheduler chrono.TaskScheduler
}

func NewTaskScheduler(scheduler chrono.TaskScheduler) *Task {
	return &Task{scheduler: scheduler}
}

func (s *Task) Schedule(task func(ctx context.Context), period time.Duration, startTime time.Time) (*chrono.ScheduledRunnableTask, error) {
	scheduledTask, err := s.scheduler.ScheduleAtFixedRate(task, period, chrono.WithTime(startTime))

	if err != nil {
		return nil, fmt.Errorf("failed to schedule a task: %w", err)
	}

	return scheduledTask.(*chrono.ScheduledRunnableTask), nil
}

func (s *Task) Cancel(scheduledTask *chrono.ScheduledRunnableTask) {
	scheduledTask.Cancel()
}

package scheduler

import (
	"time"
)

// Task denotes a scheduled task
type Task struct {
	id       uint32                // Unique identifier for the task
	callback func(timeDiff uint32) // Callback function to execute when the task is due
}

type Scheduler struct {
	timestamp time.Time // The current timestamp of the scheduler
	tasks     []Task    // List of scheduled tasks
}

func (s *Scheduler) AddTask(task Task) {
	s.tasks = append(s.tasks, task)
}

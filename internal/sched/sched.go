package sched

import "sync"

type Task func()

type Scheduler struct {
	tasks chan Task
	done  chan struct{}
	wg    sync.WaitGroup // tracks in-flight tasks
}

func New(n int) *Scheduler {
	s := &Scheduler{
		tasks: make(chan Task, n*2),
		done:  make(chan struct{}),
	}
	for i := 0; i < n; i++ {
		go s.worker()
	}
	return s
}

func (s *Scheduler) worker() {
	for {
		select {
		case task := <-s.tasks:
			task()
			s.wg.Done() // mark task as complete
		case <-s.done:
			return
		}
	}
}

func (s *Scheduler) Submit(task Task) {
	s.wg.Add(1) // register task before sending
	s.tasks <- task
}

// Wait blocks until all submitted tasks are finished.
func (s *Scheduler) Wait() {
	s.wg.Wait()
}

// Stop waits for all tasks then shuts down workers.
func (s *Scheduler) Stop() {
	s.wg.Wait()      // wait for all tasks first
	close(s.done)    // then shut down workers
}
package utils

import (
	"sync"
)

// BatchJob is a utility class for running jobs concurrently
type BatchJob struct {
	jobs []func() error
}

// NewBatchJob creates a new empty batch job
func NewBatchJob() *BatchJob {
	return &BatchJob{}
}

// Add queues a new job, all calls to Add should be done before Run!
func (b *BatchJob) Add(job func() error) {
	b.jobs = append(b.jobs, job)
}

// Run runs all the jobs at the same time, each in their own goroutine.
func (b *BatchJob) Run() error {
	// Prepare channels
	wg := sync.WaitGroup{}
	errors := make(chan error)
	done := make(chan bool)

	// For each registered job, start a goroutine running the job
	for _, current := range b.jobs {
		wg.Add(1)
		go func(job func() error) {
			defer wg.Done()

			if err := job(); err != nil {
				errors <- err
			}
		}(current)
	}

	// Wait for every goroutine to have ended
	go func(wg *sync.WaitGroup) {
		wg.Wait()
		done <- true
	}(&wg)

	for {
		select {
		// Goroutine returned error
		case err := <-errors:
			return err
		// Everything finished
		case <-done:
			return nil
		}
	}
}

// RunSync executes the jobs synchronously
func (b *BatchJob) RunSync() error {
	for _, job := range b.jobs {
		if err := job(); err != nil {
			return err
		}
	}
	return nil
}

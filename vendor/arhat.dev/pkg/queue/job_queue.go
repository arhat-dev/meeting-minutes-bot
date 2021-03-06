//go:build !noqueue_jobqueue
// +build !noqueue_jobqueue

/*
Copyright 2019 The arhat.dev Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package queue

import (
	"fmt"
	"sync"
	"sync/atomic"

	"arhat.dev/pkg/errhelper"
)

// Errors for JobQueue
const (
	ErrJobDuplicated errhelper.ErrString = "job duplicat"
	ErrJobConflict   errhelper.ErrString = "job conflict"
	ErrJobCounteract errhelper.ErrString = "job counteract"
	ErrJobInvalid    errhelper.ErrString = "job invalid"
)

type JobAction uint8

const (
	// ActionInvalid to do nothing
	ActionInvalid JobAction = iota
	// ActionAdd to add or create some resource
	ActionAdd
	// ActionUpdate to update some resource
	ActionUpdate
	// ActionDelete to delete some resource
	ActionDelete
	// ActionCleanup to eliminate all side effects of the resource
	ActionCleanup
)

func (t JobAction) String() string {
	switch t {
	case ActionInvalid:
		return "Invalid"
	case ActionAdd:
		return "Add"
	case ActionUpdate:
		return "Update"
	case ActionDelete:
		return "Delete"
	case ActionCleanup:
		return "Cleanup"
	default:
		return "<unknown>"
	}
}

// Job item to record action and related resource object
type Job[K comparable] struct {
	Action JobAction
	Key    K
}

func (w Job[K]) String() string {
	return fmt.Sprintf("%s/%v", w.Action.String(), w.Key)
}

// NewJobQueue will create a stopped new job queue,
// you can offer job to it, but any acquire will fail until
// you have called its Resume()
func NewJobQueue[K comparable]() *JobQueue[K] {
	// prepare a closed channel for this job queue
	hasJob := make(chan struct{})
	close(hasJob)

	return &JobQueue[K]{
		queue: make([]Job[K], 0, 16),
		index: make(map[Job[K]]int),

		// set job queue to closed
		hasJob:     hasJob,
		chanClosed: true,

		paused: 1,
	}
}

// JobQueue is the queue data structure designed to reduce redundant job
// as much as possible
type JobQueue[K comparable] struct {
	queue []Job[K]
	index map[Job[K]]int

	hasJob chan struct{}
	// protected by atomic
	paused     uint32
	chanClosed bool

	mu sync.RWMutex
}

func (q *JobQueue[K]) has(action JobAction, key K) bool {
	_, ok := q.index[Job[K]{Action: action, Key: key}]
	return ok
}

func (q *JobQueue[K]) add(w Job[K]) {
	q.index[w] = len(q.queue)
	q.queue = append(q.queue, w)
}

func (q *JobQueue[K]) delete(action JobAction, key K) bool {
	jobToDelete := Job[K]{Action: action, Key: key}
	if idx, ok := q.index[jobToDelete]; ok {
		delete(q.index, jobToDelete)
		q.queue = append(q.queue[:idx], q.queue[idx+1:]...)

		q.buildIndex()

		return true
	}

	return false
}

func (q *JobQueue[K]) buildIndex() {
	for i, w := range q.queue {
		q.index[w] = i
	}
}

// Remains shows what job we are still meant to do
func (q *JobQueue[K]) Remains() []Job[K] {
	q.mu.RLock()
	defer q.mu.RUnlock()

	jobs := make([]Job[K], len(q.queue))
	for i, w := range q.queue {
		jobs[i] = Job[K]{Action: w.Action, Key: w.Key}
	}
	return jobs
}

// Find the scheduled job according to its key
func (q *JobQueue[K]) Find(key K) (Job[K], bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	for _, t := range []JobAction{ActionAdd, ActionUpdate, ActionDelete, ActionCleanup} {
		i, ok := q.index[Job[K]{Action: t, Key: key}]
		if ok {
			return q.queue[i], true
		}
	}

	return Job[K]{}, false
}

// Acquire a job item from the job queue
// if shouldAcquireMore is false, w will be an empty job
func (q *JobQueue[K]) Acquire() (w Job[K], shouldAcquireMore bool) {
	// wait until we have got some job to do
	// or we have paused the job queue
	<-q.hasJob

	if q.isPaused() {
		return Job[K]{Action: ActionInvalid}, false
	}

	q.mu.Lock()
	defer func() {
		if len(q.queue) == 0 {
			if !q.isPaused() {
				q.hasJob = make(chan struct{})
				q.chanClosed = false
			}
		}

		q.mu.Unlock()
	}()

	if len(q.queue) == 0 {
		return Job[K]{Action: ActionInvalid}, true
	}

	// pop first and rebuild index
	w = q.queue[0]
	q.delete(w.Action, w.Key)

	return w, true
}

// Offer a job item to the job queue
// if offered job was not added, an error result will return, otherwise nil
func (q *JobQueue[K]) Offer(w Job[K]) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if w.Action == ActionInvalid {
		return ErrJobInvalid
	}

	_, dup := q.index[w]
	if dup {
		return ErrJobDuplicated
	}

	switch w.Action {
	case ActionAdd:
		if q.has(ActionUpdate, w.Key) {
			return ErrJobConflict
		}

		q.add(w)
	case ActionUpdate:
		if q.has(ActionAdd, w.Key) || q.has(ActionDelete, w.Key) {
			return ErrJobConflict
		}

		q.add(w)
	case ActionDelete:
		// pod need to be deleted
		if q.has(ActionAdd, w.Key) {
			// cancel according create job
			q.delete(ActionAdd, w.Key)
			return ErrJobCounteract
		}

		if q.has(ActionUpdate, w.Key) {
			// if you want to delete it now, update operation doesn't matter any more
			q.delete(ActionUpdate, w.Key)
		}

		q.add(w)
	case ActionCleanup:
		// cleanup job only requires no duplication

		q.add(w)
	}

	// we reach here means we have added some job to the queue
	// we should signal those consumers to go for it
	select {
	case <-q.hasJob:
		// we can reach here means q.hasJob has been closed
	default:
		// release the signal
		close(q.hasJob)
		// mark the channel closed to prevent a second close which would panic
		q.chanClosed = true
	}

	return nil
}

// Resume do nothing but mark you can perform acquire
// actions to the job queue
func (q *JobQueue[K]) Resume() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.chanClosed && len(q.queue) == 0 {
		// reopen signal channel for wait
		q.hasJob = make(chan struct{})
		q.chanClosed = false
	}

	atomic.StoreUint32(&q.paused, 0)
}

// Pause do nothing but mark this job queue is closed,
// you should not perform acquire actions to the job queue
func (q *JobQueue[K]) Pause() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if !q.chanClosed {
		// close wait channel to prevent wait
		close(q.hasJob)
		q.chanClosed = true
	}

	atomic.StoreUint32(&q.paused, 1)
}

func (q *JobQueue[K]) Remove(w Job[K]) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.delete(w.Action, w.Key)
}

// isPaused is just for approximate check, for real
// closed state, need to hold the lock
func (q *JobQueue[K]) isPaused() bool {
	return atomic.LoadUint32(&q.paused) == 1
}

package ranking

import (
	"sync"
	"sync/atomic"
)

type PairwiseFeedback struct {
	Context  string
	Code     string
	Selected string
	Top      string
}

type UpdateQueue struct {
	model   *PersonalModel
	ch      chan PairwiseFeedback
	stopped chan struct{}
	done    chan struct{}
	once    sync.Once
	dropped atomic.Uint64
	started atomic.Bool
	looping atomic.Bool
}

func NewUpdateQueue(model *PersonalModel, capacity int) *UpdateQueue {
	if capacity <= 0 {
		capacity = 128
	}
	q := &UpdateQueue{
		model:   model,
		ch:      make(chan PairwiseFeedback, capacity),
		stopped: make(chan struct{}),
		done:    make(chan struct{}),
	}
	return q
}

func (q *UpdateQueue) Start() {
	if q == nil || q.model == nil {
		return
	}
	if q.started.Swap(true) {
		return
	}
	q.looping.Store(true)
	go q.loop()
}

func (q *UpdateQueue) loop() {
	defer close(q.done)
	for {
		select {
		case <-q.stopped:
			q.drain()
			return
		case fb, ok := <-q.ch:
			if !ok {
				return
			}
			q.model.AcceptFeedback(fb.Context, fb.Code, fb.Selected, fb.Top)
		}
	}
}

func (q *UpdateQueue) drain() {
	for {
		select {
		case fb := <-q.ch:
			q.model.AcceptFeedback(fb.Context, fb.Code, fb.Selected, fb.Top)
		default:
			return
		}
	}
}

func (q *UpdateQueue) Enqueue(fb PairwiseFeedback) bool {
	if q == nil || q.model == nil {
		return false
	}
	q.Start()
	select {
	case <-q.stopped:
		q.dropped.Add(1)
		return false
	default:
	}
	select {
	case q.ch <- fb:
		return true
	default:
		q.dropped.Add(1)
		return false
	}
}

func (q *UpdateQueue) Dropped() uint64 {
	if q == nil {
		return 0
	}
	return q.dropped.Load()
}

func (q *UpdateQueue) Close() {
	if q == nil {
		return
	}
	q.once.Do(func() {
		close(q.stopped)
	})
	if q.looping.Load() {
		<-q.done
	}
}

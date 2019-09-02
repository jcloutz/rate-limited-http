package rate_limited_http

import (
	"errors"
	"sync"
)

type Priority int

const (
	Immediate Priority = 1
	High      Priority = 2
	Medium    Priority = 3
	Low       Priority = 4
)

// QItem is an element within the queue
type QItem struct {
	task     *ApiTask
	priority Priority
}

// NewQItem creates and returns a new QItem
func NewQItem(payload *ApiTask, priority Priority) *QItem {
	return &QItem{
		task:     payload,
		priority: priority,
	}
}

// Priority gets the priority of them
func (n *QItem) Priority() Priority {
	return n.priority
}

// Task returns the items assigned ApiTask
func (n *QItem) Task() *ApiTask {
	return n.task
}

// priorityQueue will process items based on their weighted priority. Weighted
// priority is calculated based on the current queue size for a given priority
// its priority weight value
type priorityQueue struct {
	nodeHeap         map[Priority][]*QItem
	nextPriority     Priority
	heapWeight       map[Priority]float64
	heapWeightConfig map[Priority]float64
	length           int
	mu               sync.Mutex
}

type PriorityQueueOptions struct {
	WeightImmediate float64
	WeightHigh      float64
	WeightMedium    float64
	WeightLow       float64
}

func NewPriorityQueue(optionFunc ...func(options *PriorityQueueOptions)) *priorityQueue {
	opts := PriorityQueueOptions{
		WeightImmediate: 100,
		WeightHigh:      0.8,
		WeightMedium:    0.6,
		WeightLow:       0.3,
	}

	if optionFunc != nil {
		optionFunc[0](&opts)
	}

	return &priorityQueue{
		nodeHeap:   make(map[Priority][]*QItem, 4),
		heapWeight: make(map[Priority]float64, 4),
		heapWeightConfig: map[Priority]float64{
			Immediate: opts.WeightImmediate,
			High:      opts.WeightHigh,
			Medium:    opts.WeightMedium,
			Low:       opts.WeightLow,
		},
	}
}

// Len returns the overall length of the queue
func (pq *priorityQueue) Len() int {
	pq.mu.Lock()
	length := pq.length
	pq.mu.Unlock()

	return length
}

// Empty returns the empty state of the queue
func (pq *priorityQueue) Empty() bool {
	return pq.Len() == 0
}

// Push adds a new QItem into the queue based on priority
func (pq *priorityQueue) Push(node *QItem) {
	pq.mu.Lock()

	// push QItem into appropriate heap
	pq.nodeHeap[node.Priority()] = append(pq.nodeHeap[node.Priority()], node)

	// update the weighted value based on new length
	pq.updateHeapWeight(node.Priority())

	// increment length
	pq.length++

	// re-calc next heap to pull from
	pq.setNext()

	pq.mu.Unlock()
}

// Pop removes the highest priority item from the queue based on its calculated
// weighted priority
func (pq *priorityQueue) Pop() (*QItem, error) {
	pq.mu.Lock()

	if pq.length == 0 {
		pq.mu.Unlock()
		return nil, errors.New("queue pop failed, queue empty")
	}
	// grab the first QItem
	node := pq.nodeHeap[pq.nextPriority][0]

	// shift the slice by 1
	pq.nodeHeap[pq.nextPriority] = append(pq.nodeHeap[pq.nextPriority][:0], pq.nodeHeap[pq.nextPriority][1:]...)

	// shrink the heap size tracking
	pq.updateHeapWeight(node.Priority())

	// decrement length
	pq.length--

	// calculate the next priority level to fetch from based on weighting
	pq.setNext()

	pq.mu.Unlock()

	return node, nil
}

// updateHeapWeight will re-calculate the heap weight value based on it's length
func (pq *priorityQueue) updateHeapWeight(priority Priority) {
	var length = len(pq.nodeHeap[priority])
	var weightMultiplier = pq.heapWeightConfig[priority]

	pq.heapWeight[priority] = float64(length) * weightMultiplier
}

// setNext will determine the next priority heap to pull from based on the
// weighted values of each heap
func (pq *priorityQueue) setNext() {
	var nextPriority Priority = 0
	var max float64 = 0

	for priority := Immediate; priority <= Low; priority++ {
		weight := pq.heapWeight[priority]
		if weight > max {
			nextPriority = priority
			max = weight
		}
	}

	pq.nextPriority = nextPriority
}

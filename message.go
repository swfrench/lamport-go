package lamport

// Basic message structure for Lamport lock manipulation
type Message struct {
	Type int // Message type
	Proc int // Origin process
	Time int // Logical time on origin
}

// Message types
const (
	MessageRequest = iota // Request lock acquisition
	MessageRelease = iota // Release currently held lock
	MessageAck     = iota // Acknowledge lock request
)

// Implements heap.Interface from container/heap for Message

type MessageHeap []Message

func (mh MessageHeap) Len() int {
	return len(mh)
}

func (mh MessageHeap) Less(i, j int) bool {
	if mh[i].Time < mh[j].Time {
		return true
	}
	return mh[i].Proc < mh[j].Proc
}

func (mh MessageHeap) Swap(i, j int) {
	mh[i], mh[j] = mh[j], mh[i]
}

func (mh *MessageHeap) Push(x interface{}) {
	*mh = append(*mh, x.(Message))
}

func (mh *MessageHeap) Pop() interface{} {
	prev := *mh
	n := len(prev)
	m := prev[n-1]
	*mh = prev[0 : n-1]
	return m
}

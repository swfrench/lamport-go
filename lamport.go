package lamport

import (
	"container/heap"
	"log"
	"sync"
	"time"
)

// Sleep time used in:
//  - polling for lock acquisition
//  - servicing incoming messages
const SleepTime = 50 * time.Millisecond

// Structure representing internal state of distributed lock
type LamportLockState struct {
	time int
	proc int
	seen []int
	chns []chan Message
	reqs *MessageHeap
	lock sync.Mutex
}

// Initialize the LamportLockState structure
func initState(p int, chns []chan Message) *LamportLockState {
	s := LamportLockState{
		time: 1,
		proc: p,
		seen: make([]int, len(chns)),
		chns: chns,
		reqs: &MessageHeap{}}
	heap.Init(s.reqs)
	return &s
}

// Send request to all other procs and it enqueue locally
func (state *LamportLockState) sendRequestMsg() {
	// lock the state structure
	state.lock.Lock()

	// advance logical time
	state.time += 1

	// initialize message and send
	m := Message{Type: MessageRequest, Time: state.time, Proc: state.proc}
	for p, chn := range state.chns {
		if p != state.proc {
			chn <- m
		}
	}

	// enqueue
	heap.Push(state.reqs, m)

	// unlock the state structure
	state.lock.Unlock()
}

// Send release to all other procs and dequeue locally
func (state *LamportLockState) sendReleaseMsg() {
	// check to make sure we really have the lock
	if !state.haveLock() {
		log.Fatal("Cannot send release if we do not have the lock")
	}

	// lock the state structure
	state.lock.Lock()

	// advance logical time
	state.time += 1

	// initialize message and send
	m := Message{Type: MessageRelease, Time: state.time, Proc: state.proc}
	for p, chn := range state.chns {
		if p != state.proc {
			chn <- m
		}
	}

	// dequeue
	heap.Pop(state.reqs)

	// unlock the state structure
	state.lock.Unlock()
}

// Send an acknowledgement message
// Not threadsafe on its own: called only from processMessage
func (state *LamportLockState) sendAckMsg(target int) {
	// advance logical time
	state.time += 1

	// initialize ack message and send
	r := Message{Type: MessageAck, Time: state.time, Proc: state.proc}
	state.chns[target] <- r
}

// Process the current message, updating time vector and heap
// Not threadsafe on its own: called only from serviceMessage (within locked region)
func (state *LamportLockState) processMessage(m Message) {
	// update the process-time vector and current time
	state.seen[m.Proc] = m.Time
	if m.Time > state.time {
		state.time = m.Time
	}

	// if needed (i.e. not just a MessageAck), update request heap
	if m.Type == MessageRequest {
		// new request: add to queue
		heap.Push(state.reqs, m)
		// reply with an acknowledgement
		state.sendAckMsg(m.Proc)
	} else if m.Type == MessageRelease {
		// release previous request: remove all matching entries
		kept := make([]Message, 0)
		for state.reqs.Len() > 0 {
			req := heap.Pop(state.reqs).(Message)
			if req.Proc != m.Proc {
				kept = append(kept, req)
			}
		}
		for _, req := range kept {
			heap.Push(state.reqs, req)
		}
	}
}

// Check if all *other* processes have advanced to later logical times
func (state *LamportLockState) allProcessesSeen(time int) bool {
	for p := range state.seen {
		if p != state.proc {
			if state.seen[p] < time {
				return false
			}
		}
	}
	return true
}

// Check whether the current process has the lock
func (state *LamportLockState) haveLock() bool {
	// peek at the head of the queue
	state.lock.Lock()
	if state.reqs.Len() > 0 {
		m := (*state.reqs)[0]
		if m.Proc == state.proc {
			allSeen := state.allProcessesSeen(m.Time)
			if allSeen {
				state.lock.Unlock()
				return true
			}
		}
	}
	state.lock.Unlock()
	return false
}

// Service one incoming message
func (state *LamportLockState) serviceMessage() {
	// lock the state structure
	state.lock.Lock()

	// attempt non-blocking recv from incoming channel
	select {
	case m := <-state.chns[state.proc]:
		state.processMessage(m)
	default:
	}

	// unlock the state structure
	state.lock.Unlock()
}

// Acquire the distributed lock
func (state *LamportLockState) Acquire() {
	// initiate new request
	state.sendRequestMsg()

	// now wait for acquisition ...
	for {
		ready := state.haveLock()
		if ready {
			return
		}
		time.Sleep(SleepTime)
	}
}

// Release the distributed lock
func (state *LamportLockState) Release() {
	state.sendReleaseMsg()
}

// Initialize the Lamport distributed lock, by:
//  - setting up the LamportLockState structure
//  - spinning up the progress goroutine
// The supplied array of channels are assumed to be *buffered* such that
// simultaneous Acquire() calls will not induce deadlock.
func Start(p int, chns []chan Message) *LamportLockState {
	// initialize distributed lock state
	state := initState(p, chns)

	// spin up progess routine
	go func(s *LamportLockState) {
		for {
			s.serviceMessage()
			time.Sleep(SleepTime)
		}
	}(state)

	// return the state struct
	return state
}

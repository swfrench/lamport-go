A Lamport (1978) distributed lock in Go
=======================================

The distributed "lock" originally described by Lamport (1978), based on a set of
communicating processes with associated logical clocks, provides a neat toy
problem for getting acquainted with Go - particularly Goroutines and channels.

In this implementation, communication between "processes" is implemented over Go
channels (with the former represented by interacting Goroutines in the
accompanying demo code). 
While the use of channels as the primary interface for communication may appear
limiting (i.e. to the single-node case), these can of course act as conduits to
more complex logic for exchanging messages over a network - thereby making the
example truly distributed.

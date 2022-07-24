// Package fsm implements a finite state machine
package fsm

// TODO

type State[V any] struct {
	state uint32

	value V
}

func (s *State[V]) Value() V { return s.value }
func (s *State[V]) Next()    {}

type Callback interface {
}

type Event struct {
}

// Machine is a state less structure to move state into next state
type Machine[V any] struct {
	// AvailableTransitions
}

func (m *Machine[S]) AddTransition() {

}

func (m *Machine[S]) Next(s S) {

}

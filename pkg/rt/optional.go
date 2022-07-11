package rt

import "sync/atomic"

func NewOptionalValue[T any](x T) Optional[T] {
	return Optional[T]{
		hasValue: 1,
		value:    x,
	}
}

type Optional[T any] struct {
	hasValue uint32
	value    T
}

// IsNil returns true if there is no value set
func (o *Optional[T]) IsNil() bool { return atomic.LoadUint32(&o.hasValue) == 0 }
func (o *Optional[T]) Get() T      { return o.value }
func (o *Optional[T]) GetPtr() *T  { return &o.value }
func (o *Optional[T]) Set(x T) {
	atomic.StoreUint32(&o.hasValue, 1)
	o.value = x
}

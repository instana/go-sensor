package internal

import "sync/atomic"

type Flag struct {
	value int32
}

func (f *Flag) SetIfUnset() bool {
	return atomic.CompareAndSwapInt32(&f.value, 0, 1)
}

func (f *Flag) UnsetIfSet() bool {
	return atomic.CompareAndSwapInt32(&f.value, 1, 0)
}

func (f *Flag) Set() {
	atomic.StoreInt32(&f.value, 1)
}

func (f *Flag) Unset() {
	atomic.StoreInt32(&f.value, 0)
}

func (f *Flag) IsSet() bool {
	return atomic.LoadInt32(&f.value) == 1
}

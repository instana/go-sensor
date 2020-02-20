package autoprofile

import "sync/atomic"

type flag struct {
	value int32
}

func (f *flag) SetIfUnset() bool {
	return atomic.CompareAndSwapInt32(&f.value, 0, 1)
}

func (f *flag) UnsetIfSet() bool {
	return atomic.CompareAndSwapInt32(&f.value, 1, 0)
}

func (f *flag) Set() {
	atomic.StoreInt32(&f.value, 1)
}

func (f *flag) Unset() {
	atomic.StoreInt32(&f.value, 0)
}

func (f *flag) IsSet() bool {
	return atomic.LoadInt32(&f.value) == 1
}

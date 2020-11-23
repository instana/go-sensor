package internal

import "sync/atomic"

// Flag is a boolean value that can be set and unset atomically
type Flag struct {
	value int32
}

// SetIfUnset sets the Flag to true if it's false and returns whether
// the value has been changed
func (f *Flag) SetIfUnset() bool {
	return atomic.CompareAndSwapInt32(&f.value, 0, 1)
}

// UnsetIfSet sets the Flag to false if it's true and returns whether
// the value has been changed
func (f *Flag) UnsetIfSet() bool {
	return atomic.CompareAndSwapInt32(&f.value, 1, 0)
}

// Set sets the Flag to true
func (f *Flag) Set() {
	atomic.StoreInt32(&f.value, 1)
}

// Unset sets the Flag to false
func (f *Flag) Unset() {
	atomic.StoreInt32(&f.value, 0)
}

// IsSet returns whether the Flag is set to true
func (f *Flag) IsSet() bool {
	return atomic.LoadInt32(&f.value) == 1
}

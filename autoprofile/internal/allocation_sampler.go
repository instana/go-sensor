// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package internal

import (
	"bytes"
	"errors"
	"runtime/pprof"

	"github.com/instana/go-sensor/autoprofile/internal/pprof/profile"
)

// AllocationSampler collects information about the number of memory allocations
type AllocationSampler struct{}

// NewAllocationSampler initializes a new allocation sampler
func NewAllocationSampler() *AllocationSampler {
	return &AllocationSampler{}
}

// Reset is a no-op for allocation sampler
func (as *AllocationSampler) Reset() {}

// Start is a no-op for allocation sampler
func (as *AllocationSampler) Start() error { return nil }

// Stop is a no-op for allocation sampler
func (as *AllocationSampler) Stop() error { return nil }

// Profile retrieves the head profile and converts it to the profile.Profile
func (as *AllocationSampler) Profile(duration int64, timespan int64) (*Profile, error) {
	hp, err := as.readHeapProfile()
	if err != nil {
		return nil, err
	}

	if hp == nil {
		return nil, errors.New("no profile returned")
	}

	top, err := as.createAllocationCallGraph(hp)
	if err != nil {
		return nil, err
	}

	roots := make([]*CallSite, 0)
	for _, child := range top.children {
		roots = append(roots, child)
	}

	return NewProfile(CategoryMemory, TypeMemoryAllocation, UnitByte, roots, duration, timespan), nil
}

func (as *AllocationSampler) createAllocationCallGraph(p *profile.Profile) (*CallSite, error) {
	// find "inuse_space" type index
	inuseSpaceTypeIndex := -1
	for i, s := range p.SampleType {
		if s.Type == "inuse_space" {
			inuseSpaceTypeIndex = i
			break
		}
	}

	// find "inuse_objects" type index
	inuseObjectsTypeIndex := -1
	for i, s := range p.SampleType {
		if s.Type == "inuse_objects" {
			inuseObjectsTypeIndex = i
			break
		}
	}

	if inuseSpaceTypeIndex == -1 || inuseObjectsTypeIndex == -1 {
		return nil, errors.New("unrecognized profile data")
	}

	// build call graph
	top := NewCallSite("", "", 0)

	for _, s := range p.Sample {
		if shouldSkipStack(s) {
			continue
		}

		value := s.Value[inuseSpaceTypeIndex]
		if value == 0 {
			continue
		}

		count := s.Value[inuseObjectsTypeIndex]
		current := top
		for i := len(s.Location) - 1; i >= 0; i-- {
			l := s.Location[i]
			funcName, fileName, fileLine := readFuncInfo(l)

			current = current.FindOrAddChild(funcName, fileName, fileLine)
		}

		current.Increment(float64(value), int64(count))
	}

	return top, nil
}

func (as *AllocationSampler) readHeapProfile() (*profile.Profile, error) {
	buf := bytes.NewBuffer(nil)
	if err := pprof.WriteHeapProfile(buf); err != nil {
		return nil, err
	}

	p, err := profile.Parse(buf)
	if err != nil {
		return nil, err
	}

	if err := symbolizeProfile(p); err != nil {
		return nil, err
	}

	if err := p.CheckValid(); err != nil {
		return nil, err
	}

	return p, nil
}

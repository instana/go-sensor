// (c) Copyright IBM Corp. 2021
// (c) Copyright Instana Inc. 2020

package internal

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"runtime/pprof"

	"github.com/instana/go-sensor/autoprofile/internal/pprof/profile"
)

type blockValues struct {
	delay       float64
	contentions int64
}

// BlockSampler collects information about goroutine blocking events, such as waiting on
// synchronization primitives. This sampler uses the runtime blocking profiler, enabling and
// disabling it for a period of time.
type BlockSampler struct {
	top            *CallSite
	prevValues     map[string]blockValues
	partialProfile *pprof.Profile
}

// NewBlockSampler initializes a new blocking events sampler
func NewBlockSampler() *BlockSampler {
	bs := &BlockSampler{
		top:            nil,
		prevValues:     make(map[string]blockValues),
		partialProfile: nil,
	}

	return bs
}

// Reset resets the state of a BlockSampler, starting a new call tree
func (bs *BlockSampler) Reset() {
	bs.top = NewCallSite("", "", 0)
}

// Start enables the reporting of blocking events
func (bs *BlockSampler) Start() error {
	bs.partialProfile = pprof.Lookup("block")
	if bs.partialProfile == nil {
		return errors.New("No block profile found")
	}

	runtime.SetBlockProfileRate(1e6)

	return nil
}

// Stop disables the reporting of blocking events and gathers the collected information
// into a profile
func (bs *BlockSampler) Stop() error {
	runtime.SetBlockProfileRate(0)

	p, err := bs.collectProfile()
	if err != nil {
		return err
	}

	if p == nil {
		return errors.New("no profile returned")
	}

	if err := bs.updateBlockProfile(p); err != nil {
		return err
	}

	return nil
}

// Profile return the collected profile for a given time span
func (bs *BlockSampler) Profile(duration, timespan int64) (*Profile, error) {
	roots := make([]*CallSite, 0)
	for _, child := range bs.top.children {
		roots = append(roots, child)
	}
	p := NewProfile(CategoryTime, TypeBlockingCalls, UnitMillisecond, roots, duration, timespan)
	return p, nil
}

func (bs *BlockSampler) updateBlockProfile(p *profile.Profile) error {
	contentionIndex := -1
	delayIndex := -1
	for i, s := range p.SampleType {
		if s.Type == "contentions" {
			contentionIndex = i
		} else if s.Type == "delay" {
			delayIndex = i
		}
	}

	if contentionIndex == -1 || delayIndex == -1 {
		return errors.New("Unrecognized profile data")
	}

	for _, s := range p.Sample {
		if shouldSkipStack(s) {
			continue
		}

		delay := float64(s.Value[delayIndex])
		contentions := s.Value[contentionIndex]

		valueKey := generateValueKey(s)
		delay, contentions = bs.getValueChange(valueKey, delay, contentions)

		if contentions == 0 || delay == 0 {
			continue
		}

		// to milliseconds
		delay = delay / 1e6

		current := bs.top
		for i := len(s.Location) - 1; i >= 0; i-- {
			l := s.Location[i]
			funcName, fileName, fileLine := readFuncInfo(l)

			current = current.FindOrAddChild(funcName, fileName, fileLine)
		}
		current.Increment(delay, contentions)
	}

	return nil
}

func generateValueKey(s *profile.Sample) string {
	var key string
	for _, l := range s.Location {
		key += fmt.Sprintf("%v:", l.Address)
	}

	return key
}

func (bs *BlockSampler) getValueChange(key string, delay float64, contentions int64) (float64, int64) {
	pv := bs.prevValues[key]

	delayChange := delay - pv.delay
	contentionsChange := contentions - pv.contentions

	pv.delay = delay
	pv.contentions = contentions
	bs.prevValues[key] = pv

	return delayChange, contentionsChange
}

func (bs *BlockSampler) collectProfile() (*profile.Profile, error) {
	buf := bytes.NewBuffer(nil)

	if err := bs.partialProfile.WriteTo(buf, 0); err != nil {
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

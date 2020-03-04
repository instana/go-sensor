package internal

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"runtime/pprof"

	profile "github.com/instana/go-sensor/autoprofile/internal/pprof/profile"
)

type blockValues struct {
	delay       float64
	contentions int64
}

type BlockSampler struct {
	top            *CallSite
	prevValues     map[string]*blockValues
	partialProfile *pprof.Profile
}

func NewBlockSampler() *BlockSampler {
	bs := &BlockSampler{
		top:            nil,
		prevValues:     make(map[string]*blockValues),
		partialProfile: nil,
	}

	return bs
}

func (bs *BlockSampler) Reset() {
	bs.top = NewCallSite("", "", 0)
}

func (bs *BlockSampler) Start() error {
	err := bs.startBlockSampler()
	if err != nil {
		return err
	}

	return nil
}

func (bs *BlockSampler) Stop() error {
	p, err := bs.stopBlockSampler()
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
	key := ""
	for _, l := range s.Location {
		key += fmt.Sprintf("%v:", l.Address)
	}

	return key
}

func (bs *BlockSampler) getValueChange(key string, delay float64, contentions int64) (float64, int64) {
	if pv, exists := bs.prevValues[key]; exists {
		delayChange := delay - pv.delay
		contentionsChange := contentions - pv.contentions

		pv.delay = delay
		pv.contentions = contentions

		return delayChange, contentionsChange
	} else {
		bs.prevValues[key] = &blockValues{
			delay:       delay,
			contentions: contentions,
		}

		return delay, contentions
	}
}

func (bs *BlockSampler) startBlockSampler() error {
	bs.partialProfile = pprof.Lookup("block")
	if bs.partialProfile == nil {
		return errors.New("No block profile found")
	}

	runtime.SetBlockProfileRate(1e6)

	return nil
}

func (bs *BlockSampler) stopBlockSampler() (*profile.Profile, error) {
	runtime.SetBlockProfileRate(0)

	var buf bytes.Buffer

	w := bufio.NewWriter(&buf)
	if err := bs.partialProfile.WriteTo(w, 0); err != nil {
		return nil, err
	}

	w.Flush()

	r := bufio.NewReader(&buf)
	p, err := profile.Parse(r)
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

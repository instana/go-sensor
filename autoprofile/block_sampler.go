package autoprofile

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"runtime/pprof"

	profile "github.com/instana/go-sensor/autoprofile/pprof/profile"
)

type blockValues struct {
	delay       float64
	contentions int64
}

type blockSampler struct {
	top            *callSite
	prevValues     map[string]*blockValues
	partialProfile *pprof.Profile
}

func newBlockSampler() *blockSampler {
	bs := &blockSampler{
		top:            nil,
		prevValues:     make(map[string]*blockValues),
		partialProfile: nil,
	}

	return bs
}

func (bs *blockSampler) resetSampler() {
	bs.top = newCallSite("", "", 0)
}

func (bs *blockSampler) startSampler() error {
	err := bs.startBlockSampler()
	if err != nil {
		return err
	}

	return nil
}

func (bs *blockSampler) stopSampler() error {
	p, err := bs.stopBlockSampler()
	if err != nil {
		return err
	}
	if p == nil {
		return errors.New("no profile returned")
	}

	if uerr := bs.updateBlockProfile(p); uerr != nil {
		return uerr
	}

	return nil
}

func (bs *blockSampler) buildProfile(duration int64, timespan int64) (*Profile, error) {
	roots := make([]*callSite, 0)
	for _, child := range bs.top.children {
		roots = append(roots, child)
	}
	p := newProfile(categoryTime, typeBlockingCalls, unitMillisecond, roots, duration, timespan)
	return p, nil
}

func (bs *blockSampler) updateBlockProfile(p *profile.Profile) error {
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

			if shouldSkipFrame(fileName, funcName) {
				continue
			}

			current = current.findOrAddChild(funcName, fileName, fileLine)
		}
		current.increment(delay, contentions)
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

func (bs *blockSampler) getValueChange(key string, delay float64, contentions int64) (float64, int64) {
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

func (bs *blockSampler) startBlockSampler() error {
	bs.partialProfile = pprof.Lookup("block")
	if bs.partialProfile == nil {
		return errors.New("No block profile found")
	}

	runtime.SetBlockProfileRate(1e6)

	return nil
}

func (bs *blockSampler) stopBlockSampler() (*profile.Profile, error) {
	runtime.SetBlockProfileRate(0)

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	err := bs.partialProfile.WriteTo(w, 0)
	if err != nil {
		return nil, err
	}

	w.Flush()
	r := bufio.NewReader(&buf)

	if p, perr := profile.Parse(r); perr == nil {
		if serr := symbolizeProfile(p); serr != nil {
			return nil, serr
		}

		if verr := p.CheckValid(); verr != nil {
			return nil, verr
		}

		return p, nil
	} else {
		return nil, perr
	}
}

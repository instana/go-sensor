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

type BlockValues struct {
	delay       float64
	contentions int64
}

type BlockSampler struct {
	profiler       *autoProfiler
	top            *CallSite
	prevValues     map[string]*BlockValues
	partialProfile *pprof.Profile
}

func newBlockSampler(profiler *autoProfiler) *BlockSampler {
	bs := &BlockSampler{
		profiler:       profiler,
		top:            nil,
		prevValues:     make(map[string]*BlockValues),
		partialProfile: nil,
	}

	return bs
}

func (bs *BlockSampler) resetSampler() {
	bs.top = newCallSite("", "", 0)
}

func (bs *BlockSampler) startSampler() error {
	err := bs.startBlockSampler()
	if err != nil {
		return err
	}

	return nil
}

func (bs *BlockSampler) stopSampler() error {
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

func (bs *BlockSampler) buildProfile(duration int64, timespan int64) (*Profile, error) {
	roots := make([]*CallSite, 0)
	for _, child := range bs.top.children {
		roots = append(roots, child)
	}
	p := newProfile(CategoryTime, TypeBlockingCalls, UnitMillisecond, roots, duration, timespan)
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
		if !bs.profiler.IncludeSensorFrames && isSensorStack(s) {
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

			if (!bs.profiler.IncludeSensorFrames && isSensorFrame(fileName)) || funcName == "runtime.goexit" {
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

func (bs *BlockSampler) getValueChange(key string, delay float64, contentions int64) (float64, int64) {
	if pv, exists := bs.prevValues[key]; exists {
		delayChange := delay - pv.delay
		contentionsChange := contentions - pv.contentions

		pv.delay = delay
		pv.contentions = contentions

		return delayChange, contentionsChange
	} else {
		bs.prevValues[key] = &BlockValues{
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

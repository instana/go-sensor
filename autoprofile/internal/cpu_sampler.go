package internal

import (
	"bytes"
	"errors"
	"runtime/pprof"
	"time"

	"github.com/instana/go-sensor/autoprofile/internal/pprof/profile"
)

type CPUSampler struct {
	top       *CallSite
	buf       *bytes.Buffer
	startNano int64
}

func NewCPUSampler() *CPUSampler {
	return &CPUSampler{}
}

func (cs *CPUSampler) Reset() {
	cs.top = NewCallSite("", "", 0)
}

func (cs *CPUSampler) Start() error {
	if cs.buf != nil {
		return nil
	}

	cs.buf = bytes.NewBuffer(nil)
	cs.startNano = time.Now().UnixNano()

	if err := pprof.StartCPUProfile(cs.buf); err != nil {
		return err
	}

	return nil
}

func (cs *CPUSampler) Stop() error {
	if cs.buf == nil {
		return nil
	}

	pprof.StopCPUProfile()

	p, err := cs.collectProfile()
	if err != nil {
		return err
	}

	if p == nil {
		return errors.New("no profile returned")
	}

	if uerr := cs.updateCPUProfile(p); uerr != nil {
		return uerr
	}

	return nil
}

func (cs *CPUSampler) Profile(duration int64, timespan int64) (*Profile, error) {
	roots := make([]*CallSite, 0)
	for _, child := range cs.top.children {
		roots = append(roots, child)
	}
	p := NewProfile(CategoryCPU, TypeCPUUsage, UnitMillisecond, roots, duration, timespan)

	return p, nil
}

func (cs *CPUSampler) updateCPUProfile(p *profile.Profile) error {
	samplesIndex := -1
	cpuIndex := -1
	for i, s := range p.SampleType {
		if s.Type == "samples" {
			samplesIndex = i
		} else if s.Type == "cpu" {
			cpuIndex = i
		}
	}

	if samplesIndex == -1 || cpuIndex == -1 {
		return errors.New("Unrecognized profile data")
	}

	// build call graph
	for _, s := range p.Sample {
		if shouldSkipStack(s) {
			continue
		}

		stackSamples := s.Value[samplesIndex]
		stackDuration := float64(s.Value[cpuIndex])

		current := cs.top
		for i := len(s.Location) - 1; i >= 0; i-- {
			l := s.Location[i]
			funcName, fileName, fileLine := readFuncInfo(l)

			current = current.FindOrAddChild(funcName, fileName, fileLine)
		}

		current.Increment(stackDuration, stackSamples)
	}

	return nil
}

func (cs *CPUSampler) collectProfile() (*profile.Profile, error) {
	defer func() {
		cs.buf = nil
	}()

	p, err := profile.Parse(cs.buf)
	if err != nil {
		return nil, err
	}

	if p.TimeNanos == 0 {
		p.TimeNanos = cs.startNano
	}

	if p.DurationNanos == 0 {
		p.DurationNanos = time.Now().UnixNano() - cs.startNano
	}

	if err := symbolizeProfile(p); err != nil {
		return nil, err
	}

	if err := p.CheckValid(); err != nil {
		return nil, err
	}

	return p, nil
}

func readFuncInfo(l *profile.Location) (funcName string, fileName string, fileLine int64) {
	for li := range l.Line {
		if fn := l.Line[li].Function; fn != nil {
			return fn.Name, fn.Filename, l.Line[li].Line
		}
	}

	return "", "", 0
}

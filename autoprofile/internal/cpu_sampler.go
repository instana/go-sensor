package internal

import (
	"bufio"
	"bytes"
	"errors"
	"runtime/pprof"
	"time"

	"github.com/instana/go-sensor/autoprofile/internal/pprof/profile"
)

type CPUSampler struct {
	top        *CallSite
	profWriter *bufio.Writer
	profBuffer *bytes.Buffer
	startNano  int64
}

func NewCPUSampler() *CPUSampler {
	cs := &CPUSampler{
		top:        nil,
		profWriter: nil,
		profBuffer: nil,
		startNano:  0,
	}

	return cs
}

func (cs *CPUSampler) Reset() {
	cs.top = NewCallSite("", "", 0)
}

func (cs *CPUSampler) Start() error {
	err := cs.startCPUSampler()
	if err != nil {
		return err
	}

	return nil
}

func (cs *CPUSampler) Stop() error {
	p, err := cs.stopCPUSampler()
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

func (cs *CPUSampler) startCPUSampler() error {
	cs.profBuffer = &bytes.Buffer{}
	cs.profWriter = bufio.NewWriter(cs.profBuffer)
	cs.startNano = time.Now().UnixNano()

	err := pprof.StartCPUProfile(cs.profWriter)
	if err != nil {
		return err
	}

	return nil
}

func (cs *CPUSampler) stopCPUSampler() (*profile.Profile, error) {
	pprof.StopCPUProfile()

	cs.profWriter.Flush()
	r := bufio.NewReader(cs.profBuffer)

	if p, perr := profile.Parse(r); perr == nil {
		cs.profWriter = nil
		cs.profBuffer = nil

		if p.TimeNanos == 0 {
			p.TimeNanos = cs.startNano
		}
		if p.DurationNanos == 0 {
			p.DurationNanos = time.Now().UnixNano() - cs.startNano
		}

		if serr := symbolizeProfile(p); serr != nil {
			return nil, serr
		}

		if verr := p.CheckValid(); verr != nil {
			return nil, verr
		}

		return p, nil
	} else {
		cs.profWriter = nil
		cs.profBuffer = nil

		return nil, perr
	}
}

func readFuncInfo(l *profile.Location) (funcName string, fileName string, fileLine int64) {
	for li := range l.Line {
		if fn := l.Line[li].Function; fn != nil {
			return fn.Name, fn.Filename, l.Line[li].Line
		}
	}

	return "", "", 0
}

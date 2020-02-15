package autoprofile

import (
	"bufio"
	"bytes"
	"errors"
	"runtime/pprof"

	profile "github.com/instana/go-sensor/autoprofile/pprof/profile"
)

type AllocationSampler struct {
	profiler *AutoProfiler
}

func newAllocationSampler(profiler *AutoProfiler) *AllocationSampler {
	as := &AllocationSampler{
		profiler: profiler,
	}

	return as
}

func (as *AllocationSampler) resetSampler() {
}

func (as *AllocationSampler) startSampler() error {
	return nil
}

func (as *AllocationSampler) stopSampler() error {
	return nil
}

func (as *AllocationSampler) buildProfile(duration int64, timespan int64) (*Profile, error) {
	hp, err := as.readHeapProfile()
	if err != nil {
		return nil, err
	}
	if hp == nil {
		return nil, errors.New("no profile returned")
	}

	if top, aerr := as.createAllocationCallGraph(hp); err != nil {
		return nil, aerr
	} else {
		roots := make([]*CallSite, 0)
		for _, child := range top.children {
			roots = append(roots, child)
		}
		p := newProfile(CategoryMemory, TypeMemoryAllocation, UnitByte, roots, duration, timespan)
		return p, nil
	}
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

	// find "inuse_space" type index
	inuseObjectsTypeIndex := -1
	for i, s := range p.SampleType {
		if s.Type == "inuse_objects" {
			inuseObjectsTypeIndex = i
			break
		}
	}

	if inuseSpaceTypeIndex == -1 || inuseObjectsTypeIndex == -1 {
		return nil, errors.New("Unrecognized profile data")
	}

	// build call graph
	top := newCallSite("", "", 0)

	for _, s := range p.Sample {
		if !as.profiler.IncludeSensorFrames && isSensorStack(s) {
			continue
		}

		value := s.Value[inuseSpaceTypeIndex]
		count := s.Value[inuseObjectsTypeIndex]
		if value == 0 {
			continue
		}

		curren := top
		for i := len(s.Location) - 1; i >= 0; i-- {
			l := s.Location[i]
			funcName, fileName, fileLine := readFuncInfo(l)

			if (!as.profiler.IncludeSensorFrames && isSensorFrame(fileName)) || funcName == "runtime.goexit" {
				continue
			}

			curren = curren.findOrAddChild(funcName, fileName, fileLine)
		}
		curren.increment(float64(value), int64(count))
	}

	return top, nil
}

func (as *AllocationSampler) readHeapProfile() (*profile.Profile, error) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	err := pprof.WriteHeapProfile(w)
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

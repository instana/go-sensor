package internal

import (
	"bytes"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	RuntimeGolang = "golang"

	CategoryCPU    = "cpu"
	CategoryMemory = "memory"
	CategoryTime   = "time"

	TypeCPUUsage         = "cpu-usage"
	TypeMemoryAllocation = "memory-allocations"
	TypeBlockingCalls    = "blocking-calls"

	UnitSample      = "sample"
	UnitMillisecond = "millisecond"
	UnitMicrosecond = "microsecond"
	UnitNanosecond  = "nanosecond"
	UnitByte        = "byte"
	UnitKilobyte    = "kilobyte"
	UnitPercent     = "percent"
)

type CallSite struct {
	MethodName  string
	FileName    string
	FileLine    int64
	Metadata    map[string]string
	measurement float64
	numSamples  int64
	children    map[string]*CallSite
	updateLock  *sync.RWMutex
}

type Profile struct {
	ID        string
	ProcessID string
	Runtime   string
	Category  string
	Type      string
	Unit      string
	Roots     []*CallSite
	Duration  int64
	Timespan  int64
	Timestamp int64
}

func NewProfile(category string, typ string, unit string, roots []*CallSite, duration int64, timespan int64) *Profile {
	p := &Profile{
		ProcessID: strconv.Itoa(os.Getpid()),
		ID:        GenerateUUID(),
		Runtime:   RuntimeGolang,
		Category:  category,
		Type:      typ,
		Unit:      unit,
		Roots:     roots,
		Duration:  duration / int64(time.Millisecond),
		Timespan:  timespan * 1000,
		Timestamp: time.Now().Unix() * 1000,
	}

	return p
}

func (p *Profile) ToMap() map[string]interface{} {
	rootsMap := make([]interface{}, 0)

	for _, root := range p.Roots {
		rootsMap = append(rootsMap, root.ToMap())
	}

	profileMap := map[string]interface{}{
		"pid":       p.ProcessID,
		"id":        p.ID,
		"runtime":   p.Runtime,
		"category":  p.Category,
		"type":      p.Type,
		"unit":      p.Unit,
		"roots":     rootsMap,
		"duration":  p.Duration,
		"timespan":  p.Timespan,
		"timestamp": p.Timestamp,
	}

	return profileMap
}

func NewCallSite(methodName string, fileName string, fileLine int64) *CallSite {
	cn := &CallSite{
		MethodName: methodName,
		FileName:   fileName,
		FileLine:   fileLine,
		children:   make(map[string]*CallSite),
		updateLock: &sync.RWMutex{},
	}

	return cn
}

func (cs *CallSite) FindOrAddChild(methodName, fileName string, fileLine int64) *CallSite {
	child := cs.findChild(methodName, fileName, fileLine)
	if child == nil {
		child = NewCallSite(methodName, fileName, fileLine)
		cs.addChild(child)
	}

	return child
}

func (cs *CallSite) Increment(value float64, numSamples int64) {
	cs.measurement += value
	cs.numSamples += numSamples
}

func (cs *CallSite) Measurement() (value float64, numSamples int64) {
	return cs.measurement, cs.numSamples
}

func (cs *CallSite) ToMap() map[string]interface{} {
	childrenMap := make([]interface{}, 0)
	for _, child := range cs.children {
		childrenMap = append(childrenMap, child.ToMap())
	}

	m, ns := cs.Measurement()
	callSiteMap := map[string]interface{}{
		"method_name": cs.MethodName,
		"file_name":   cs.FileName,
		"file_line":   cs.FileLine,
		"measurement": m,
		"num_samples": ns,
		"children":    childrenMap,
	}

	return callSiteMap
}

func (cs *CallSite) findChild(methodName, fileName string, fileLine int64) *CallSite {
	cs.updateLock.RLock()
	defer cs.updateLock.RUnlock()

	if child, exists := cs.children[createKey(methodName, fileName, fileLine)]; exists {
		return child
	}

	return nil
}

func (cs *CallSite) addChild(child *CallSite) {
	cs.updateLock.Lock()
	defer cs.updateLock.Unlock()

	cs.children[createKey(child.MethodName, child.FileName, child.FileLine)] = child
}

func createKey(methodName, fileName string, fileLine int64) string {
	var b bytes.Buffer

	b.WriteString(methodName)
	b.WriteString(" (")
	b.WriteString(fileName)
	b.WriteString(":")
	b.WriteString(strconv.FormatInt(fileLine, 10))
	b.WriteString(")")

	return b.String()
}

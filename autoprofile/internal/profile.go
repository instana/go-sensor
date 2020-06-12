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

// AgentProfile is a presenter type used to serialize a collected profile
// to JSON format supported by Instana profile sensor
type AgentProfile struct {
	ID        string          `json:"id"`
	ProcessID string          `json:"pid"`
	Runtime   string          `json:"runtime"`
	Category  string          `json:"category"`
	Type      string          `json:"type"`
	Unit      string          `json:"unit"`
	Roots     []AgentCallSite `json:"roots"`
	Duration  int64           `json:"duration"`
	Timespan  int64           `json:"timespan"`
	Timestamp int64           `json:"timestamp"`
}

func NewAgentProfile(p *Profile) AgentProfile {
	callSites := make([]AgentCallSite, 0, len(p.Roots))
	for _, root := range p.Roots {
		callSites = append(callSites, NewAgentCallSite(root))
	}

	return AgentProfile{
		ProcessID: p.ProcessID,
		ID:        p.ID,
		Runtime:   p.Runtime,
		Category:  p.Category,
		Type:      p.Type,
		Unit:      p.Unit,
		Roots:     callSites,
		Duration:  p.Duration,
		Timespan:  p.Timespan,
		Timestamp: p.Timestamp,
	}
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
	return &Profile{
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
}

// AgentCallSite is a presenter type used to serialize a call site
// to JSON format supported by Instana profile sensor
type AgentCallSite struct {
	MethodName  string          `json:"method_name"`
	FileName    string          `json:"file_name"`
	FileLine    int64           `json:"file_line"`
	Measurement float64         `json:"measurement"`
	NumSamples  int64           `json:"num_samples"`
	Children    []AgentCallSite `json:"children"`
}

func NewAgentCallSite(cs *CallSite) AgentCallSite {
	children := make([]AgentCallSite, 0, len(cs.children))
	for _, child := range cs.children {
		children = append(children, NewAgentCallSite(child))
	}

	m, ns := cs.Measurement()

	return AgentCallSite{
		MethodName:  cs.MethodName,
		FileName:    cs.FileName,
		FileLine:    cs.FileLine,
		Measurement: m,
		NumSamples:  ns,
		Children:    children,
	}
}

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

package autoprofile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

const RuntimeGolang string = "golang"

const CategoryCPU string = "cpu"
const CategoryMemory string = "memory"
const CategoryTime string = "time"

const TypeCPUUsage string = "cpu-usage"
const TypeMemoryAllocation string = "memory-allocations"
const TypeBlockingCalls string = "blocking-calls"

const UnitSample string = "sample"
const UnitMillisecond string = "millisecond"
const UnitMicrosecond string = "microsecond"
const UnitNanosecond string = "nanosecond"
const UnitByte string = "byte"
const UnitKilobyte string = "kilobyte"
const UnitPercent string = "percent"

type CallSite struct {
	methodName  string
	fileName    string
	fileLine    int64
	metadata    map[string]string
	measurement float64
	numSamples  int64
	counter     int64
	children    map[string]*CallSite
	updateLock  *sync.RWMutex
}

func newCallSite(methodName string, fileName string, fileLine int64) *CallSite {
	cn := &CallSite{
		methodName:  methodName,
		fileName:    fileName,
		fileLine:    fileLine,
		measurement: 0,
		numSamples:  0,
		children:    make(map[string]*CallSite),
		updateLock:  &sync.RWMutex{},
	}

	return cn
}

func createKey(methodName string, fileName string, fileLine int64) string {
	var b bytes.Buffer
	b.WriteString(methodName)
	b.WriteString(" (")
	b.WriteString(fileName)
	b.WriteString(":")
	b.WriteString(strconv.FormatInt(fileLine, 10))
	b.WriteString(")")
	return b.String()
}

func (cs *CallSite) findChild(methodName string, fileName string, fileLine int64) *CallSite {
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

	cs.children[createKey(child.methodName, child.fileName, child.fileLine)] = child
}

func (cs *CallSite) removeChild(child *CallSite) {
	cs.updateLock.Lock()
	defer cs.updateLock.Unlock()

	delete(cs.children, createKey(child.methodName, child.fileName, child.fileLine))
}

func (cs *CallSite) findOrAddChild(methodName string, fileName string, fileLine int64) *CallSite {
	child := cs.findChild(methodName, fileName, fileLine)
	if child == nil {
		child = newCallSite(methodName, fileName, fileLine)
		cs.addChild(child)
	}

	return child
}

func (cs *CallSite) filter(fromLevel int, min float64, max float64) {
	cs.filterLevel(1, fromLevel, min, max)
}

func (cs *CallSite) filterLevel(currentLevel int, fromLevel int, min float64, max float64) {
	for key, child := range cs.children {
		if currentLevel >= fromLevel && (child.measurement < min || child.measurement > max) {
			delete(cs.children, key)
		} else {
			child.filterLevel(currentLevel+1, fromLevel, min, max)
		}
	}
}

func (cs *CallSite) depth() int {
	max := 0
	for _, child := range cs.children {
		cd := child.depth()
		if cd > max {
			max = cd
		}
	}

	return max + 1
}

func (cs *CallSite) increment(value float64, numSamples int64) {
	cs.measurement += value
	cs.numSamples += numSamples
}

func (cs *CallSite) toMap() map[string]interface{} {
	childrenMap := make([]interface{}, 0)
	for _, child := range cs.children {
		childrenMap = append(childrenMap, child.toMap())
	}

	callSiteMap := map[string]interface{}{
		"method_name": cs.methodName,
		"file_name":   cs.fileName,
		"file_line":   cs.fileLine,
		"measurement": cs.measurement,
		"num_samples": cs.numSamples,
		"children":    childrenMap,
	}

	return callSiteMap
}

func (cs *CallSite) printLevel(level int) string {
	str := ""

	for i := 0; i < level; i++ {
		str += "  "
	}

	str += fmt.Sprintf("%v (%v:%v) - %v (%v)\n", cs.methodName, cs.fileName, cs.fileLine, cs.measurement, cs.numSamples)
	for _, child := range cs.children {
		str += child.printLevel(level + 1)
	}

	return str
}

type Profile struct {
	processID string
	id        string
	runtime   string
	category  string
	typ       string
	unit      string
	roots     []*CallSite
	duration  int64
	timespan  int64
	timestamp int64
}

func newProfile(category string, typ string, unit string, roots []*CallSite, duration int64, timespan int64) *Profile {
	p := &Profile{
		processID: strconv.Itoa(os.Getpid()),
		id:        generateUUID(),
		runtime:   RuntimeGolang,
		category:  category,
		typ:       typ,
		unit:      unit,
		roots:     roots,
		duration:  duration / int64(time.Millisecond),
		timespan:  timespan * 1000,
		timestamp: time.Now().Unix() * 1000,
	}

	return p
}

func (p *Profile) toMap() map[string]interface{} {
	rootsMap := make([]interface{}, 0)

	for _, root := range p.roots {
		rootsMap = append(rootsMap, root.toMap())
	}

	profileMap := map[string]interface{}{
		"pid":       p.processID,
		"id":        p.id,
		"runtime":   p.runtime,
		"category":  p.category,
		"type":      p.typ,
		"unit":      p.unit,
		"roots":     rootsMap,
		"duration":  p.duration,
		"timespan":  p.timespan,
		"timestamp": p.timestamp,
	}

	return profileMap
}

func (p *Profile) toIndentedJson() string {
	profileJson, _ := json.MarshalIndent(p.toMap(), "", "\t")
	return string(profileJson)
}

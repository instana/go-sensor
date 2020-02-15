package autoprofile

import (
	"testing"
)

func TestCallSiteFilter(t *testing.T) {
	root := newCallSite("root", "", 0)
	root.measurement = 10

	child1 := newCallSite("child1", "", 0)
	child1.measurement = 9
	root.addChild(child1)

	child2 := newCallSite("child2", "", 0)
	child2.measurement = 1
	root.addChild(child2)

	child2child1 := newCallSite("child2child1", "", 0)
	child2child1.measurement = 1
	child2.addChild(child2child1)

	root.filter(2, 3, 100)

	if root.findChild("child1", "", 0) == nil {
		t.Errorf("child1 should not be filtered")
	}

	if root.findChild("child2", "", 0) == nil {
		t.Errorf("child2 should not be filtered")
	}

	if child2.findChild("child2child1", "", 0) != nil {
		t.Errorf("child2child1 should be filtered")
	}
}

func TestCallSiteDepth(t *testing.T) {
	root := newCallSite("root", "", 0)

	child1 := newCallSite("child1", "", 0)
	root.addChild(child1)

	child2 := newCallSite("child2", "", 0)
	root.addChild(child2)

	child2child1 := newCallSite("child2child1", "", 0)
	child2.addChild(child2child1)

	if root.depth() != 3 {
		t.Errorf("root depth should be 3, but is %v", root.depth())
	}

	if child1.depth() != 1 {
		t.Errorf("child1 depth should be 1, but is %v", child1.depth())
	}

	if child2.depth() != 2 {
		t.Errorf("child2 depth should be 2, but is %v", child2.depth())
	}
}

func TestCallSiteIncrement(t *testing.T) {
	root := newCallSite("root", "", 0)

	root.increment(12.3, 1)
	root.increment(0, 0)
	root.increment(5, 2)

	if root.measurement != 17.3 {
		t.Errorf("root measurement should be 17.3, but is %v", root.measurement)
	}

	if root.numSamples != 3 {
		t.Errorf("root numSamples should be 3, but is %v", root.numSamples)
	}
}

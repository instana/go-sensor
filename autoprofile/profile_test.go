package autoprofile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCallSite_Filter(t *testing.T) {
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

	assert.Equal(t, child1, root.findChild("child1", "", 0))
	assert.Equal(t, child2, root.findChild("child2", "", 0))
	assert.Nil(t, child2.findChild("child2child1", "", 0))
}

func TestCallSite_Depth(t *testing.T) {
	root := newCallSite("root", "", 0)

	child1 := newCallSite("child1", "", 0)
	root.addChild(child1)

	child2 := newCallSite("child2", "", 0)
	root.addChild(child2)

	child2child1 := newCallSite("child2child1", "", 0)
	child2.addChild(child2child1)

	assert.Equal(t, 3, root.depth())
	assert.Equal(t, 1, child1.depth())
	assert.Equal(t, 2, child2.depth())
}

func TestCallSite_Increment(t *testing.T) {
	root := newCallSite("root", "", 0)

	root.increment(12.3, 1)
	root.increment(0, 0)
	root.increment(5, 2)

	assert.Equal(t, 17.3, root.measurement)
	assert.EqualValues(t, 3, root.numSamples)
}

package support_test

import (
	"testing"

	"github.com/go-spatial/proj4go/support"

	"github.com/stretchr/testify/assert"
)

func TestPair(t *testing.T) {
	assert := assert.New(t)

	pair1 := support.Pair{Key: "k", Value: "v"}
	pair2 := support.Pair{Key: "k", Value: "v"}

	assert.Equal(pair1, pair2)
}

func TestPairListOperations(t *testing.T) {
	assert := assert.New(t)

	p1 := support.Pair{Key: "k1", Value: "v1"}
	p2 := support.Pair{Key: "k2", Value: "v2"}

	pl1 := support.NewPairList()
	assert.NotNil(pl1)
	assert.Equal(0, pl1.Len())

	pl1.Add(p1)
	assert.Equal(1, pl1.Len())
	assert.Equal(p1, pl1.Get(0))

	pl2 := support.NewPairList()
	pl2.Add(p2)

	pl1.AddList(pl2)
	assert.Equal(2, pl1.Len())

	assert.Equal("k1", pl1.Get(0).Key)
	assert.Equal("k2", pl1.Get(1).Key)
	assert.True(pl1.ContainsKey("k1"))
	assert.False(pl1.ContainsKey("k3"))
	assert.Equal(1, pl1.CountKey("k1"))

	pl1.Add(p1)
	assert.Equal(2, pl1.CountKey("k1"))
	assert.Equal(0, pl1.CountKey("k3"))
	assert.Equal(p1, pl1.Get(2))
}

func TestPairListGets(t *testing.T) {
	assert := assert.New(t)

	p1 := support.Pair{Key: "k1", Value: "v1"}
	p2 := support.Pair{Key: "k2", Value: "2.2"}
	p3 := support.Pair{Key: "k3", Value: "3.0,3.1,3.2"}
	p4 := support.Pair{Key: "k4", Value: "678"}

	pl := support.NewPairList()
	pl.Add(p1)
	pl.Add(p2)
	pl.Add(p3)
	pl.Add(p4)

	vs, ok := pl.GetAsString("k99")
	assert.False(ok)

	vs, ok = pl.GetAsString("k2")
	assert.True(ok)
	assert.Equal("2.2", vs)

	vf, ok := pl.GetAsFloat("k2")
	assert.True(ok)
	assert.Equal(2.2, vf)

	_, ok = pl.GetAsFloat("k1")
	assert.False(ok)

	vfs, ok := pl.GetAsFloats("k2")
	assert.True(ok)
	assert.Equal([]float64{2.2}, vfs)

	vfs, ok = pl.GetAsFloats("k3")
	assert.True(ok)
	assert.Equal([]float64{3.0, 3.1, 3.2}, vfs)

	vi, ok := pl.GetAsInt("k4")
	assert.True(ok)
	assert.Equal(678, vi)
}

func TestPairListParsing(t *testing.T) {
	assert := assert.New(t)

	pl, err := support.NewPairListFromString("")
	assert.NoError(err)
	assert.Equal(0, pl.Len())

	_, err = support.NewPairListFromString("k1=v1=v2")
	assert.Error(err)

	pl, err = support.NewPairListFromString("  +k1=v1 +k2=v2 k3=v3 \t\t k4= k5")
	assert.NoError(err)
	assert.Equal(5, pl.Len())

	assert.True(pl.ContainsKey("k1"))
	assert.True(pl.ContainsKey("k2"))
	assert.True(pl.ContainsKey("k3"))
	assert.True(pl.ContainsKey("k4"))
	assert.True(pl.ContainsKey("k5"))
	assert.False(pl.ContainsKey("+k1"))

	assert.Equal("", pl.Get(3).Value)
	assert.Equal("", pl.Get(4).Value)
}

package pilosa

import (
	"fmt"
	"testing"

	"github.com/umbel/pilosa/pql"
	"github.com/yasushi-saito/rbtree"
)

func (bp BitmapPair) String() string {
	return fmt.Sprintf("Tile:%v", bp.ID)
}

func (b *Bitmap) String() string {
	return fmt.Sprintf("Profiles:%v", b.Bits())
}

type bicliqueResult struct {
	Tiles    []uint64
	Profiles []uint64
}

func (b bicliqueResult) Equals(other bicliqueResult) bool {
	if len(b.Tiles) != len(other.Tiles) || len(b.Profiles) != len(other.Profiles) {
		return false
	}
	for i, t := range b.Tiles {
		if other.Tiles[i] != t {
			return false
		}
	}
	for i, t := range b.Profiles {
		if other.Profiles[i] != t {
			return false
		}
	}
	return true
}

func TestBicliqueFind(t *testing.T) {
	// set up 3 tiles and 3 profiles
	bms := make([]BitmapPair, 3)
	bms[0].ID = 0
	bms[0].Bitmap = &Bitmap{tree: rbtree.NewTree(rbtreeItemCompare)}
	bms[0].Bitmap.setBit(2)
	bms[1].ID = 1
	bms[1].Bitmap = &Bitmap{tree: rbtree.NewTree(rbtreeItemCompare)}
	bms[1].Bitmap.setBit(0)
	bms[1].Bitmap.setBit(2)
	bms[2].ID = 2
	bms[2].Bitmap = &Bitmap{tree: rbtree.NewTree(rbtreeItemCompare)}
	//bms[2].Bitmap.setBit(0)
	bms[2].Bitmap.setBit(1)
	bms[2].Bitmap.setBit(2)

	results := make(chan Biclique, 30)
	bicliqueFind(bms, nil, []BitmapPair{}, bms, []BitmapPair{}, &pql.Bicliques{}, results) // could block if too many results
	close(results)

	expectedBicliques := []bicliqueResult{
		{Tiles: []uint64{0, 1, 2}, Profiles: []uint64{2}},
		{Tiles: []uint64{1}, Profiles: []uint64{0, 2}},
		{Tiles: []uint64{2}, Profiles: []uint64{1, 2}},
	}
	i := 0
	for res := range results {
		biclique := newBicliqueResult(res)
		if !biclique.Equals(expectedBicliques[i]) {
			t.Fatalf("unexpected biclique result %v, expected: %v", biclique, expectedBicliques[i])
		}
		i++
	}
}

func newBicliqueResult(biclique Biclique) bicliqueResult {
	bcr := bicliqueResult{Tiles: biclique.Tiles, Profiles: biclique.Bitmap.Bits()}
	return bcr
}

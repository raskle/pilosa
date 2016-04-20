package pilosa

import (
	"fmt"

	"github.com/umbel/pilosa/pql"
)

type Biclique struct {
	Tiles  []uint64
	Bitmap *Bitmap `json:"-"`
	Count  uint64
	Score  uint64
}

func (b Biclique) Equals(other Biclique) bool {
	if b.Count != other.Count || b.Score != other.Score {
		return false
	}
	if len(b.Tiles) != len(other.Tiles) {
		return false
	}
	set := make(map[uint64]struct{})
	for _, t := range b.Tiles {
		set[t] = struct{}{}
	}
	for _, t := range other.Tiles {
		_, ok := set[t]
		if !ok {
			return false
		}
		delete(set, t)
	}
	return true
}

func (b Biclique) String() string {
	return fmt.Sprintf("{Tiles: %v, Count: %v, Score: %v}", b.Tiles, b.Count, b.Score)
}

type BCList []Biclique

func (bcl BCList) Len() int {
	return len(bcl)
}
func (bcl BCList) Less(i, j int) bool {
	return bcl[i].Score > bcl[j].Score
}

func (bcl BCList) Swap(i, j int) {
	bcl[i], bcl[j] = bcl[j], bcl[i]
}

func (f *Fragment) MaxBiclique(c *pql.Bicliques) chan Biclique {
	n := c.N
	f.mu.Lock()
	f.cache.Invalidate()
	pairs := f.cache.Top() // slice of bitmapPairs
	f.mu.Unlock()

	topPairs := pairs
	if n < len(pairs) {
		topPairs = pairs[:n]
	}

	bicliques := make(chan Biclique, 0) // TODO tweak length for perf
	go func() {
		bicliqueFind(topPairs, nil, []BitmapPair{}, topPairs, []BitmapPair{}, c, bicliques)
		close(bicliques)
	}()

	return bicliques
}

func bicliqueFind(G []BitmapPair, L *Bitmap, R []BitmapPair, P []BitmapPair, Q []BitmapPair, c *pql.Bicliques, results chan Biclique) {
	// G is topPairs
	// L should start with all bits set (L == U) (it will actually start nil, and we'll special case it below)
	// R starts empty
	// P starts as topPairs (all tiles are candidates)
	// Q starts empty

	for len(P) > 0 {
		// P ← P\{x};
		x := P[0]
		P = P[1:]

		// R ← R ∪ {x};
		newR := append(R, x)

		//  L' ← {u ∈ L | (u, x) ∈ E(G)};
		var newL *Bitmap
		if L == nil {
			newL = x.Bitmap.Clone()
		} else {
			newL = L.Clone()
		}
		newL = newL.Intersect(x.Bitmap)
		newLcnt := newL.BitCount()

		// P' ← ∅; Q' ← ∅;
		newP := []BitmapPair{}
		newQ := []BitmapPair{}

		// Check maximality.
		isMaximal := true
		for _, v := range Q {
			// get the neighbors of v in L'
			neighbors := v.Bitmap.Intersect(newL)
			ncnt := neighbors.BitCount()
			// Observation 4: end of branch
			if ncnt == newLcnt {
				isMaximal = false
				break
			} else if ncnt > 0 {
				newQ = append(newQ, v)
			}
		}

		if isMaximal {
			for _, v := range P {
				// get the neighbors of v in L'
				neighbors := v.Bitmap.Intersect(newL)
				ncnt := neighbors.BitCount()
				// Observation 3: expand to maximal
				if ncnt == newLcnt {
					newR = append(newR, v)
				} else if ncnt > 0 {
					// keep vertice adjacent to some vertex in newL
					newP = append(newP, v)
				}
			}
			// report newR as maximal biclique
			report(newR, c, results)
			if len(newP) > 0 {
				bicliqueFind(G, newL, newR, newP, newQ, c, results)
			}
		}
		Q = append(Q, x)
	}
}

func report(bmPairs []BitmapPair, c *pql.Bicliques, results chan Biclique) {
	tiles := getTileIDs(bmPairs)
	if len(tiles) < c.MinTiles {
		return
	}
	bicliqueBitmap := intersectPairs(bmPairs)
	count := bicliqueBitmap.BitCount()
	bicliqueBitmap.SetCount(count)
	results <- Biclique{
		Tiles:  tiles,
		Bitmap: bicliqueBitmap,
		Count:  count,
		Score:  count * uint64(len(tiles)),
	}
}

func getTileIDs(pairs []BitmapPair) []uint64 {
	tileIDs := make([]uint64, len(pairs))
	for i := 0; i < len(pairs); i++ {
		tileIDs[i] = pairs[i].ID
	}
	return tileIDs
}

// intersectPairs generates a bitmap which represents all profiles which have all of the tiles in pairs
func intersectPairs(pairs []BitmapPair) *Bitmap {
	result := pairs[0].Bitmap.Clone()
	for i := 1; i < len(pairs); i++ {
		result = result.Intersect(pairs[i].Bitmap)
	}
	result.SetCount(result.BitCount())
	return result
}

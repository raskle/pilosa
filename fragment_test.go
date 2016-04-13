package pilosa_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"fmt"
	"math/rand"
	"sort"

	"github.com/davecgh/go-spew/spew"
	"github.com/umbel/pilosa"
)

// SliceWidth is a helper reference to use when testing.
const SliceWidth = pilosa.SliceWidth

// Ensure a fragment can set a bit and retrieve it.
func TestFragment_SetBit(t *testing.T) {
	f := MustOpenFragment("d", "f", 0)
	defer f.Close()

	// Set bits on the fragment.
	if _, err := f.SetBit(120, 1, nil, 0); err != nil {
		t.Fatal(err)
	} else if _, err := f.SetBit(120, 6, nil, 0); err != nil {
		t.Fatal(err)
	} else if _, err := f.SetBit(121, 0, nil, 0); err != nil {
		t.Fatal(err)
	}

	// Verify counts on bitmaps.
	if n := f.Bitmap(120).Count(); n != 2 {
		t.Fatalf("unexpected count: %d", n)
	} else if n := f.Bitmap(121).Count(); n != 1 {
		t.Fatalf("unexpected count: %d", n)
	}

	// Close and reopen the fragment & verify the data.
	if err := f.Reopen(); err != nil {
		t.Fatal(err)
	} else if n := f.Bitmap(120).Count(); n != 2 {
		t.Fatalf("unexpected count (reopen): %d", n)
	} else if n := f.Bitmap(121).Count(); n != 1 {
		t.Fatalf("unexpected count (reopen): %d", n)
	}
}

// Ensure a fragment can clear a set bit.
func TestFragment_ClearBit(t *testing.T) {
	f := MustOpenFragment("d", "f", 0)
	defer f.Close()

	// Set and then clear bits on the fragment.
	if _, err := f.SetBit(1000, 1, nil, 0); err != nil {
		t.Fatal(err)
	} else if _, err := f.SetBit(1000, 2, nil, 0); err != nil {
		t.Fatal(err)
	} else if _, err := f.ClearBit(1000, 1); err != nil {
		t.Fatal(err)
	}

	// Verify count on bitmap.
	if n := f.Bitmap(1000).Count(); n != 1 {
		t.Fatalf("unexpected count: %d", n)
	}

	// Close and reopen the fragment & verify the data.
	if err := f.Reopen(); err != nil {
		t.Fatal(err)
	} else if n := f.Bitmap(1000).Count(); n != 1 {
		t.Fatalf("unexpected count (reopen): %d", n)
	}
}

// Ensure a fragment can snapshot correctly.
func TestFragment_Snapshot(t *testing.T) {
	f := MustOpenFragment("d", "f", 0)
	defer f.Close()

	// Set and then clear bits on the fragment.
	if _, err := f.SetBit(1000, 1, nil, 0); err != nil {
		t.Fatal(err)
	} else if _, err := f.SetBit(1000, 2, nil, 0); err != nil {
		t.Fatal(err)
	} else if _, err := f.ClearBit(1000, 1); err != nil {
		t.Fatal(err)
	}

	// Snapshot bitmap and verify data.
	if err := f.Snapshot(); err != nil {
		t.Fatal(err)
	} else if n := f.Bitmap(1000).Count(); n != 1 {
		t.Fatalf("unexpected count: %d", n)
	}

	// Close and reopen the fragment & verify the data.
	if err := f.Reopen(); err != nil {
		t.Fatal(err)
	} else if n := f.Bitmap(1000).Count(); n != 1 {
		t.Fatalf("unexpected count (reopen): %d", n)
	}
}

// Test biclique junk
func TestFragment_maxBiclique(t *testing.T) {
	f := MustOpenFragment("d", "f", 0)
	defer f.Close()

	// test data
	for i := uint64(0); i < 10; i++ {
		if i%2 == 0 {
			f.MustSetBits(1, i)
		}
		if i%2 == 1 {
			f.MustSetBits(2, i)
		}
		if i != 5 {
			f.MustSetBits(3, i)
		}
		f.MustSetBits(uint64(rand.Intn(8)), i)
		f.MustSetBits(uint64(rand.Intn(8)), i)
		f.MustSetBits(uint64(rand.Intn(8)), i)
		f.MustSetBits(uint64(rand.Intn(8)), i)
		f.MustSetBits(uint64(rand.Intn(8)), i)
		f.MustSetBits(uint64(rand.Intn(8)), i)
	}

	ret := f.MaxBiclique(3)
	retList := make(pilosa.BCList, 0)
	for bc := range ret {
		retList = append(retList, bc)
	}
	sort.Sort(retList)
	fmt.Println("SWEET HOT RESULTS BELOW!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	for i := 0; i < len(retList); i++ {
		fmt.Println(retList[i])
		if i > 5 {
			break
		}
	}
	fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!END SWEET HOT RESULTS")
	// fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
	// spew.Dump(ret)

}

// Ensure a fragment can return the top n results.
func TestFragment_Top(t *testing.T) {
	f := MustOpenFragment("d", "f", 0)
	defer f.Close()

	// Set bits on the bitmaps 100, 101, & 102.
	f.MustSetBits(100, 1, 3, 200)
	f.MustSetBits(101, 1)
	f.MustSetBits(102, 1, 2)

	// Retrieve top bitmaps.
	if pairs, err := f.Top(pilosa.TopOptions{N: 2}); err != nil {
		t.Fatal(err)
	} else if len(pairs) != 2 {
		t.Fatalf("unexpected count: %d", len(pairs))
	} else if pairs[0] != (pilosa.Pair{Key: 100, Count: 3}) {
		t.Fatalf("unexpected pair(0): %v", pairs[0])
	} else if pairs[1] != (pilosa.Pair{Key: 102, Count: 2}) {
		t.Fatalf("unexpected pair(1): %v", pairs[1])
	}
}

// Ensure a fragment can filter bitmaps when retrieving the top n bitmaps.
func TestFragment_Top_Filter(t *testing.T) {
	f := MustOpenFragment("d", "f", 0)
	defer f.Close()

	// Set bits on the bitmaps 100, 101, & 102.
	f.MustSetBits(100, 1, 3, 200)
	f.MustSetBits(101, 1)
	f.MustSetBits(102, 1, 2)

	// Assign attributes.
	f.BitmapAttrStore.SetAttrs(101, map[string]interface{}{"x": uint64(10)})
	f.BitmapAttrStore.SetAttrs(102, map[string]interface{}{"x": uint64(20)})

	// Retrieve top bitmaps.
	if pairs, err := f.Top(pilosa.TopOptions{
		N:            2,
		FilterField:  "x",
		FilterValues: []interface{}{uint64(10), uint64(15), uint64(20)},
	}); err != nil {
		t.Fatal(err)
	} else if len(pairs) != 2 {
		t.Fatalf("unexpected count: %d", len(pairs))
	} else if pairs[0] != (pilosa.Pair{Key: 102, Count: 2}) {
		t.Fatalf("unexpected pair(0): %v", pairs[0])
	} else if pairs[1] != (pilosa.Pair{Key: 101, Count: 1}) {
		t.Fatalf("unexpected pair(1): %v", pairs[1])
	}
}

// Ensure a fragment can return top bitmaps that intersect with an input bitmap.
func TestFragment_TopN_Intersect(t *testing.T) {
	f := MustOpenFragment("d", "f", 0)
	defer f.Close()

	// Create an intersecting input bitmap.
	src := pilosa.NewBitmap(1, 2, 3)

	// Set bits on various bitmaps.
	f.MustSetBits(100, 1, 10, 11, 12)    // one intersection
	f.MustSetBits(101, 1, 2, 3, 4)       // three intersections
	f.MustSetBits(102, 1, 2, 4, 5, 6)    // two intersections
	f.MustSetBits(103, 1000, 1001, 1002) // no intersection

	// Retrieve top bitmaps.
	if pairs, err := f.Top(pilosa.TopOptions{N: 3, Src: src}); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(pairs, []pilosa.Pair{
		{Key: 101, Count: 3},
		{Key: 102, Count: 2},
		{Key: 100, Count: 1},
	}) {
		t.Fatalf("unexpected pairs: %s", spew.Sdump(pairs))
	}
}

// Ensure a fragment can return top bitmaps that have many bits set.
func TestFragment_TopN_Intersect_Large(t *testing.T) {
	if testing.Short() {
		t.Skip("short mode")
	}

	f := MustOpenFragment("d", "f", 0)
	defer f.Close()

	// Create an intersecting input bitmap.
	src := pilosa.NewBitmap(
		980, 981, 982, 983, 984, 985, 986, 987, 988, 989,
		990, 991, 992, 993, 994, 995, 996, 997, 998, 999,
	)

	// Set bits on bitmaps 0 - 999. Higher bitmaps have higher bit counts.
	for i := uint64(0); i < 1000; i++ {
		for j := uint64(0); j < i; j++ {
			f.MustSetBits(i, j)
		}
	}

	// Retrieve top bitmaps.
	if pairs, err := f.Top(pilosa.TopOptions{N: 10, Src: src}); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(pairs, []pilosa.Pair{
		{Key: 999, Count: 19},
		{Key: 998, Count: 18},
		{Key: 997, Count: 17},
		{Key: 996, Count: 16},
		{Key: 995, Count: 15},
		{Key: 994, Count: 14},
		{Key: 993, Count: 13},
		{Key: 992, Count: 12},
		{Key: 991, Count: 11},
		{Key: 990, Count: 10},
	}) {
		t.Fatalf("unexpected pairs: %s", spew.Sdump(pairs))
	}
}

// Ensure a fragment can return top bitmaps when specified by ID.
func TestFragment_TopN_BitmapIDs(t *testing.T) {
	f := MustOpenFragment("d", "f", 0)
	defer f.Close()

	// Set bits on various bitmaps.
	f.MustSetBits(100, 1, 2, 3)
	f.MustSetBits(101, 4, 5, 6, 7)
	f.MustSetBits(102, 8, 9, 10, 11, 12)

	// Retrieve top bitmaps.
	if pairs, err := f.Top(pilosa.TopOptions{BitmapIDs: []uint64{100, 101, 200}}); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(pairs, []pilosa.Pair{
		{Key: 101, Count: 4},
		{Key: 100, Count: 3},
	}) {
		t.Fatalf("unexpected pairs: %s", spew.Sdump(pairs))
	}
}

// Ensure a fragment's cache can be persisted between restarts.
func TestFragment_LRUCache_Persistence(t *testing.T) {
	f := MustOpenFragment("d", "f", 0)
	defer f.Close()

	// Set bits on the fragment.
	for i := uint64(0); i < 1000; i++ {
		if _, err := f.SetBit(i, 0, nil, 0); err != nil {
			t.Fatal(err)
		}
	}

	// Verify correct cache type and size.
	if cache, ok := f.Cache().(*pilosa.LRUCache); !ok {
		t.Fatalf("unexpected cache: %T", f.Cache())
	} else if cache.Len() != 1000 {
		t.Fatalf("unexpected cache len: %d", cache.Len())
	}

	// Reopen the fragment.
	if err := f.Reopen(); err != nil {
		t.Fatal(err)
	}

	// Re-verify correct cache type and size.
	if cache, ok := f.Cache().(*pilosa.LRUCache); !ok {
		t.Fatalf("unexpected cache: %T", f.Cache())
	} else if cache.Len() != 1000 {
		t.Fatalf("unexpected cache len: %d", cache.Len())
	}
}

// Ensure a fragment's cache can be persisted between restarts.
func TestFragment_RankCache_Persistence(t *testing.T) {
	f := MustOpenFragment("d", "f.n", 0)
	defer f.Close()

	// Set bits on the fragment.
	for i := uint64(0); i < 1000; i++ {
		if _, err := f.SetBit(i, 0, nil, 0); err != nil {
			t.Fatal(err)
		}
	}

	// Verify correct cache type and size.
	if cache, ok := f.Cache().(*pilosa.RankCache); !ok {
		t.Fatalf("unexpected cache: %T", f.Cache())
	} else if cache.Len() != 1000 {
		t.Fatalf("unexpected cache len: %d", cache.Len())
	}

	// Reopen the fragment.
	if err := f.Reopen(); err != nil {
		t.Fatal(err)
	}

	// Re-verify correct cache type and size.
	if cache, ok := f.Cache().(*pilosa.RankCache); !ok {
		t.Fatalf("unexpected cache: %T", f.Cache())
	} else if cache.Len() != 1000 {
		t.Fatalf("unexpected cache len: %d", cache.Len())
	}
}

// Ensure a fragment can be copied to another fragment.
func TestFragment_WriteTo_ReadFrom(t *testing.T) {
	f0 := MustOpenFragment("d", "f", 0)
	defer f0.Close()

	// Set and then clear bits on the fragment.
	if _, err := f0.SetBit(1000, 1, nil, 0); err != nil {
		t.Fatal(err)
	} else if _, err := f0.SetBit(1000, 2, nil, 0); err != nil {
		t.Fatal(err)
	} else if _, err := f0.ClearBit(1000, 1); err != nil {
		t.Fatal(err)
	}

	// Verify cache is populated.
	if n := f0.Cache().Len(); n != 1 {
		t.Fatalf("unexpected cache size: %d", n)
	}

	// Write fragment to a buffer.
	var buf bytes.Buffer
	wn, err := f0.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}

	// Read into another fragment.
	f1 := MustOpenFragment("d", "f", 0)
	if rn, err := f1.ReadFrom(&buf); err != nil {
		t.Fatal(err)
	} else if wn != rn {
		t.Fatalf("read/write byte count mismatch: wn=%d, rn=%d", wn, rn)
	}

	// Verify cache is in other fragment.
	if n := f1.Cache().Len(); n != 1 {
		t.Fatalf("unexpected cache size: %d", n)
	}

	// Verify data in other fragment.
	if a := f1.Bitmap(1000).Bits(); !reflect.DeepEqual(a, []uint64{2}) {
		t.Fatalf("unexpected bits: %+v", a)
	}

	// Close and reopen the fragment & verify the data.
	if err := f1.Reopen(); err != nil {
		t.Fatal(err)
	} else if n := f1.Cache().Len(); n != 1 {
		t.Fatalf("unexpected cache size (reopen): %d", n)
	} else if a := f1.Bitmap(1000).Bits(); !reflect.DeepEqual(a, []uint64{2}) {
		t.Fatalf("unexpected bits (reopen): %+v", a)
	}
}

// Fragment is a test wrapper for pilosa.Fragment.
type Fragment struct {
	*pilosa.Fragment
	BitmapAttrStore *AttrStore
}

// NewFragment returns a new instance of Fragment with a temporary path.
func NewFragment(db, frame string, slice uint64) *Fragment {
	file, err := ioutil.TempFile("", "pilosa-fragment-")
	if err != nil {
		panic(err)
	}
	file.Close()

	f := &Fragment{
		Fragment:        pilosa.NewFragment(file.Name(), db, frame, slice),
		BitmapAttrStore: MustOpenAttrStore(),
	}
	f.Fragment.BitmapAttrStore = f.BitmapAttrStore.AttrStore
	return f
}

// MustOpenFragment creates and opens an fragment at a temporary path. Panic on error.
func MustOpenFragment(db, frame string, slice uint64) *Fragment {
	f := NewFragment(db, frame, slice)
	if err := f.Open(); err != nil {
		panic(err)
	}
	return f
}

// Close closes the fragment and removes all underlying data.
func (f *Fragment) Close() error {
	defer os.Remove(f.Path())
	defer os.Remove(f.CachePath())
	defer f.BitmapAttrStore.Close()
	return f.Fragment.Close()
}

// Reopen closes the fragment and reopens it as a new instance.
func (f *Fragment) Reopen() error {
	path := f.Path()
	if err := f.Fragment.Close(); err != nil {
		return err
	}

	f.Fragment = pilosa.NewFragment(path, f.DB(), f.Frame(), f.Slice())
	f.Fragment.BitmapAttrStore = f.BitmapAttrStore.AttrStore
	if err := f.Open(); err != nil {
		return err
	}
	return nil
}

// MustSetBits sets bits on a bitmap. Panic on error.
// This function does not accept a timestamp or quantum.
func (f *Fragment) MustSetBits(bitmapID uint64, profileIDs ...uint64) {
	for _, profileID := range profileIDs {
		if _, err := f.SetBit(bitmapID, profileID, nil, 0); err != nil {
			panic(err)
		}
	}
}

// MustClearBits clears bits on a bitmap. Panic on error.
func (f *Fragment) MustClearBits(bitmapID uint64, profileIDs ...uint64) {
	for _, profileID := range profileIDs {
		if _, err := f.ClearBit(bitmapID, profileID); err != nil {
			panic(err)
		}
	}
}

// BitmapAttrStore provides simple storage for attributes.
type BitmapAttrStore struct {
	attrs map[uint64]map[string]interface{}
}

// NewBitmapAttrStore returns a new instance of BitmapAttrStore.
func NewBitmapAttrStore() *BitmapAttrStore {
	return &BitmapAttrStore{
		attrs: make(map[uint64]map[string]interface{}),
	}
}

// BitmapAttrs returns the attributes set to a bitmap id.
func (s *BitmapAttrStore) BitmapAttrs(id uint64) (map[string]interface{}, error) {
	return s.attrs[id], nil
}

// SetBitmapAttrs assigns a set of attributes to a bitmap id.
func (s *BitmapAttrStore) SetBitmapAttrs(id uint64, m map[string]interface{}) {
	s.attrs[id] = m
}

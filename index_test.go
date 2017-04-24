package pilosa_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/pilosa/pilosa"
)

// Ensure index can open and retrieve a frame.
func TestIndex_CreateFrameIfNotExists(t *testing.T) {
	index := MustOpenIndex()
	defer index.Close()

	// Create frame.
	f, err := index.CreateFrameIfNotExists("f", pilosa.FrameOptions{})
	if err != nil {
		t.Fatal(err)
	} else if f == nil {
		t.Fatal("expected frame")
	}

	// Retrieve existing frame.
	other, err := index.CreateFrameIfNotExists("f", pilosa.FrameOptions{})
	if err != nil {
		t.Fatal(err)
	} else if f.Frame != other.Frame {
		t.Fatal("frame mismatch")
	}

	if f.Frame != index.Frame("f") {
		t.Fatal("frame mismatch")
	}
}

// Ensure index defaults the time quantum on new frames.
func TestIndex_CreateFrame_TimeQuantum(t *testing.T) {
	index := MustOpenIndex()
	defer index.Close()

	// Set index time quantum.
	if err := index.SetTimeQuantum(pilosa.TimeQuantum("YM")); err != nil {
		t.Fatal(err)
	}

	// Create frame.
	f, err := index.CreateFrame("f", pilosa.FrameOptions{})
	if err != nil {
		t.Fatal(err)
	} else if q := f.TimeQuantum(); q != pilosa.TimeQuantum("YM") {
		t.Fatalf("unexpected frame time quantum: %s", q)
	}
}

// Ensure index can delete a frame.
func TestIndex_DeleteFrame(t *testing.T) {
	index := MustOpenIndex()
	defer index.Close()

	// Create frame.
	if _, err := index.CreateFrameIfNotExists("f", pilosa.FrameOptions{}); err != nil {
		t.Fatal(err)
	}

	// Delete frame & verify it's gone.
	if err := index.DeleteFrame("f"); err != nil {
		t.Fatal(err)
	} else if index.Frame("f") != nil {
		t.Fatal("expected nil frame")
	}

	// Delete again to make sure it doesn't error.
	if err := index.DeleteFrame("f"); err != nil {
		t.Fatal(err)
	}
}

// Ensure index can set the default time quantum.
func TestIndex_SetTimeQuantum(t *testing.T) {
	index := MustOpenIndex()
	defer index.Close()

	// Set & retrieve time quantum.
	if err := index.SetTimeQuantum(pilosa.TimeQuantum("YMDH")); err != nil {
		t.Fatal(err)
	} else if q := index.TimeQuantum(); q != pilosa.TimeQuantum("YMDH") {
		t.Fatalf("unexpected quantum: %s", q)
	}

	// Reload index and verify that it is persisted.
	if err := index.Reopen(); err != nil {
		t.Fatal(err)
	} else if q := index.TimeQuantum(); q != pilosa.TimeQuantum("YMDH") {
		t.Fatalf("unexpected quantum (reopen): %s", q)
	}
}

// Index represents a test wrapper for pilosa.Index.
type Index struct {
	*pilosa.Index
}

// NewIndex returns a new instance of Index.
func NewIndex() *Index {
	path, err := ioutil.TempDir("", "pilosa-index-")
	if err != nil {
		panic(err)
	}
	index, err := pilosa.NewIndex(path, "i")
	if err != nil {
		panic(err)
	}
	return &Index{Index: index}
}

// MustOpenIndex returns a new, opened index at a temporary path. Panic on error.
func MustOpenIndex() *Index {
	index := NewIndex()
	if err := index.Open(); err != nil {
		panic(err)
	}
	return index
}

// Close closes the index and removes the underlying data.
func (i *Index) Close() error {
	defer os.RemoveAll(i.Path())
	return i.Index.Close()
}

// Reopen closes the index and reopens it.
func (i *Index) Reopen() error {
	var err error
	if err := i.Index.Close(); err != nil {
		return err
	}

	path, name := i.Path(), i.Name()
	i.Index, err = pilosa.NewIndex(path, name)
	if err != nil {
		return err
	}

	if err := i.Open(); err != nil {
		return err
	}
	return nil
}

// CreateFrame creates a frame with the given options.
func (i *Index) CreateFrame(name string, opt pilosa.FrameOptions) (*Frame, error) {
	f, err := i.Index.CreateFrame(name, opt)
	if err != nil {
		return nil, err
	}
	return &Frame{Frame: f}, nil
}

// CreateFrameIfNotExists creates a frame with the given options if it doesn't exist.
func (i *Index) CreateFrameIfNotExists(name string, opt pilosa.FrameOptions) (*Frame, error) {
	f, err := i.Index.CreateFrameIfNotExists(name, opt)
	if err != nil {
		return nil, err
	}
	return &Frame{Frame: f}, nil
}

// Ensure index can delete a frame.
func TestIndex_InvalidName(t *testing.T) {
	path, err := ioutil.TempDir("", "pilosa-index-")
	if err != nil {
		panic(err)
	}
	index, err := pilosa.NewIndex(path, "ABC")
	if index != nil {
		t.Fatalf("unexpected index name %s", index)
	}
}

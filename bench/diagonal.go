package bench

import (
	"fmt"

	"flag"
	"io/ioutil"

	"context"
	"time"
)

// DiagonalSetBits sets bits with increasing profile id and bitmap id.
type DiagonalSetBits struct {
	HasClient
	BaseBitmapID  int
	BaseProfileID int
	Iterations    int
	DB            string
}

func (b *DiagonalSetBits) Usage() string {
	return `
diagonal-set-bits sets bits with increasing profile id and bitmap id.

Usage: diagonal-set-bits [arguments]

The following arguments are available:

	-base-bitmap-id int
		bits being set will all be greater than BaseBitmapID

	-base-profile-id int
		profile id num to start from

	-iterations int
		number of bits to set

	-db string
		pilosa db to use

	-client-type string
		Can be 'single' (all agents hitting one host) or 'round_robin'

`[1:]
}

func (b *DiagonalSetBits) ConsumeFlags(args []string) ([]string, error) {
	fs := flag.NewFlagSet("DiagonalSetBits", flag.ContinueOnError)
	fs.SetOutput(ioutil.Discard)
	fs.IntVar(&b.BaseBitmapID, "base-bitmap-id", 0, "")
	fs.IntVar(&b.BaseProfileID, "base-profile-id", 0, "")
	fs.IntVar(&b.Iterations, "iterations", 100, "")
	fs.StringVar(&b.DB, "db", "benchdb", "")
	fs.StringVar(&b.ClientType, "client-type", "single", "")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return fs.Args(), nil
}

// Run runs the DiagonalSetBits benchmark
func (b *DiagonalSetBits) Run(agentNum int) map[string]interface{} {
	results := make(map[string]interface{})
	if b.cli == nil {
		results["error"] = fmt.Errorf("No client set for DiagonalSetBits agent: %v", agentNum)
		return results
	}
	s := NewStats()
	var start time.Time
	for n := 0; n < b.Iterations; n++ {
		iterID := agentizeNum(n, b.Iterations, agentNum)
		query := fmt.Sprintf("SetBit(%d, 'frame.n', %d)", b.BaseBitmapID+iterID, b.BaseProfileID+iterID)
		start = time.Now()
		b.cli.ExecuteQuery(context.TODO(), b.DB, query, true)
		s.Add(time.Now().Sub(start))
	}
	AddToResults(s, results)
	return results
}

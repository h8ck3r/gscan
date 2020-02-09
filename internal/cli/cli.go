package cli

import (
	"flag"
	"github.com/h8ck3r/gscan/internal/log"
	"github.com/h8ck3r/gscan/pkg/util"
	"os"
	"time"
)

var (
	Verbose bool
	GoroutineCap int
	Timeout time.Duration
	Targets []string
)

func Parse() {
	flag.BoolVar(&Verbose, "verbose", false, "verbose output")
	flag.IntVar(&GoroutineCap, "cap", 0, "maximum amount of concurrent goroutines")
	flag.DurationVar(&Timeout, "timeout", time.Millisecond * 500, "maximum time to wait for connection response")
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatalf("%s takes exactly one argument. %d provided\n", os.Args[0], flag.NArg())
	}
	var err error
	Targets, err = util.GetTargets(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
}


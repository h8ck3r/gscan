package cli

import (
	"flag"
	"github.com/h8ck3r/gscan/internal/log"
	"github.com/h8ck3r/gscan/pkg/util"
	"os"
	"regexp"
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
	Targets = getTargets(flag.Arg(0))
}

func getTargets(argument string) []string {
	var targets []string

	if regexp.MustCompile(`^([0-9]{1,3}\.){3}[0-9]{1,3}$`).MatchString(argument) {
		targets = []string{argument}
	} else if regexp.MustCompile(`^([0-9]{1,3}(-[0-9]{1,3})?\.){3}[0-9]{1,3}(-[0-9]{1,3})?$`).MatchString(argument) {
		_, _ = util.GetHostsForIPRange(argument)
		log.Fatalln("ip range definitions are not yet supported")
		os.Exit(1)
	} else if regexp.MustCompile(`^([0-9]{1,3}\.){3}[0-9]{1,3}/[0-9]{1,2}$`).MatchString(argument) {
		var err error
		targets, err = util.GetHostsForSubnet(argument)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatalf("invalid argument: %s\n", argument)
	}

	return targets
}


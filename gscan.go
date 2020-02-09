package main

import (
	"github.com/h8ck3r/gscan/internal/cli"
	"github.com/h8ck3r/gscan/internal/log"
	"github.com/h8ck3r/gscan/pkg/types"
)

var (
	scan types.Scan
	scanner types.Scanner
	verbose *bool
	err error
	hostResults []*types.HostResult
)

func main() {
	cli.Parse()
	verbose = cli.GetVerbose()
	scan.Targets, err = cli.GetTargets()
	if err != nil {
		log.Fatal(err)
	}

	scan.Protocol = cli.GetProtocol()
	scan.Timeout = cli.GetTimeout()
	scan.GoroutineCap = cli.GetGoroutineCap()

	scan.Ports = cli.GetPorts()
	scanner = &scan
	hostResults, err = scanner.Scan()
	if err != nil {
		log.Error(err)
	}
	for _, result := range hostResults {
		for _, portResult := range result.PortResults {
			if portResult.State == types.Open {
				log.Printf("discovered open port %d on %s\n", *portResult.Port, result.Host)
			} else {
				if *verbose {
					log.Printf("port %d on %s is closed\n", *portResult.Port, result.Host)
				}
			}
		}
	}
}
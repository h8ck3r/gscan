package main

import (
	"github.com/h8ck3r/gscan/internal/cli"
	"github.com/h8ck3r/gscan/internal/log"
	"github.com/h8ck3r/gscan/pkg/types"
	"github.com/h8ck3r/gscan/pkg/util"
)

var (
	scan types.Scan
	scanner types.Scanner
	err error
	hostResults []*types.HostResult
)

func main() {
	cli.Parse()
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
	util.Summarize(hostResults)
}
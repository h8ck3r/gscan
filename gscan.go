package main

import (
	"github.com/h8ck3r/gscan/internal/cli"
	"github.com/h8ck3r/gscan/internal/log"
	"github.com/h8ck3r/gscan/pkg/types"
	"github.com/h8ck3r/gscan/pkg/util"
	"os"
	"runtime/pprof"
	"runtime/trace"
	"time"
)

var (
	scan types.Scan
	scanner types.Scanner
	err error
	hostResults []*types.HostResult
)

func main() {
	startTime := time.Now()
	cli.Parse()
	if *cli.GetDebug() {
		outputFile, err := os.Create("trace.out")
		if os.IsExist(err) {
			_, _ = outputFile.Write([]byte{})
		} else if err != nil {
			log.Fatal(err)
		}
		if err := trace.Start(outputFile); err != nil {
			log.Fatal(err)
		}
		defer func() {
			trace.Stop()
			_ = outputFile.Close()
		}()

		pprofFile, err := os.Create("pprof.out")
		if os.IsExist(err) {
			_, _ = pprofFile.Write([]byte{})
		} else if err != nil {
			log.Fatal(err)
		}
		if err := pprof.StartCPUProfile(pprofFile); err != nil {
			log.Fatal(err)
		}
		defer func() {
			pprof.StopCPUProfile()
			_ = pprofFile.Close()
		}()
	}

	scan.Targets, err = cli.GetTargets()
	if err != nil {
		log.Fatal(err)
	}
	scan.Protocol = cli.GetProtocol()
	scan.Timeout = cli.GetTimeout()
	scan.GoroutineCap = cli.GetGoroutineCap()
	scan.Ports = cli.GetPorts()
	scan.Debug = cli.GetDebug()
	scan.Verbose = cli.GetVerbose()

	scanner = &scan
	hostResults, err = scanner.Scan()
	if err != nil {
		log.Error(err)
	}
	util.Summarize(hostResults, cli.GetVerbose())
	log.Printf("\nScan finished after: %v\n", time.Since(startTime))
}
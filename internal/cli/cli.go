package cli

import (
	"flag"
	"github.com/h8ck3r/gscan/internal/log"
	"github.com/h8ck3r/gscan/pkg/types"
	"github.com/h8ck3r/gscan/pkg/util"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func Parse() {
	flag.Bool("verbose", false, "verbose output")
	flag.Int("cap", 0, "maximum amount of concurrent goroutines")
	flag.Duration("timeout", time.Millisecond * 500, "maximum time to wait for connection response")
	flag.String("protocol", string(types.TCP), "protocol to use during scan")
	flag.String("ports", "80", "ports to scan")
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatalf("%s takes exactly one argument, %d provided\n", os.Args[0], flag.NArg())
	}
}

func GetTargets() ([]*types.Target, error) {
	return util.GetTargets(flag.Arg(0))
}

func GetVerbose() *bool {
	v := flag.Lookup("verbose").Value.(flag.Getter).Get().(bool)
	return &v
}

func GetTimeout() *time.Duration {
	t := flag.Lookup("timeout").Value.(flag.Getter).Get().(time.Duration)
	return &t
}

func GetGoroutineCap() *int {
	c := flag.Lookup("cap").Value.(flag.Getter).Get().(int)
	return &c
}

func GetProtocol() *types.Protocol {
	p := types.Protocol(flag.Lookup("protocol").Value.(flag.Getter).Get().(string))
	return &p
}

func GetPorts() []*types.Port {
	var ports []*types.Port

	p := flag.Lookup("ports").Value.(flag.Getter).Get().(string)
	if regexp.MustCompile(`^[0-9]+$`).MatchString(p) {
		port, err := strconv.Atoi(p)
		if err != nil {
			log.Fatal(err)
		}
		returnPort := types.Port(port)
		ports = append(ports, &returnPort)
		return ports
	} else if regexp.MustCompile(`^[0-9]{1,5}-[0-9]{1,5}$`).MatchString(p) {
		rangePorts := strings.Split(p, "-")

		startPort, err := strconv.Atoi(rangePorts[0])
		if err != nil {
			log.Fatal(err)
		}
		endPort, err := strconv.Atoi(rangePorts[1])
		if err != nil {
			log.Fatal(err)
		}

		for i := startPort; i <= endPort; i++ {
			port := types.Port(i)
			ports = append(ports, &port)
		}
		return ports
	} else {
		log.Fatalf("invalid port format: %v\n", p)
	}

	return ports
}
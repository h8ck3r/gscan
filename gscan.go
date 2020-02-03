package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"strings"
	"sync"
	"syscall"

	"github.com/pkg/errors"
)

var (
	verbose   = false
	hosts     []string
	ports     []int
	logger    = log.New(os.Stdout, "", 0)

	hostResults []*HostResult
	hostWaitGroup sync.WaitGroup
)

const (
	Closed   PortState = "closed"
	Open     PortState = "open"
)

type HostResult struct {
	Host        string
	Ports       []PortResult
	OpenPorts   []int
	ClosedPorts []int
}

type PortState string

type PortResult struct {
	Port  int
	State PortState
}

func init() {
	logger.SetPrefix("")
}

func generateHostList(hostArg string) ([]string, error) {
	if strings.Contains(hostArg, ",") {
		return strings.Split(hostArg, ","), nil
	} else if strings.Contains(hostArg, "-") {
		return nil, errors.Errorf("IP range definitions are not yet supported")
	} else {
		addr := net.ParseIP(hostArg)
		if addr == nil {
			hostname, err := net.LookupHost(hostArg)
			if err != nil {
				return nil, err
			} else if hostname == nil {
				return nil, errors.Errorf("failed to receive IP addresses or hostnames for %s", hostArg)
			} else {
				return hostname, nil
			}
		} else {
			return []string{addr.String()}, nil
		}
	}
}

func validateArgs() {
	if len(os.Args) < 2 {
		_, _ = syscall.Write(2, []byte("Please specify at least one IP to scan\n"))
		os.Exit(1)
	}

	for _, arg := range os.Args[1:] {
		if arg == "-v" || arg == "--verbose" {
			verbose = true
		} else {
			hostList, err := generateHostList(arg)
			if err != nil {
				panic(err)
			}
			if hostList != nil {
				hosts = append(hosts, hostList...)
			}
		}
	}
}

func portWorker(host string, port int, portResultChan chan *PortResult, portWaitGroup *sync.WaitGroup) {
	address := fmt.Sprintf(host+":%d", port)
	conn, err := net.Dial("tcp", address)
	var portResult PortResult

	if err != nil {
		portResult.Port = port
		portResult.State = Closed
	} else {
		portResult.Port = port
		portResult.State = Open
	}
	if conn != nil {
		_ = conn.Close()
	}

	portResultChan <- &portResult
	portWaitGroup.Done()
}

func hostWorker(host string, ports []int, hostResultChan chan *HostResult, hostWaitGroup *sync.WaitGroup, hostMutex *sync.Mutex) {
	logger.Printf("Initializing scan for %s\n", host)
	portResultChan := make(chan *PortResult)
	var portWaitGroup sync.WaitGroup
	var hostResult HostResult

	portWaitGroup.Add(len(ports))
	for _, port := range ports {
		go portWorker(host, port, portResultChan, &portWaitGroup)
	}

	for range ports {
		select {
		case portResult := <-portResultChan:
			hostMutex.Lock()
			if portResult.Port != 0 {
				hostResult.OpenPorts = append(hostResult.OpenPorts, portResult.Port)
			} else {
				hostResult.ClosedPorts = append(hostResult.ClosedPorts, portResult.Port)
			}
			hostMutex.Unlock()
		}
	}
	portWaitGroup.Wait()
	hostResultChan <- &hostResult
	close(portResultChan)
	hostWaitGroup.Done()
}

func summary(hostResults []*HostResult) {

	for _, result := range hostResults {
		sort.Ints(result.OpenPorts)
		sort.Ints(result.ClosedPorts)
		for _, port := range result.Ports {
			if port.State == Open {
				logger.Printf("Discovered open port %d on %s\n", port.Port, result.Host)
			} else if port.State == Closed && verbose {
				logger.Printf("Port %d on is %s\n", port.Port, result.Host)
			}
		}
	}
}

func main() {
	validateArgs()

	hostResultChan := make(chan *HostResult)

	var hostMutex sync.Mutex
	var hostResultMutex sync.Mutex

	logger.Println("starting scan...")

	for i := 1; i <= 1024; i++ {
		ports = append(ports, i)
	}

	for _, host := range hosts {
		hostWaitGroup.Add(1)
		go hostWorker(host, ports, hostResultChan, &hostWaitGroup, &hostMutex)
	}

	for range hosts {
		go func() {
			hostResult := <-hostResultChan
			hostResultMutex.Lock()
			hostResults = append(hostResults, hostResult)
			hostResultMutex.Unlock()
		}()
	}
	hostWaitGroup.Wait()
	summary(hostResults)

	close(hostResultChan)

	logger.Printf("\nscan done.\n")
}

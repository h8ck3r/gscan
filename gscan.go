package main

import (
	"fmt"
	"net"
	"os"
	"sort"
	"strings"
	"sync"
	"syscall"

	"github.com/pkg/errors"
)

var (
	verbose = false
	hosts   []string
	hostWaitGroup sync.WaitGroup
)

const (
	Closed   PortState = "closed"
	Open     PortState = "open"
	Filtered PortState = "filtered"
)

type HostResult struct {
	Host string
	Ports []PortResult
	OpenPorts []int
	ClosedPorts []int
}

type PortState string

type PortResult struct {
	Port  int
	State PortState
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
			}
		} else {
			return []string{addr.String()}, nil
		}
	}
	return nil, errors.Errorf("an unknown error occurred")
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

func portWorkers(host string, portResultsChan chan *PortResult, portsChan chan int, portWaitGroup *sync.WaitGroup) {
	port := <-portsChan
	address := fmt.Sprintf(host + ":%d", port)
	conn, err := net.Dial("tcp", address)

	if err != nil || conn == nil {
		portResultsChan <- &PortResult{
			Port:  port,
			State: Closed,
		}
	} else {
		conn.Close()
		portResultsChan <- &PortResult{
			Port: port,
			State: Open,
		}
	}

	portWaitGroup.Done()
}

func hostWorkers(hostsChan chan string, hostResultsChan chan *HostResult, portResultsChan chan *PortResult, portsChan chan int, hostWaitGroup *sync.WaitGroup) {
	host := <-hostsChan
	var portWaitGroup sync.WaitGroup
	var hostResult = &HostResult{}

	for i:=1; i <= cap(portsChan); i++ {
		portWaitGroup.Add(1)
		go portWorkers(host, portResultsChan, portsChan, &portWaitGroup)
	}

	for i := 1; i <= cap(portsChan); i++ {
		portsChan <- i
		portResult := <-portResultsChan
		if portResult.Port != 0 {
			hostResult.OpenPorts = append(hostResult.OpenPorts, portResult.Port)
		} else {
			hostResult.ClosedPorts = append(hostResult.ClosedPorts, portResult.Port)
		}
	}

	portWaitGroup.Wait()
	hostResultsChan <- hostResult
	hostWaitGroup.Done()
}

func summary(hostResults []*HostResult) {
	for _, result := range hostResults {
		sort.Ints(result.OpenPorts)
		sort.Ints(result.ClosedPorts)
		for _, port := range result.Ports {
			if port.State == Open {
				fmt.Printf("Discovered open port %d on %s\n", port.Port, result.Host)
			} else if port.State == Closed && verbose {
				fmt.Printf("Port %d on is %s\n", port.Port, result.Host)
			}
		}
	}
}

func main() {
	validateArgs()

	hostsChan := make(chan string, len(hosts))
	portsChan := make(chan int, 1024)
	hostResultsChan := make(chan *HostResult)
	portResultsChan := make(chan *PortResult)

	var hostResults []*HostResult

	fmt.Println("starting scan...")

	for i:=0; i < cap(hostsChan); i++ {
		hostWaitGroup.Add(1)
		fmt.Printf("Initializing scan for %s\n", hosts[i])
		go hostWorkers(hostsChan, hostResultsChan, portResultsChan, portsChan, &hostWaitGroup)
	}

	for _, host := range hosts {
		hostsChan <- host
		hostResults = append(hostResults, <-hostResultsChan)
	}

	hostWaitGroup.Wait()
	summary(hostResults)

	close(hostsChan)
	close(portsChan)
	close(hostResultsChan)
	close(portResultsChan)

	fmt.Println("scan done.")
}
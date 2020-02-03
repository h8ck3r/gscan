package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"strings"
	"sync"
	

	"github.com/pkg/errors"
)

var (
	verbose = false
	hosts   []string
	ports   []int
	logger  = log.New(os.Stdout, "", 0)
	errLogger  = log.New(os.Stderr, "", 1)
	hostResults   []*HostResult
	hostWaitGroup sync.WaitGroup
	hostArg string
)

const (
	Closed PortState = "closed"
	Open   PortState = "open"
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

func generateHostList(arg string) ([]string, error) {
	if strings.Contains(arg, ",") {
		return strings.Split(arg, ","), nil
	} else if strings.Contains(arg, "-") {
		return nil, errors.Errorf("IP range definitions are not yet supported")
	} else if strings.Contains(arg, "/") {
		return getHostsForSubnet(arg)
	} else {
		addr := net.ParseIP(arg)
		if addr == nil {
			hostname, err := net.LookupHost(arg)
			if err != nil {
				return nil, err
			} else if hostname == nil {
				return nil, errors.Errorf("failed to receive IP addresses or hostnames for %s", arg)
			} else {
				return hostname, nil
			}
		} else {
			return []string{addr.String()}, nil
		}
	}
}

func validateArgs() {
	flag.Parse()
	if flag.NArg() < 1 {
		errLogger.Fatalf("Please specify at least one IP to scan\n")
	}

	for _, arg := range os.Args[1:] {
		if arg == "-v" || arg == "--verbose" {
			verbose = true
		} else {
			hostArg = arg
			hostList, err := generateHostList(arg)
			if err != nil {
				errLogger.Fatal(err)
			}
			if hostList != nil {
				hosts = append(hosts, hostList...)
			}
			sort.Strings(hosts)
		}
	}
}

func getHostsForSubnet(network string) ([]string, error) {
	var ipList []string

	ip, ipnet, err := net.ParseCIDR(network)
	if err != nil {
		return ipList, err
	}
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); func(ip net.IP) net.IP {
		for j := len(ip)-1; j>=0; j-- {
			ip[j]++
			if ip[j] > 0 {
				break
			}
		}
		return ip
	}(ip) {
		ipList = append(ipList, ip.String())
	}
	return ipList, nil
}

func portWorker(host string, port int, portResultChan chan *PortResult, portWaitGroup *sync.WaitGroup) {
	address := fmt.Sprintf(host+":%d", port)
	portResult := PortResult{}
	conn, err := net.Dial("tcp", address)
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

func hostWorker(host string, ports []int, hostResultChan chan *HostResult, hostWaitGroup *sync.WaitGroup) {
	portResultChan := make(chan *PortResult)
	var portWaitGroup sync.WaitGroup
	var hostResult HostResult

	hostResult.Host = host

	for _, port := range ports {
		portWaitGroup.Add(1)
		go portWorker(host, port, portResultChan, &portWaitGroup)
	}

	for range ports {
		portResult := <-portResultChan
		hostResult.Ports = append(hostResult.Ports, *portResult)
		if portResult.State == Open {
			hostResult.OpenPorts = append(hostResult.OpenPorts, portResult.Port)
		} else {
			hostResult.ClosedPorts = append(hostResult.ClosedPorts, portResult.Port)
		}
	}
	portWaitGroup.Wait()
	hostResultChan <- &hostResult
	close(portResultChan)
	hostWaitGroup.Done()
}

func summary(result *HostResult) {
	sort.Ints(result.OpenPorts)
	sort.Ints(result.ClosedPorts)
	for _, port := range result.OpenPorts {
		logger.Printf("Discovered open port %d on %s\n", port, result.Host)
	}
	if verbose {
		for _, port := range result.ClosedPorts {
			logger.Printf("Port %d on is %s\n", port, result.Host)
		}
	}
}

func main() {
	validateArgs()

	hostResultChan := make(chan *HostResult)

	logger.Println("starting scan...")

	for i := 1; i <= 100; i++ {
		ports = append(ports, i)
	}
	logger.Printf("Initializing scan for %s\n", hostArg)
	for _, host := range hosts {
		hostWaitGroup.Add(1)
		go hostWorker(host, ports, hostResultChan, &hostWaitGroup)
	}

	for range hosts {
		summary(<-hostResultChan)
	}
	hostWaitGroup.Wait()

	close(hostResultChan)

	logger.Printf("\nscan done.\n")
}

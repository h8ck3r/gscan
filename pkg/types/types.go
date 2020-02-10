package types

import (
	"fmt"
	"github.com/h8ck3r/gscan/internal/log"
	"net"
	"time"
)

type Port int

func (p *Port) Int() int {
	return int(*p)
}

type PortState string

type Target string

func (t *Target) String() string {
	return string(*t)
}

type Protocol string

const (
	TCP Protocol = "tcp"
	UDP Protocol = "udp"
	Open PortState = "open"
	Closed PortState = "closed"
	Filtered PortState = "filtered"
)

type Scan struct {
	Targets []*Target
	Ports []*Port
	Protocol *Protocol
	Timeout *time.Duration
	GoroutineCap *int
	Debug *bool
	Verbose *bool
}

type Scanner interface {
	Scan() ([]*HostResult, error)
}

func (s *Scan) Scan() ([]*HostResult, error) {
	return Execute(s)
}

type HostResult struct {
	Host *Target
	PortResults []*PortResult
}

type PortResult struct {
	Target *Target
	Port *Port
	Protocol *Protocol
	State PortState
}

type HostResultChan chan *HostResult
type PortResultChan chan *PortResult

func scanTarget(target *Target, scan *Scan, hostResultChan chan *HostResult) {
	var startTime time.Time
	var portResults []*PortResult
	var hostResult HostResult

	portResultChan := make(chan *PortResult, len(scan.Ports))

	if *scan.Debug {
		startTime = time.Now()
	}

	hostResult.Host = target

	for i := 0; i < cap(portResultChan); i++ {
		portResults = append(portResults, scanPort(scan.Protocol, target, scan.Ports[i], scan.Timeout, scan))
	}

	/*for i := 0; i < cap(portResultChan); i++ {
		select {
		case portResult := <- portResultChan:
			portResults = append(portResults, portResult)
		}
	}*/

	hostResult.PortResults = portResults
	if *scan.Debug {
		log.Printf("\nScanned %s within: %v\n", *target, time.Since(startTime))
	}
	hostResultChan <- &hostResult
}

func scanPort(protocol *Protocol, target *Target, port *Port, timeout *time.Duration, scan *Scan) *PortResult {
	var startTime time.Time
	var portResult PortResult

	if *scan.Debug && *scan.Verbose {
		startTime = time.Now()
	}

	address := fmt.Sprintf("%s:%d", target.String(), port.Int())
	portResult.Protocol = scan.Protocol
	portResult.Target = target
	portResult.Port = port
	_, err := net.DialTimeout(string(*protocol), address, *timeout)
	if err != nil {
		portResult.State = Closed
	} else {
		log.Println("OPEN")
		portResult.State = Open
		//_ = conn.Close()
	}
	if *scan.Debug && *scan.Verbose {
		log.Printf("\nScanned %v within: %v\n", address, time.Since(startTime))
	}
	return &portResult
}

func Execute(scan *Scan) ([]*HostResult, error) {
	var hostResults []*HostResult
	hostResultChan := make(chan *HostResult, len(scan.Targets))

	for i := 0; i < cap(hostResultChan); i++ {
		go scanTarget(scan.Targets[i], scan, hostResultChan)
	}

	for i := 0; i < cap(hostResultChan); i++ {
		select {
		case hostResult := <- hostResultChan:
			hostResults = append(hostResults, hostResult)
		}
	}

	return hostResults, nil
}
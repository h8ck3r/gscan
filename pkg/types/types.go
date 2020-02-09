package types

import (
	"fmt"
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
	portResultChan := make(chan *PortResult, len(scan.Ports))
	var portResults []*PortResult
	var hostResult HostResult
	hostResult.Host = target
	for i := 0; i < cap(portResultChan); i++ {
		scanPort(scan.Protocol, target, scan.Ports[i], scan.Timeout, scan, portResultChan)
	}
	for i := 0; i < cap(portResultChan); i++ {
		select {
		case portResult := <- portResultChan:
			portResults = append(portResults, portResult)
		}
	}
	hostResult.PortResults = portResults
	hostResultChan <- &hostResult
}

func scanPort(protocol *Protocol, target *Target, port *Port, timeout *time.Duration, scan *Scan, portResultChan chan *PortResult) {
	address := fmt.Sprintf("%s:%d", target.String(), port.Int())
	var portResult PortResult
	portResult.Protocol = scan.Protocol
	portResult.Target = target
	portResult.Port = port
	conn, err := net.DialTimeout(string(*protocol), address, *timeout)
	if err != nil {
		portResult.State = Closed
	} else {
		portResult.State = Open
		_ = conn.Close()
	}
	portResultChan <- &portResult
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
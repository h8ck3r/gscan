package types

import "time"

type Port int

func (p *Port) Int() int {
	return int(*p)
}

type PortState string

type Host string

func (h *Host) String() string {
	return string(*h)
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
	Hosts []Host
	Ports []Port
	Protocol Protocol
	Timeout time.Duration
	GoroutineCap int
}

type Scanner interface {
	scan() error
}

func (s *Scan) scan() error {
	return nil
}

type HostResult struct {
	Host Host
	PortResults []*PortResult
}

type PortResult struct {
	Host Host
	Port Port
	Protocol Protocol
	State PortState
}

type HostResultChan chan *HostResult
type PortResultChan chan *PortResult
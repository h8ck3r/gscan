package main

import (
	"github.com/h8ck3r/gscan/internal/cli"
	"github.com/h8ck3r/gscan/internal/log"
	"github.com/h8ck3r/gscan/pkg/types"
)

var (
	scan types.Scan
	scanner types.Scanner
	ports []types.Port
	verbose *bool
)

func main() {
	cli.Parse()
	verbose = cli.GetVerbose()
	var err error
	scan.Targets, err = cli.GetTargets()
	if err != nil {
		log.Fatal(err)
	}

	scan.Protocol = cli.GetProtocol()
	scan.Timeout = cli.GetTimeout()
	scan.GoroutineCap = cli.GetGoroutineCap()

	for i := 1; i <= 100; i++ {
		ports = append(ports, types.Port(i))
	}
	scan.Ports = cli.GetPorts()
	scanner = &scan
	hostResults, err := scanner.Scan()
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

/*
var (
	verbose = false
	hosts   []string
	ports   []int
	goroutineCap int
	timeout time.Duration
	logger  = log.New(os.Stdout, "", 0)
	errLogger  = log.New(os.Stderr, "", 1)
	hostArg string
)

const (
	Closed PortState = "closed"
	Open   PortState = "open"
)

func init() {
	logger.SetPrefix("")
	flag.BoolVar(&verbose, "verbose", false, "verbose output")
	flag.IntVar(&goroutineCap, "cap", 0, "the maximum number of parallel goroutines (0 is infinite)")
	flag.DurationVar(&timeout, "timeout", 8000, "connection timeout (in milliseconds)")
}

type Scan struct {
	Hosts []string
	Ports []int
	Protocol string
	Timeout time.Duration
	GoroutineCap int
	WaitGroup *sync.WaitGroup
	Results []HostResult
}

type Scanner interface {
	scan() error
}

type HostResult struct {
	Host        string
	Ports       []*PortResult
}

type PortState string

type PortResult struct {
	Port  int
	State PortState
}

func PortWorker(host string, port int, protocol string, timeout time.Duration, waitGroup *sync.WaitGroup, portResultChan chan *PortResult) {
	var portResult PortResult
	portResult.Port = port
	conn, err := net.DialTimeout(protocol, fmt.Sprintf("%s:%d", host, port), timeout)
	if err == nil {
		portResult.State = Open
		_ = conn.Close()
	} else {
		portResult.State = Closed
	}
	portResultChan <- &portResult
	waitGroup.Done()
}

func HostWorker(host string, ports []int, protocol string, timeout time.Duration, goroutineCap *int, waitGroup *sync.WaitGroup, hostResultChan chan *HostResult) {
	var portWaitGroup sync.WaitGroup
	var portResults []*PortResult
	var hostResult HostResult

	var goroutineStack = 0

	portResultChan := make(chan *PortResult, len(ports))

	for _, port := range ports {
		if *goroutineCap != 0 && goroutineStack % *goroutineCap == 0 {
			portWaitGroup.Wait()
		}
		portWaitGroup.Add(1)
		go PortWorker(host, port, protocol, timeout, &portWaitGroup, portResultChan)
	}
	for i := 0; i < cap(portResultChan); i++ {
		select {
		case portResult := <-portResultChan:
			portResults = append(portResults, portResult)
		}
	}
	portWaitGroup.Wait()
	hostResult.Ports = portResults
	hostResult.Host = host
	hostResultChan <- &hostResult
	close(portResultChan)
	waitGroup.Done()
}

func (s *Scan) scan() error {
	hostResultChan := make(chan *HostResult, len(hosts))

	for _, host := range s.Hosts {
		s.WaitGroup.Add(1)
		go HostWorker(host, s.Ports, s.Protocol, s.Timeout, &s.GoroutineCap, s.WaitGroup, hostResultChan)
	}
	for i := 0; i < cap(hostResultChan); i++ {
		select {
		case hostResult := <-hostResultChan:
			s.Results = append(s.Results, *hostResult)
		}
	}
	s.WaitGroup.Wait()
	close(hostResultChan)
	return nil
}

func generateHostList(arg string) ([]string, error) {
	if strings.Contains(arg, ",") {
		hostStrings := strings.Split(arg, ",")
		for _, host := range hostStrings {
			hosts = append(hosts, host)
		}
	} else if regexp.MustCompile(`^([0-9]{1,3}(-[0-9]{1,3})?\.){3}[0-9]{1,3}(-[0-9]{1,3})?$`).MatchString(arg) {
		return nil, errors.Errorf("IP range definitions are not yet supported")
	} else if strings.Contains(arg, "/") {
		return getHostsForSubnet(arg)
	} else {
		addr := net.ParseIP(arg)
		if addr == nil {
			hostnames, err := net.LookupHost(arg)
			if err != nil {
				return nil, err
			} else if hostnames == nil {
				return nil, errors.Errorf("failed to receive IP addresses or hostnames for %s", arg)
			} else {
				return hostnames, nil
			}
		} else {
			return []string{addr.String()}, nil
		}
	}
	return nil, nil
}

func validateArgs() {
	flag.Parse()
	if flag.NArg() < 1 {
		errLogger.Fatalf("Please specify at least one IP to scan\n")
	}

	if goroutineCap < 0 {
		errLogger.Fatalf("Please set the goroutine cap to a value greater than 0, or zero for infinite\n")
	}

	for _, arg := range os.Args[1:] {
		if arg == "-verbose" || arg == "-timeout" || arg == "-cap" {
			continue
		} else {
			hostArg = arg
			hostList, err := generateHostList(arg)
			if err != nil {
				errLogger.Fatal(err)
			}
			var hostStrings []string
			if hostList != nil {
				for _, host := range hostList {
					hostStrings = append(hostStrings, host)
				}
			}
			sort.Strings(hostStrings)
			for _, host := range hostStrings {
				hosts = append(hosts, host)
			}
		}
	}
}

func convertToBinary(ip net.IP) string {
	bin := binary.BigEndian.Uint32(ip)
	return strconv.FormatUint(uint64(bin), 2)
}

func getBroadCastAddress(ipNet *net.IPNet) (net.IP, error) {
	gateway := convertToBinary(ipNet.IP)
	var broadcast net.IP

	mask := ipNet.Mask
	newMask := convertToBinary(net.IP(mask))
	var broadcastBuf bytes.Buffer

	for index, maskBit := range newMask {
		uintMaskBit := uint32(maskBit)
		bit := string(uintMaskBit)
		if bit == "1" {
			uintGatewayBit := uint32(gateway[index])
			broadcastBuf.WriteString(string(uintGatewayBit))
		} else {
			broadcastBuf.WriteString("1")
		}
	}

	if len(broadcastBuf.String()) != 32 {
		panic("invalid ip size")
	} else {
		var broadcastBytes []string
		broadcastBytes = append(broadcastBytes, broadcastBuf.String()[0:8])
		broadcastBytes = append(broadcastBytes, broadcastBuf.String()[8:16])
		broadcastBytes = append(broadcastBytes, broadcastBuf.String()[16:24])
		broadcastBytes = append(broadcastBytes, broadcastBuf.String()[24:])

		ip1, err := strconv.ParseInt(broadcastBytes[0], 2, 64)
		if err != nil {
			panic(err)
		}

		ip2, err := strconv.ParseInt(broadcastBytes[1], 2, 64)
		if err != nil {
			panic(err)
		}

		ip3, err := strconv.ParseInt(broadcastBytes[2], 2, 64)
		if err != nil {
			panic(err)
		}

		ip4, err := strconv.ParseInt(broadcastBytes[3], 2, 64)
		if err != nil {
			panic(err)
		}
		broadcast = net.IPv4(byte(ip1), byte(ip2), byte(ip3), byte(ip4))
	}

	return broadcast, nil
}

func getHostsForSubnet(network string) ([]string, error) {
	address, ipNet, err := net.ParseCIDR(network)
	if err != nil {
		return nil, err
	}
	broadcast, err := getBroadCastAddress(ipNet)
	if err != nil {
		panic(err)
	}

	var ipAddresses []string
	for ip := address.Mask(ipNet.Mask); ipNet.Contains(ip); func(ip net.IP){
		for j := len(ip) - 1; j >= 0; j-- {
			ip[j]++
			if ip[j] > 0 {
				break
			}
		}
	}(ip) {
		if ip.String() != ipNet.IP.String() && ip.String() != broadcast.String() {
			ipAddresses = append(ipAddresses, ip.String())
		}
	}

	return ipAddresses, nil
}

func summary(hostResults []HostResult) {
	for _, hostResult := range hostResults {
		for _, portResult := range hostResult.Ports {
			if portResult.State == Open {
				logger.Printf("Discovered open port %d on %s\n", portResult.Port, hostResult.Host)
			} else if portResult.State == Closed && verbose {
				logger.Printf("Port %d on %s is %s\n", portResult.Port, hostResult.Host, Closed)
			}
		}
	}
}

func main() {
	validateArgs()

	for i := 1; i <= 100; i++ {
		ports = append(ports, i)
	}
	logger.Printf("Initializing scan for %s\n", hostArg)

	var protocol = "tcp"
	var scanWaitGroup sync.WaitGroup

	sort.Strings(hosts)

	scan := Scan{
		Hosts:        hosts,
		Ports:        ports,
		Protocol:     protocol,
		Timeout:      timeout,
		GoroutineCap: goroutineCap,
		WaitGroup:    &scanWaitGroup,
	}

	var scanner Scanner = &scan

	if err := scanner.scan(); err != nil {
		panic(err)
	}
	summary(scan.Results)

	logger.Printf("\nscan done.\n")
}
*/

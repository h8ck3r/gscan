package util

import (
	"bytes"
	"encoding/binary"
	"github.com/pkg/errors"
	"net"
	"regexp"
	"strconv"
)

func GetTargets(argument string) ([]string, error) {
	var targets []string

	if regexp.MustCompile(`^([0-9]{1,3}\.){3}[0-9]{1,3}$`).MatchString(argument) {
		targets = []string{argument}
		return targets, nil
	} else if regexp.MustCompile(`^([0-9]{1,3}(-[0-9]{1,3})?\.){3}[0-9]{1,3}(-[0-9]{1,3})?$`).MatchString(argument) {
		_, _ = getHostsForIPRange(argument)
		return targets, errors.Errorf("ip range definitions are not yet supported")
	} else if regexp.MustCompile(`^([0-9]{1,3}\.){3}[0-9]{1,3}/[0-9]{1,2}$`).MatchString(argument) {
		return getHostsForSubnet(argument)
	} else {
		return targets, errors.Errorf("invalid argument: %s\n", argument)
	}
}

func getHostsForIPRange(network string) ([]string, error) {
	network = ""
	return []string{network}, nil
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
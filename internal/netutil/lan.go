package netutil

import (
	"net"
	"sort"
	"time"
)

func LANIPv4s() []string {
	seen := make(map[string]struct{})
	var out []string
	add := func(ip net.IP) {
		ip = ip.To4()
		if ip == nil || !ip.IsPrivate() {
			return
		}
		text := ip.String()
		if _, ok := seen[text]; ok {
			return
		}
		seen[text] = struct{}{}
		out = append(out, text)
	}

	ifaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range ifaces {
			if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
				continue
			}
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}
			addAddrList(addrs, add)
		}
	}

	if addrs, err := net.InterfaceAddrs(); err == nil {
		addAddrList(addrs, add)
	}

	for _, ip := range routedIPv4s() {
		add(ip)
	}

	sort.Strings(out)
	return out
}

func addAddrList(addrs []net.Addr, add func(net.IP)) {
	for _, addr := range addrs {
		switch v := addr.(type) {
		case *net.IPNet:
			add(v.IP)
		case *net.IPAddr:
			add(v.IP)
		}
	}
}

func routedIPv4s() []net.IP {
	targets := []string{
		"223.5.5.5:53",
		"8.8.8.8:53",
		"192.168.1.1:9",
		"10.0.0.1:9",
		"172.16.0.1:9",
	}
	var out []net.IP
	dialer := net.Dialer{Timeout: 200 * time.Millisecond}
	for _, target := range targets {
		conn, err := dialer.Dial("udp4", target)
		if err != nil {
			continue
		}
		if addr, ok := conn.LocalAddr().(*net.UDPAddr); ok {
			out = append(out, addr.IP)
		}
		_ = conn.Close()
	}
	return out
}

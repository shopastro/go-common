package net

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

func HostPort(addr string, port interface{}) string {
	host := addr
	if strings.Count(addr, ":") > 0 {
		host = fmt.Sprintf("[%s]", addr)
	}

	if v, ok := port.(string); ok && v == "" {
		return host
	} else if v, ok := port.(int); ok && v == 0 && net.ParseIP(host) == nil {
		return host
	}

	return fmt.Sprintf("%s:%v", host, port)
}

func Listen(addr string, fn func(string) (net.Listener, error)) (net.Listener, error) {
	if strings.Count(addr, ":") == 1 && strings.Count(addr, "-") == 0 {
		return fn(addr)
	}

	host, ports, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	prange := strings.Split(ports, "-")

	if len(prange) < 2 {
		return fn(addr)
	}

	min, err := strconv.Atoi(prange[0])
	if err != nil {
		return nil, errors.New("unable to extract port range")
	}

	max, err := strconv.Atoi(prange[1])
	if err != nil {
		return nil, errors.New("unable to extract port range")
	}

	for port := min; port <= max; port++ {
		ln, err := fn(HostPort(host, port))
		if err == nil {
			return ln, nil
		}

		if port == max {
			return nil, err
		}
	}

	return nil, fmt.Errorf("unable to bind to %s", addr)
}

// NewEndpoint new an Endpoint URL.
func NewEndpoint(scheme, host string, isSecure bool) *url.URL {
	var query string
	if isSecure {
		query = "isSecure=true"
	}
	return &url.URL{Scheme: scheme, Host: host, RawQuery: query}
}

// Extract returns a private addr and port.
func Extract(hostPort string, lis net.Listener) (string, error) {
	addr, port, err := net.SplitHostPort(hostPort)
	if err != nil && lis == nil {
		return "", err
	}
	if lis != nil {
		if p, ok := Port(lis); ok {
			port = strconv.Itoa(p)
		} else {
			return "", fmt.Errorf("failed to extract port: %v", lis.Addr())
		}
	}
	if len(addr) > 0 && (addr != "0.0.0.0" && addr != "[::]" && addr != "::") {
		return net.JoinHostPort(addr, port), nil
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	lowest := int(^uint(0) >> 1)
	var result net.IP
	for _, iface := range ifaces {
		if (iface.Flags & net.FlagUp) == 0 {
			continue
		}
		if iface.Index < lowest || result == nil {
			lowest = iface.Index
		} else if result != nil {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, rawAddr := range addrs {
			var ip net.IP
			switch addr := rawAddr.(type) {
			case *net.IPAddr:
				ip = addr.IP
			case *net.IPNet:
				ip = addr.IP
			default:
				continue
			}
			if isValidIP(ip.String()) {
				result = ip
			}
		}
	}
	if result != nil {
		return net.JoinHostPort(result.String(), port), nil
	}
	return "", nil
}

func isValidIP(addr string) bool {
	ip := net.ParseIP(addr)
	return ip.IsGlobalUnicast() && !ip.IsInterfaceLocalMulticast()
}

// Port return a real port.
func Port(lis net.Listener) (int, bool) {
	if addr, ok := lis.Addr().(*net.TCPAddr); ok {
		return addr.Port, true
	}
	return 0, false
}

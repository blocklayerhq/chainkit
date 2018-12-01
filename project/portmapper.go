package project

import (
	"errors"
	"fmt"
	"net"

	"github.com/blocklayerhq/chainkit/ui"
)

const (
	// minPort is the minimum port that will be used
	minPort = 42000
	// maxPort is the maximum port that will be used
	maxPort = 60000
	// numPorts is the number of ports that will be used
	numPorts = 4
	// portStep is the step between port ranges
	portStep = 10
)

var (
	// ErrPortsUnavailable is returned when no ports can be found.
	ErrPortsUnavailable = errors.New("unable to allocate ports")
)

// PortMapper holds port configuration.
type PortMapper struct {
	Explorer      int
	TendermintRPC int
	TendermintP2P int
	IPFS          int
}

// AllocatePorts will allocate a set of ports
func AllocatePorts() (*PortMapper, error) {
	for port := minPort; port < maxPort; port += portStep {
		if !portRangeAvailable(port, numPorts) {
			continue
		}
		if port != minPort {
			ui.Error("Port range %d-%d not available, using %d-%d instead",
				minPort, minPort+numPorts,
				port, port+numPorts)
		}
		return &PortMapper{
			Explorer:      port + 0,
			TendermintRPC: port + 1,
			TendermintP2P: port + 2,
			IPFS:          port + 3,
		}, nil
	}

	return nil, ErrPortsUnavailable
}

func portRangeAvailable(base, n int) bool {
	// We are dialing in addition to listening because for some reason,
	// if the port is being used by a container, it will listen just fine
	// rather than throwing an address already in use.

	for i := 0; i < n; i++ {
		port := base + i

		// First, try to listen to that port.
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return false
		}
		l.Close()

		// Double check by also attempting a connection.
		c, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			c.Close()
			return false
		}
	}

	return true
}

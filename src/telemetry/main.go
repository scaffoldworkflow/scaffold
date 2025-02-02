// main.go

package main

import (
	"fmt"
	"net"
	"time"

	"scaffold/telemetry/config"
	"scaffold/telemetry/constants"

	logger "github.com/jfcarter2358/go-logger"
)

type Status struct {
	Response string `json:"response"`
	Resource string `json:"resource"`
	Output   string `json:"output"`
}

func (s Status) SendResults() {
	fmt.Printf("%v\n", s)
	return
}

// func printHop(hop traceroute.TracerouteHop) string {
// 	out := ""
// 	addr := fmt.Sprintf("%v.%v.%v.%v", hop.Address[0], hop.Address[1], hop.Address[2], hop.Address[3])
// 	hostOrAddr := addr
// 	if hop.Host != "" {
// 		hostOrAddr = hop.Host
// 	}
// 	if hop.Success {
// 		out += fmt.Sprintf("%-3d %v (%v)  %v\n", hop.TTL, hostOrAddr, addr, hop.ElapsedTime)
// 	} else {
// 		out += fmt.Sprintf("%-3d *\n", hop.TTL)
// 	}
// 	return out
// }

// func address(address [4]byte) string {
// 	return fmt.Sprintf("%v.%v.%v.%v", address[0], address[1], address[2], address[3])
// }

// func report(status Status) {

// }

// func checkConnection(endpoint string) {
// 	ipAddr, err := net.ResolveIPAddr("ip", endpoint)
// 	if err != nil {
// 		return
// 	}

// 	output := ""

// 	options := traceroute.TracerouteOptions{}
// 	options.SetRetries(0)
// 	options.SetMaxHops(traceroute.DEFAULT_MAX_HOPS + 1)
// 	options.SetFirstHop(traceroute.DEFAULT_FIRST_HOP)

// 	output += fmt.Sprintf("traceroute to %v (%v), %v hops max, %v byte packets\n", endpoint, ipAddr, options.MaxHops(), options.PacketSize())

// 	c := make(chan traceroute.TracerouteHop, 0)
// 	go func() {
// 		for {
// 			hop, ok := <-c
// 			if !ok {
// 				output += "\n"
// 				return
// 			}
// 			output += printHop(hop)
// 		}
// 	}()

// 	result, err := traceroute.Traceroute(endpoint, &options, c)
// 	if err != nil {
// 		fmt.Printf("Error: %s\n", err.Error())
// 	}
// 	fmt.Printf("Got result: %v\n", result)
// 	fmt.Printf("Got output: %s\n", output)
// }

func main() {
	config.LoadConfig()
	logger.SetLevel(logger.LOG_LEVEL_INFO)
	for _, endpoint := range config.Config.TraceRoutes {
		status := Status{
			Resource: endpoint,
		}
		logger.Infof("", "Checking endpoint: %s", endpoint)
		timeout := 1 * time.Second
		conn, err := net.DialTimeout("tcp", endpoint, timeout)
		if err != nil {
			logger.Errorf("", "Site unreachable, error: %s", err.Error())
			status.Output = fmt.Sprintf("Site unreachable: %s", err.Error())
			status.Response = constants.CONNECTION_FAILURE
			status.SendResults()
			continue
		}
		defer conn.Close()
		logger.Successf("", "%s is reachable", endpoint)
		status.Output = fmt.Sprintf("%s is reachable", endpoint)
		status.Response = constants.CONNECTION_SUCCESS
		status.SendResults()
	}
}

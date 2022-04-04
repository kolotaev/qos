package qos

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
)

// TCPAdminServer control plane server for TCPFileServers
type TCPAdminServer struct {
	throttlers map[string]*Throttler
	listener   net.Listener
	logger     *log.Logger
}

// NewTCPAdminServer TCPAdminServer ctor
func NewTCPAdminServer(throttlers map[string]*Throttler, logger *log.Logger) *TCPAdminServer {
	return &TCPAdminServer{
		throttlers: throttlers,
		logger:     logger,
	}
}

// Handle connection
func (s *TCPAdminServer) Handle(conn net.Conn) {
	defer conn.Close()
	for {
		netData, err := bufio.NewReader(conn).ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			s.logger.Println(fmt.Errorf("failed to read net data: %s", err))
			continue
		}

		cmd, err := ParseInput(string(netData))
		if err != nil {
			errorRespond(conn, err)
			continue
		}

		if cmd.IsHalt {
			textRespond(conn, "BYE!")
			break
		}
		switch cmd.Action {
		case "THROTTLE":
			err := s.enableThrottling(cmd.GetArg(0), cmd.GetArg(1))
			if err != nil {
				s.logger.Println(err)
				errorRespond(conn, err)
			} else {
				s.logger.Printf("Throttling for server enabled? `%s`", cmd.GetArg(1))
				okRespond(conn)
			}
		case "SLIMIT":
			err := s.setServerLimit(cmd.GetArg(0), cmd.GetArg(1))
			if err != nil {
				s.logger.Println(err)
				errorRespond(conn, err)
			} else {
				s.logger.Printf("Limit `%s` for server `%s` was set", cmd.GetArg(1), cmd.GetArg(0))
				okRespond(conn)
			}
		case "CLIMIT":
			err := s.setConnectionLimit(cmd.GetArg(0), cmd.GetArg(1))
			if err != nil {
				s.logger.Println(err)
				errorRespond(conn, err)
			} else {
				s.logger.Printf("Limit `%s` for connection `%s` was set", cmd.GetArg(1), cmd.GetArg(0))
				okRespond(conn)
			}
		case "CLIST":
			err := s.setConnectionLimit(cmd.GetArg(0), cmd.GetArg(1))
			if err != nil {
				s.logger.Println(err)
				errorRespond(conn, err)
			} else {
				s.logger.Printf("Limit `%s` for connection `%s` was set", cmd.GetArg(1), cmd.GetArg(0))
				okRespond(conn)
			}
		}
	}
}

// Serve service
func (s *TCPAdminServer) Serve(protocol, address string) error {
	s.logger.Printf("TCP Admin Server listens on %s %s\n", protocol, address)
	listener, err := net.Listen(protocol, address)
	if err != nil {
		s.logger.Println(err)
		return err
	}
	defer listener.Close()
	s.listener = listener

	for {
		c, err := listener.Accept()
		if err != nil {
			s.logger.Println(err)
			return err
		}
		s.logger.Printf("ADMIN SERVER: Got a new client connection %s\n", c.RemoteAddr().String())
		go s.Handle(c)
	}
}

// Stop stop listening for incoming connections.
func (s *TCPAdminServer) Stop() {
	s.logger.Println("TCP File Server stops")
	if s.listener != nil {
		s.listener.Close()
	}
}

func (s *TCPAdminServer) enableThrottling(serverName, doEnable string) error {
	if _, ok := s.throttlers[serverName]; !ok {
		return fmt.Errorf("unknown server %s", serverName)
	}

	throttler := s.throttlers[serverName]
	if doEnable == "yes" {
		throttler.Enable()
	} else {
		throttler.Disable()
	}
	return nil
}

func (s *TCPAdminServer) setServerLimit(serverName, limit string) error {
	if _, ok := s.throttlers[serverName]; !ok {
		return fmt.Errorf("unknown server %s", serverName)
	}

	lim, err := strconv.ParseInt(limit, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse limit number `%s`", limit)
	}
	s.throttlers[serverName].SetBandwidthLimit(lim)
	return nil
}

func (s *TCPAdminServer) setConnectionLimit(connectionAddress, limit string) error {
	lim, err := strconv.ParseInt(limit, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse limit number `%s`", limit)
	}
	for _, throttler := range s.throttlers {
		throttler.SetBandwidthLimitForConnection(lim, connectionAddress)
	}
	return nil
}

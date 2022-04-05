package qos

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
)

// TCPFileServer server for serving files over TCP from a base directory.
type TCPFileServer struct {
	throttler     *Throttler
	baseDirectory string
	logger        *log.Logger
}

// NewTCPFileServer TCPFileServer ctor
func NewTCPFileServer(throttler *Throttler, baseDirectory string, logger *log.Logger) *TCPFileServer {
	return &TCPFileServer{
		throttler:     throttler,
		baseDirectory: baseDirectory,
		logger:        logger,
	}
}

// Handle serve a file over a TCP connection.
func (s *TCPFileServer) Handle(conn net.Conn) {
	connectionAddress := conn.RemoteAddr().String()
	defer s.throttler.UnregisterConnection(connectionAddress)
	defer conn.Close()

	for {
		netData, err := bufio.NewReader(conn).ReadString('\n')
		if err == io.EOF {
			s.logger.Println(fmt.Errorf("client %s has left", connectionAddress))
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
		if cmd.Action == "FILE" {
			err := s.writeFile(cmd.GetArg(0), conn, connectionAddress)
			if err != nil && err != io.EOF {
				s.logger.Println(err)
				errorRespond(conn, err)
				continue
			}
		}
	}
}

// Serve listen for incoming connections and run server.
func (s *TCPFileServer) Serve(protocol, address string) error {
	s.logger.Printf("TCP File Server listens on %s %s\n", protocol, address)
	err := s.throttler.Listen(protocol, address)
	if err != nil {
		return err
	}
	defer s.throttler.Close()

	for {
		c, err := s.throttler.Accept()
		if err != nil {
			return err
		}
		s.logger.Printf("Got a new client connection %s\n", c.RemoteAddr().String())
		go s.Handle(c)
	}
}

// Stop stop listening for incoming connections.
func (s *TCPFileServer) Stop() error {
	s.logger.Println("TCP File Server stops")
	err := s.throttler.Close()
	if err != nil {
		return err
	}
	return nil
}

func (s *TCPFileServer) writeFile(fileName string, conn net.Conn, connectionKey string) error {
	filePath, err := filepath.Abs(filepath.Join(s.baseDirectory, strings.TrimSpace(fileName)))
	if err != nil {
		return err
	}
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	ctx := context.Background()
	n, err := s.throttler.Write(ctx, conn, connectionKey, file)
	if err != nil {
		return err
	}

	s.logger.Println(n, "bytes sent")
	return nil
}

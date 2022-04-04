package qos_test

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kolotaev/qos"
)

func TestEnd2End(t *testing.T) {
	// Set up:
	serverAddresses := map[string]string{
		"fs":    "127.0.0.1:55888",
		"admin": "127.0.0.1:55999",
	}
	logger := log.Default()
	throttlers := map[string]*qos.Throttler{
		"srv1": qos.NewThrottler(2, true),
	}
	fileServer := qos.NewTCPFileServer(throttlers["srv1"], "./example/files", logger)
	adminServer := qos.NewTCPAdminServer(throttlers, logger)

	// Run File Server
	go func() {
		fileServer.Serve("tcp4", serverAddresses["fs"])
	}()
	defer fileServer.Stop()

	// Run Admin Server
	go func() {
		adminServer.Serve("tcp4", serverAddresses["admin"])
	}()
	defer adminServer.Stop()

	time.Sleep(1 * time.Second)

	// Create clients
	assert.NotEqual(t, "", serverAddresses["fs"])
	fsClient, err := net.Dial("tcp4", serverAddresses["fs"])
	assert.NoError(t, err)
	fsClientReader := bufio.NewReader(fsClient)
	adminClient, err := net.Dial("tcp4", serverAddresses["admin"])
	assert.NoError(t, err)
	adminClientReader := bufio.NewReader(adminClient)

	// Test cases:

	fmt.Println("Test case #1. Send wrong command to client.")
	fsClient.Write([]byte("foobar\n"))
	res, err := fsClientReader.ReadString('\n')
	assert.NoError(t, err)
	assert.Equal(t, "Error: received unknown command: `foobar`\n", res)

	fmt.Println("Test case #2. Try to read inexistent file.")
	fsClient.Write([]byte("FILE qq.txt\n"))
	res, err = fsClientReader.ReadString('\n')
	assert.NoError(t, err)
	assert.Contains(t, res, "example/files/qq.txt: no such file or directory")

	fmt.Println("Test case #3. Try to read file within approximately 7 seconds.")
	start := time.Now()
	fsClient.Write([]byte("FILE small.txt\n"))
	res, err = fsClientReader.ReadString('.')
	stop := start.Add(time.Since(start))
	assert.NoError(t, err)
	assert.Equal(t, "Go is awesome.", res)
	assert.WithinDuration(t, start.Add(7*time.Second), stop, 1*time.Second)

	fmt.Println("Test case #4. Change bandwidth limit for a File Server.")
	adminClient.Write([]byte("SLIMIT srv1 20\n"))
	res, err = adminClientReader.ReadString('\n')
	assert.NoError(t, err)
	assert.Equal(t, "OK\n", res)

	fmt.Println("Test case #5. Try to read file with server's changed limits takes around 1 second.")
	start = time.Now()
	fsClient.Write([]byte("FILE small.txt\n"))
	res, err = fsClientReader.ReadString('.')
	stop = start.Add(time.Since(start))
	assert.NoError(t, err)
	assert.Equal(t, "Go is awesome.", res)
	assert.WithinDuration(t, start.Add(2*time.Second), stop, 1*time.Second)

	fmt.Println("Test case #6. Change bandwidth limit for a client.")
	adminClient.Write([]byte(fmt.Sprintf("CLIMIT %s 2\n", fsClient.LocalAddr().String())))
	res, err = adminClientReader.ReadString('\n')
	assert.NoError(t, err)
	assert.Equal(t, "OK\n", res)

	fmt.Println("Test case #5. Try to read file with connections's changed limits takes around 7 seconds.")
	start = time.Now()
	fsClient.Write([]byte("FILE small.txt\n"))
	res, err = fsClientReader.ReadString('.')
	stop = start.Add(time.Since(start))
	assert.NoError(t, err)
	assert.Equal(t, "Go is awesome.", res)
	assert.WithinDuration(t, start.Add(7*time.Second), stop, 1*time.Second)
}

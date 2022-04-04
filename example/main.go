package main

import (
	"log"
	"os"
	"sync"

	"github.com/kolotaev/qos"
)

func main() {
	orchestarator := new(sync.WaitGroup)
	baseDir := "./example/files"
	throttlers := map[string]*qos.Throttler{
		"srv1": qos.NewThrottler(10, true),
		"srv2": qos.NewThrottler(20, true),
	}
	fileServer1 := qos.NewTCPFileServer(throttlers["srv1"], baseDir, log.New(os.Stdout, "FILE SRV #1 ", log.LstdFlags))
	fileServer2 := qos.NewTCPFileServer(throttlers["srv2"], baseDir, log.New(os.Stdout, "FILE SRV #2 ", log.LstdFlags))
	adminServer := qos.NewTCPAdminServer(throttlers, log.New(os.Stdout, "ADMIN SRV ", log.LstdFlags))

	orchestarator.Add(1)
	go func() {
		fileServer1.Serve("tcp4", ":3000")
		orchestarator.Done()
	}()

	orchestarator.Add(1)
	go func() {
		fileServer2.Serve("tcp4", ":4000")
		orchestarator.Done()
	}()

	orchestarator.Add(1)
	go func() {
		adminServer.Serve("tcp4", ":5000")
		orchestarator.Done()
	}()

	orchestarator.Wait()
}

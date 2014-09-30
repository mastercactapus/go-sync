package main

import (
	"fmt"
	"github.com/alecthomas/kingpin"
	"os"
	"runtime"
	"sync"
)

var (
	app          = kingpin.New("gosync", "A tool to sync a directory over a high-speed network.")
	mainPort     = app.Flag("port", "port number to listen or connect to when syncing").Short('p').Default("32011").Uint64()
	bufferSize   = app.Flag("buffer", "The size (in kB) of the buffer for sending/receiving.").Short('b').Default("131072").Int()
	hostCommand  = app.Command("host", "Make a directory available on the network.")
	hostPath     = hostCommand.Arg("path", "The directory to host.").Default(".").ExistingDir()
	hostHttp     = hostCommand.Flag("http", "Host the directory via http in addition to gosync.").Default("true").Bool()
	hostHttpAddr = hostCommand.Flag("http-addr", "The address to start the http server on").Default(":32080").String()
	recvCommand  = app.Command("recv", "Sync a directory from the network, locally.")
	recvPath     = recvCommand.Arg("path", "The directory to sync into.").Default(".").File()
	recvHost     = recvCommand.Flag("host", "Specify a direct address to connect to.").String()
)

func StartRecv() {
	var wg sync.WaitGroup
	if *hostHttp {
		wg.Add(1)
		go func() {
			HostHTTP(*hostPath, *hostHttpAddr)
			wg.Done()
		}()
	}

	fmt.Printf("Scanning...")
	m := GetManifest(*hostPath)
	fmt.Printf("\r%-20s\n", "Found:")
	m.Print()

	wg.Add(1)
	go func() {
		HostSync(m, uint16(*mainPort))
		wg.Done()
	}()

	wg.Wait()
}

func StartRecv() {
	var host string = *recvHost
	if host == "" {
		host = GetHost()
	}

	Recv(recvPath.Name(), host)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case "host":
		StartHost()
	case "recv":
		StartRecv()
	default:
		kingpin.Usage()
	}

}

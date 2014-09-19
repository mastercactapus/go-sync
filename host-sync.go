package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"github.com/armon/mdns"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
)

func SendSync(m *Manifest, c net.Conn, mEncoded []byte) {
	defer c.Close()

	buf := make([]byte, 8)

	_, err := io.ReadFull(c, buf[:3])
	if err != nil {
		return
	}
	if string(buf[:3]) != "hi\n" {
		return
	}

	var mLen int64 = int64(len(mEncoded))

	err = binary.Write(c, binary.BigEndian, &mLen)
	if err != nil {
		return
	}

	_, err = c.Write(mEncoded)
	if err != nil {
		return
	}

	_, err = io.ReadFull(c, buf[:5])
	if err != nil {
		return
	}

	if string(buf[:5]) != "woot\n" {
		return
	}

	//we want a big buffer, so we can shove small files together
	//while the network is working
	writer := bufio.NewWriterSize(c, 1024*1024*32)

	for _, v := range m.Nodes {
		if v.IsDir {
			continue
		}
		file, err := os.Open(filepath.Join(m.Root, v.RelativePath))
		if err != nil {
			continue
		}
		_, err = io.Copy(writer, file)
		file.Close()
		if err != nil {
			//cancel process
			return
		}
	}

	writer.Flush()

}

func HostSync(m *Manifest, port uint16) {
	var service *mdns.MDNSService
	var server *mdns.Server

	host, err := os.Hostname()
	if err != nil {
		log.Println("Could not get hostname: ", err)
		goto listen
	}
	service = &mdns.MDNSService{
		Instance: host,
		Service:  "_gosync._tcp",
		Port:     int(port),
		Info:     "gosync hosted directory",
	}
	err = service.Init()
	if err != nil {
		log.Println("Could not init mdns service: ", err)
		goto listen
	}

	server, err = mdns.NewServer(&mdns.Config{Zone: service})
	if err != nil {
		log.Println("Could not create mdns service: ", err)
	}
	defer server.Shutdown()

listen:

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(int(port)))
	defer listener.Close()
	if err != nil {
		log.Fatalln("Could not start server: ", err)
	}

	buf := new(bytes.Buffer)
	w, err := gzip.NewWriterLevel(buf, 9)
	if err != nil {
		log.Fatalln("Could not create compression stream: ", err)
	}

	fmt.Println("Compressing manifest...")
	enc := gob.NewEncoder(w)
	enc.Encode(m)
	w.Flush()
	w.Close()

	manifestEncoded := buf.Bytes()
	buf.Reset()

	fmt.Printf("\ngosync online and listening at :%d\n", port)

	for {
		c, err := listener.Accept()
		if err != nil {
			log.Fatalln("Failed to accept connection: ", err)
		}

		go SendSync(m, c, manifestEncoded)
	}
}

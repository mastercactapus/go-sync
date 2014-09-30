package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func RecvManifest(c io.ReadWriter) *SyncManifest {

	_, err = io.WriteString(c, "hi\n")
	if err != nil {
		log.Fatalln("Failed to communicate with host: ", err)
	}

	var manifestSize int64
	err = binary.Read(c, binary.BigEndian, &manifestSize)
	if err != nil {
		log.Fatalln("Transmission error: ", err)
	}

	buf := make([]byte, manifestSize)
	_, err = io.ReadFull(c, buf)
	if err != nil {
		log.Fatalln("Error downloading manifest: ", err)
	}

	bread := bytes.NewBuffer(buf)
	gread, err := gzip.NewReader(bread)
	if err != nil {
		log.Fatalln("Failed to decompress manifest: ", err)
	}
	dec := gob.NewDecoder(gread)

	s := new(SyncManifest)

	err = dec.Decode(&s)
	if err != nil {
		log.Fatalln("Failed to decode manifest: ", err)
	}
	return s
}

func Get(root string, host string) {
	c, err := net.Dial("tcp", host)
	if err != nil {
		log.Fatalln("Could not connect to host: ", err)
	}
	defer c.Close()

	m := RecvManifest(c)

	m.Print()
	fmt.Println("\nPress ENTER to proceed.")
	fmt.Scanln()

	io.WriteString(c, "woot\n")

	w := new(Watcher)
	w.Conn = c
	w.Printing = true
	w.Total = m.Size
	w.Transfered = 0

	//big buffer to cover for small files
	reader := bufio.NewReaderSize(w, *bufferSize)

	for _, n := range m.Nodes {
		npath := filepath.Join(root, strings.Replace(n.RelativePath, "\\", "/", -1))
		fmt.Println("\rReceived:", npath, PrettySize(n.Size))
		w.Print()
		if n.IsDir {
			os.MkdirAll(npath, 0777)
			continue
		} else if n.IsLink {
			os.Link(n.LinkPath, npath)
			continue
		}
		os.MkdirAll(filepath.Dir(npath), 0777)
		file, err := os.Create(npath)
		if err != nil {
			log.Fatalln("\rCould not write file: ", err)
		}

		if n.Size == 0 {
			file.Close()
			continue
		}
		_, err = io.CopyN(file, reader, n.Size)
		file.Close()
		if err != nil {
			log.Fatalln("\rTransfer failed: ", err)
		}
	}
	fmt.Println("\r                                   ")

}

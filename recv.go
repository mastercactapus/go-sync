package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func RecvManifest(c io.ReadWriter) *SyncManifest {
	var err error
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

func WriteFile(path string, data []byte) {
	ioutil.WriteFile(path, data, 0777)
}

func Recv(root string, host string) {
	c, err := net.Dial("tcp", host)
	if err != nil {
		log.Fatalln("Could not connect to host: ", err)
	}
	defer c.Close()

	m := RecvManifest(c)

	m.Print()
	fmt.Println("\nPress ENTER to proceed.")
	fmt.Scanln()

	for _, v := range m.Directories {
		npath := filepath.Join(root, strings.Replace(v, "\\", "/", -1))
		os.MkdirAll(npath, 0777)
	}
	for _, v := range m.Links {
		npath := filepath.Join(root, strings.Replace(v.NewName, "\\", "/", -1))
		os.Link(v.OldName, npath)
	}

	io.WriteString(c, "woot\n")

	w := new(TransferWatcher)
	w.Conn = c
	w.Printing = true
	w.Total = m.Size
	w.Transfered = 0

	//big buffer to cover for small files
	reader := bufio.NewReaderSize(w, *bufferSize)

	wg := new(sync.WaitGroup)

	maxSize := int64(*bufferSize / 4)

	for _, n := range m.Files {
		npath := filepath.Join(root, strings.Replace(n.Name, "\\", "/", -1))
		fmt.Println("\rReceived:", npath, PrettySize(n.Size))
		w.Print()

		if n.Size < maxSize {
			buff := make([]byte, n.Size)
			_, err = io.ReadFull(reader, buff)
			if err != nil {
				log.Fatalln("\rTransfer failed: ", err)
			}
			wg.Add(1)
			go func() {
				err := ioutil.WriteFile(npath, buff, 0666)
				if err != nil {
					log.Fatalln("\rWrite failed: ", err)
				}
				wg.Done()
			}()
			continue
		}

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
	wg.Wait()
	fmt.Println("\r                                   ")
}

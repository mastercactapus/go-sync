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
)

func PrintProgress(done int64, total int64) {
	fmt.Printf("\r%s of %s                           ", PrettySize(done), PrettySize(total))
}

type Watcher struct {
	Conn       net.Conn
	Transfered int64
	Total      int64
	Printing   bool
}

func (w *Watcher) Read(b []byte) (int, error) {
	n, err := w.Conn.Read(b)
	w.Transfered += int64(n)
	if w.Printing {
		PrintProgress(w.Transfered, w.Total)
	}
	return n, err
}
func (w *Watcher) Close() error {
	w.Printing = false
	return w.Conn.Close()
}

func GetAddr(entry *mdns.ServiceEntry) net.IP {
	var addr net.IP
	if entry.Addr != nil {
		addr = entry.Addr
	} else if entry.AddrV4 != nil {
		addr = entry.AddrV4
	} else if entry.AddrV6 != nil {
		addr = entry.AddrV6
	}
	return addr
}

func GetHost() string {
	fmt.Println("Searching for hosts...")
	ch := make(chan *mdns.ServiceEntry, 8)
	mdns.Lookup("_gosync._tcp", ch)
	close(ch)
	if len(ch) == 0 {
		log.Fatalln("No hosts available, specify manually or try again")
	}
	entries := make([]mdns.ServiceEntry, len(ch))
	i := 1
	for entry := range ch {
		addr := GetAddr(entry)
		fmt.Printf("[%d] << %s >> %s:%d\n", i, entry.Host, addr, entry.Port)
		entries[i-1] = *entry
		i++
	}
	i = 0
	first := true
	for i < 1 || i > len(entries) {
		if !first {
			fmt.Println("Invalid, try again")
		}
		first = false
		fmt.Printf("Selection: ")
		fmt.Scanf("%d", &i)
	}

	fmt.Println()

	return fmt.Sprintf("%s:%d", GetAddr(&entries[i-1]), entries[i-1].Port)
}

func Get(root string, host string) {
	c, err := net.Dial("tcp", host)
	defer c.Close()
	if err != nil {
		log.Fatalln("Could not connect to host: ", err)
	}

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

	m := new(Manifest)

	err = dec.Decode(&m)
	if err != nil {
		log.Fatalln("Failed to decode manifest: ", err)
	}

	PrintManifest(m)
	fmt.Println("\nPress ENTER to proceed.")
	fmt.Scanln()

	io.WriteString(c, "woot\n")

	w := new(Watcher)
	w.Conn = c
	w.Printing = true
	w.Total = m.Size
	w.Transfered = 0

	//big buffer to cover for small files
	reader := bufio.NewReaderSize(w, 1024*1024*32)

	for _, n := range m.Nodes {
		npath := filepath.Join(root, n.RelativePath)
		fmt.Println("\rReceived:", npath, PrettySize(n.Size))
		if n.IsDir {
			os.MkdirAll(npath, 0777)
			continue
		}
		os.MkdirAll(filepath.Dir(npath), 0777)
		file, err := os.Create(npath)
		if err != nil {
			log.Fatalln("Could not write file: ", err)
		}

		if n.Size == 0 {
			file.Close()
			continue
		}
		_, err = io.CopyN(file, reader, n.Size)
		file.Close()
		if err != nil {
			log.Fatalln("Transfer failed: ", err)
		}
	}
	fmt.Println("\r                                   ")

}

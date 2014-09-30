package main

import (
	"fmt"
	"net"
	"time"
)

func PrintProgress(done int64, total int64) {
	fmt.Printf("\r%s of %s                           ", PrettySize(done), PrettySize(total))
}

type Watcher struct {
	Conn       net.Conn
	Transfered int64
	Total      int64
	Printing   bool
	LastPrint  time.Time
}

func (w *Watcher) Print() {
	w.LastPrint = time.Now()
	PrintProgress(w.Transfered, w.Total)
}

func (w *Watcher) Read(b []byte) (int, error) {
	n, err := w.Conn.Read(b)
	w.Transfered += int64(n)
	if w.Printing && time.Since(w.LastPrint) > 100*time.Millisecond {
		w.Print()
	}
	return n, err
}
func (w *Watcher) Close() error {
	w.Printing = false
	return w.Conn.Close()
}

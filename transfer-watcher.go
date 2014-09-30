package main

import (
	"fmt"
	"net"
	"time"
)

func PrintProgress(done int64, total int64) {
	fmt.Printf("\r%s of %s                           ", PrettySize(done), PrettySize(total))
}

type TransferWatcher struct {
	Conn       net.Conn
	Transfered int64
	Total      int64
	Printing   bool
	LastPrint  time.Time
}

func (w *TransferWatcher) Print() {
	w.LastPrint = time.Now()
	PrintProgress(w.Transfered, w.Total)
}

func (w *TransferWatcher) Read(b []byte) (int, error) {
	n, err := w.Conn.Read(b)
	w.Transfered += int64(n)
	if w.Printing && time.Since(w.LastPrint) > 100*time.Millisecond {
		w.Print()
	}
	return n, err
}
func (w *TransferWatcher) Close() error {
	w.Printing = false
	return w.Conn.Close()
}

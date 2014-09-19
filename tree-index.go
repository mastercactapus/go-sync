package main

import (
	"io/ioutil"
	"log"
	"math"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
)

var Labels = [...]string{"B", "KiB", "MiB", "GiB", "TiB"}

type ManifestNode struct {
	RelativePath string
	IsDir        bool
	Size         int64
}

type Manifest struct {
	Nodes          []ManifestNode
	Root           string
	FileCount      int
	DirectoryCount int
	Size           int64
}

type ByPath []ManifestNode

func (n ByPath) Len() int {
	return len(n)
}
func (n ByPath) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}
func (n ByPath) Less(i, j int) bool {
	return n[i].RelativePath < n[j].RelativePath
}

func PrettySize(bytes int64) string {
	fbyte := float64(bytes)
	for i := len(Labels) - 1; i >= 0; i-- {
		if fbyte >= math.Pow(1000.0, float64(i))-1 {
			return strconv.FormatFloat(fbyte/math.Pow(1024.0, float64(i)), 'f', 2, 64) + Labels[i]
		}
	}

	panic("byte count out of range")
}

func GetManifest(path string) *Manifest {
	var wg sync.WaitGroup
	pipe := make(chan ManifestNode, 1000)
	var total uint64

	wg.Add(1)
	go GetNodes(path, pipe, &wg, &total)
	wg.Wait()

	m := new(Manifest)
	m.Root = path
	m.Nodes = make([]ManifestNode, total)

	for i := range m.Nodes {
		node := <-pipe
		m.Nodes[i] = node
		if node.IsDir {
			m.DirectoryCount++
		} else {
			m.FileCount++
			m.Size += node.Size
		}
	}
	sort.Sort(ByPath(m.Nodes))

	return m
}

func GetNodes(path string, pipe chan ManifestNode, wg *sync.WaitGroup, total *uint64) {
	contents, err := ioutil.ReadDir(path)
	if err != nil {
		wg.Done()
		log.Print("Could not read dir: ", err, "\n")
		return
	}
	atomic.AddUint64(total, uint64(len(contents)))
	nodes := make([]ManifestNode, len(contents))
	for k, v := range contents {
		node := ManifestNode{
			IsDir:        v.IsDir(),
			Size:         v.Size(),
			RelativePath: path + "/" + v.Name(),
		}

		if node.IsDir {
			wg.Add(1)
			go GetNodes(node.RelativePath, pipe, wg, total)
		}

		nodes[k] = node
	}

	wg.Done()

	for _, v := range nodes {
		pipe <- v
	}
}

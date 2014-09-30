package main

import (
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
)

var Labels = [...]string{"B", "KiB", "MiB", "GiB", "TiB"}

func PrettySize(bytes int64) string {
	fbyte := float64(bytes)
	for i := len(Labels) - 1; i >= 0; i-- {
		if fbyte >= math.Pow(1000.0, float64(i))-1 {
			return strconv.FormatFloat(fbyte/math.Pow(1024.0, float64(i)), 'f', 2, 64) + Labels[i]
		}
	}

	panic("byte count out of range")
}

func GetManifest(root string) *Manifest {
	var wg sync.WaitGroup
	var err error
	pipe := make(chan ManifestNode, 1000)
	var total uint64

	wg.Add(1)
	go GetNodes(root, pipe, &wg, &total)
	wg.Wait()

	m := new(Manifest)
	m.Root = root
	m.Nodes = make([]ManifestNode, total)

	for i := range m.Nodes {
		node := <-pipe
		node.RelativePath, err = filepath.Rel(root, node.RelativePath)
		if err != nil {
			log.Fatalln("Failed to process path: ", err)
		}
		m.Nodes[i] = node
		if node.IsDir {
			m.DirectoryCount++
		} else if node.IsLink {
			m.LinkCount++
		} else {
			m.FileCount++
			m.Size += node.Size
		}
	}
	sort.Sort(ByPath(m.Nodes))

	return m
}

func GetNodes(root string, pipe chan ManifestNode, wg *sync.WaitGroup, total *uint64) {
	contents, err := ioutil.ReadDir(root)
	if err != nil {
		wg.Done()
		log.Print("Could not read dir: ", err, "\n")
		return
	}
	atomic.AddUint64(total, uint64(len(contents)))
	nodes := make([]ManifestNode, len(contents))
	for k, v := range contents {
		mode := v.Mode()
		var isLink bool
		var linkPath string
		var path string = filepath.Join(root, v.Name())
		if !mode.IsRegular() && !mode.IsDir() {
			isLink = true
			linkPath, err = os.Readlink(path)
			if err != nil {
				log.Println("Failed to read: ", err)
				continue
			}
		}

		node := ManifestNode{
			IsDir:        v.IsDir(),
			IsLink:       isLink,
			LinkPath:     linkPath,
			Size:         v.Size(),
			RelativePath: path,
		}

		if node.IsDir {
			wg.Add(1)
			go GetNodes(path, pipe, wg, total)
		}

		nodes[k] = node
	}

	wg.Done()

	for _, v := range nodes {
		pipe <- v
	}
}

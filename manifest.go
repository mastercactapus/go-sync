package main

import (
	"fmt"
)

type Manifest struct {
	Nodes          []ManifestNode
	Root           string
	FileCount      int
	DirectoryCount int
	LinkCount      int
	Size           int64
}
type ManifestNode struct {
	RelativePath string
	IsDir        bool
	IsLink       bool
	LinkPath     string
	Size         int64
}

type SyncManifest struct {
	Root        string
	Directories []string
	Links       []Link
	Files       []File
	Size        int64
}
type Link struct {
	OldName string
	NewName string
}

type File struct {
	Name string
	Size int64
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

func (m *Manifest) Print() {
	fmt.Printf(" %d Files\n", m.FileCount)
	fmt.Printf(" %d Links\n", m.LinkCount)
	fmt.Printf(" %d Directories\n", m.DirectoryCount)
	fmt.Printf(" Total Size: %s\n", PrettySize(m.Size))
}
func (s *SyncManifest) Print() {
	fmt.Printf(" %d Files\n", len(s.Files))
	fmt.Printf(" %d Links\n", len(s.Links))
	fmt.Printf(" %d Directories\n", len(s.Directories))
	fmt.Printf(" Total Size: %s\n", PrettySize(s.Size))
}

func (m *Manifest) MakeSync() *SyncManifest {
	s := new(SyncManifest)
	s.Files = make([]File, m.FileCount)
	s.Directories = make([]string, m.DirectoryCount)
	s.Links = make([]Link, m.LinkCount)
	f, d, l := 0, 0, 0
	for _, v := range m.Nodes {
		if v.IsDir {
			s.Directories[d] = v.RelativePath
			d++
		} else if v.IsLink {
			s.Links[l] = Link{
				OldName: v.RelativePath,
				NewName: v.LinkPath,
			}
			l++
		} else {
			s.Files[f] = File{
				Name: v.RelativePath,
				Size: v.Size,
			}
			f++
		}
	}
	s.Size = m.Size
	s.Root = m.Root
	return s
}

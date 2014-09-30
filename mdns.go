package main

import (
	"fmt"
	"github.com/armon/mdns"
	"log"
)

func GetAddr(entry *mdns.ServiceEntry) string {
	//if entry.AddrV6 != nil {
	//	return "[" + entry.AddrV6.String() + "]"
	//}

	if entry.AddrV4 != nil {
		return entry.AddrV4.String()
	}
	panic("Unknown address")
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

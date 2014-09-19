package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"sync"
)

func getPath(args []string) string {
	if len(args) > 0 {
		return args[0]
	} else {
		return "."
	}
}

func PrintManifest(m *Manifest) {
	fmt.Printf("File Count     : %d\n", m.FileCount)
	fmt.Printf("Directory Count: %d\n", m.DirectoryCount)
	fmt.Printf("Total Size     : %s\n", PrettySize(m.Size))
}

func main() {
	var mainPort uint16
	gosyncCommand := &cobra.Command{
		Use:   "gosync",
		Short: "gosync is a tool to sync a directory over a high-speed network",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}
	gosyncCommand.PersistentFlags().Uint16VarP(&mainPort, "port", "p", 32011, "port number to listen or connect to when syncing")

	var hostHttp bool
	var httpPort uint16
	hostCommand := &cobra.Command{
		Use:   "host [path]",
		Short: "host a directory on the network",
		Long:  "hosts a directory on the network, broadcasting it's existence.",
		Run: func(cmd *cobra.Command, args []string) {
			var wg sync.WaitGroup
			if hostHttp {
				wg.Add(1)
				go func() {
					HostHTTP(getPath(args), httpPort)
					wg.Done()
				}()
			}

			fmt.Println("Scanning...")
			m := GetManifest(getPath(args))
			fmt.Println("\nHosted Contents:")
			PrintManifest(m)

			wg.Add(1)
			go func() {
				HostSync(m, mainPort)
				wg.Done()
			}()

			wg.Wait()
		},
	}
	hostCommand.Flags().BoolVar(&hostHttp, "http", true, "enable/disable built-in http server while hosting")
	hostCommand.Flags().Uint16Var(&httpPort, "http-port", 32080, "port number to listen for built-in http server")

	var recvHost string
	getCommand := &cobra.Command{
		Use:   "get [path]",
		Short: "receives a hosted directory locally",
		Long:  "receives a hosted directory to a local path",
		Run: func(cmd *cobra.Command, args []string) {
			path := getPath(args)
			if recvHost == "" {
				recvHost = GetHost()
			}

			Get(path, recvHost)
		},
	}
	getCommand.Flags().StringVarP(&recvHost, "host", "h", "", "A specific address to receive from")

	gosyncCommand.AddCommand(hostCommand, getCommand)
	gosyncCommand.Execute()
}

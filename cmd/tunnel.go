package cmd

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"

	"github.com/spf13/cobra"
)

var serverAddr string

var tunnelCmd = &cobra.Command{
	Use:   "tunnel [local_port]",
	Short: "Start a tunnel to a local port",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		localPortStr := args[0]
		localPort, err := strconv.Atoi(localPortStr)
		if err != nil {
			log.Fatalf("Invalid local port: %s", localPortStr)
		}

		log.Printf("Starting tunnel to local port %d via server %s...\n", localPort, serverAddr)

		// Connect to the VM's control port
		serverConn, err := net.Dial("tcp", serverAddr)
		if err != nil {
			log.Fatalf("Error connecting to server %s: %v", serverAddr, err)
		}
		defer serverConn.Close()
		log.Printf("Connected to server at %s. Waiting for incoming traffic...\n", serverAddr)

		// Connect to the local application
		localAppAddr := fmt.Sprintf("localhost:%d", localPort)
		localConn, err := net.Dial("tcp", localAppAddr)
		if err != nil {
			log.Fatalf("Error connecting to local application %s: %v", localAppAddr, err)
		}
		defer localConn.Close()
		log.Printf("Connected to local application at %s. Piping traffic...\n", localAppAddr)

		go func() {
			_, err := io.Copy(localConn, serverConn)
			if err != nil {
				log.Printf("Error piping server -> local: %v\n", err)
			}
			log.Println("Server -> Local stream closed.")
		}()

		_, err = io.Copy(serverConn, localConn)
		if err != nil {
			log.Printf("Error piping local -> server: %v\n", err)
		}
		log.Println("Local -> Server stream closed. Tunnel stopping.")
	},
}

func init() {
	rootCmd.AddCommand(tunnelCmd)

	tunnelCmd.Flags().StringVarP(&serverAddr, "server", "s", "localhost:9999", "Address of the Tempok server's control port")
}

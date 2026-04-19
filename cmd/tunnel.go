package cmd

import (
	"fmt"
	"io"
	"log"
	"net"
	"strconv"

	"github.com/hashicorp/yamux"
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

		// Setup Yamux Client
		session, err := yamux.Client(serverConn, nil)
		if err != nil {
			log.Fatalf("Error creating yamux client session: %v", err)
		}

		// Accept streams continuously
		for {
			stream, err := session.Accept()
			if err != nil {
				log.Fatalf("Yamux session closed: %v", err)
			}
			log.Println("New stream accepted from server.")

			go func(stream net.Conn) {
				defer stream.Close()

				// Connect to the local application
				localAppAddr := fmt.Sprintf("localhost:%d", localPort)
				localConn, err := net.Dial("tcp", localAppAddr)
				if err != nil {
					log.Printf("Error connecting to local application %s: %v", localAppAddr, err)
					return
				}
				defer localConn.Close()
				
				go func() {
					_, err := io.Copy(localConn, stream)
					if err != nil && err != io.EOF {
						// log.Printf("Error piping stream -> local: %v\n", err)
					}
				}()

				_, err = io.Copy(stream, localConn)
				if err != nil && err != io.EOF {
					// log.Printf("Error piping local -> stream: %v\n", err)
				}
				log.Println("Local stream closed.")
			}(stream)
		}
	},
}

func init() {
	rootCmd.AddCommand(tunnelCmd)

	tunnelCmd.Flags().StringVarP(&serverAddr, "server", "s", "localhost:9999", "Address of the Tempok server's control port")
}

package cmd

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/spf13/cobra"
)

var controlPort int
var publicPort int

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the Tempok server",
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("Starting Tempok server on control port %d and public port %d...\n", controlPort, publicPort)

		// Listen for the client (tunnel) connection on controlPort
		controlAddr := fmt.Sprintf(":%d", controlPort)
		controlListener, err := net.Listen("tcp", controlAddr)
		if err != nil {
			log.Fatalf("Error listening on control port %d: %v", controlPort, err)
		}
		defer controlListener.Close()

		log.Printf("Waiting for Tempok client to connect on %s...\n", controlAddr)
		clientConn, err := controlListener.Accept()
		if err != nil {
			log.Fatalf("Error accepting client connection: %v", err)
		}
		log.Printf("Client connected from %s\n", clientConn.RemoteAddr())

		// Listen for public internet traffic on publicPort
		publicAddr := fmt.Sprintf(":%d", publicPort)
		publicListener, err := net.Listen("tcp", publicAddr)
		if err != nil {
			log.Fatalf("Error listening on public port %d: %v", publicPort, err)
		}
		defer publicListener.Close()

		log.Printf("Waiting for public connections on %s...\n", publicAddr)
		publicConn, err := publicListener.Accept()
		if err != nil {
			log.Fatalf("Error accepting public connection: %v", err)
		}
		log.Printf("Public connection received from %s\n", publicConn.RemoteAddr())

		// Pipe traffic back and forth (Phase 1: Proof of concept, single request)
		log.Println("Piping traffic between public connection and client connection...")
		
		go func() {
			_, err := io.Copy(clientConn, publicConn)
			if err != nil {
				log.Printf("Error piping public -> client: %v\n", err)
			}
			log.Println("Public -> Client stream closed.")
		}()

		_, err = io.Copy(publicConn, clientConn)
		if err != nil {
			log.Printf("Error piping client -> public: %v\n", err)
		}
		log.Println("Client -> Public stream closed. Server stopping.")
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().IntVarP(&controlPort, "control-port", "c", 9999, "Port for the local tempok client to connect to")
	serverCmd.Flags().IntVarP(&publicPort, "public-port", "p", 8000, "Port for public internet traffic")
}

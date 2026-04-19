package cmd

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/hashicorp/yamux"
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

		// Wrap the control connection in a Yamux server session
		session, err := yamux.Server(clientConn, nil)
		if err != nil {
			log.Fatalf("Error creating yamux session: %v", err)
		}

		// Listen for public internet traffic on publicPort
		publicAddr := fmt.Sprintf(":%d", publicPort)
		publicListener, err := net.Listen("tcp", publicAddr)
		if err != nil {
			log.Fatalf("Error listening on public port %d: %v", publicPort, err)
		}
		defer publicListener.Close()

		actualPort := publicListener.Addr().(*net.TCPAddr).Port
		log.Printf("Waiting for public connections on port %d...\n", actualPort)

		for {
			publicConn, err := publicListener.Accept()
			if err != nil {
				log.Printf("Error accepting public connection: %v\n", err)
				continue
			}
			log.Printf("Public connection received from %s\n", publicConn.RemoteAddr())

			go func(pConn net.Conn) {
				defer pConn.Close()

				// Open a new logical stream via Yamux
				stream, err := session.Open()
				if err != nil {
					log.Printf("Error opening yamux stream: %v\n", err)
					return
				}
				defer stream.Close()

				go func() {
					_, err := io.Copy(stream, pConn)
					if err != nil && err != io.EOF {
						// log.Printf("Error piping public -> stream: %v\n", err)
					}
				}()

				_, err = io.Copy(pConn, stream)
				if err != nil && err != io.EOF {
					// log.Printf("Error piping stream -> public: %v\n", err)
				}
				log.Println("Stream closed.")
			}(publicConn)
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().IntVarP(&controlPort, "control-port", "c", 9999, "Port for the local tempok client to connect to")
	serverCmd.Flags().IntVarP(&publicPort, "public-port", "p", 0, "Port for public internet traffic (default 0 for random)")
}

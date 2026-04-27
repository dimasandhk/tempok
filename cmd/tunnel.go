package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/hashicorp/yamux"
	"github.com/spf13/cobra"
	"github.com/dimasandhk/tempok/internal/api"
)

var serverAddr string
var tunnelSecret string
var expires string

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

		if tunnelSecret == "" {
			log.Fatal("Tunnel requires a --secret flag to connect")
		}

		log.Printf("Starting tunnel to local port %d via server %s...\n", localPort, serverAddr)

		serverConn, err := net.Dial("tcp", serverAddr)
		if err != nil {
			log.Fatalf("Error connecting to server %s: %v", serverAddr, err)
		}
		defer serverConn.Close()

		// Perform handshake
		encoder := json.NewEncoder(serverConn)
		req := api.HandshakeRequest{
			Secret: tunnelSecret,
			Type:   "tunnel",
		}
		if err := encoder.Encode(req); err != nil {
			log.Fatalf("Error sending handshake: %v", err)
		}

		decoder := json.NewDecoder(serverConn)
		var resp api.HandshakeResponse
		if err := decoder.Decode(&resp); err != nil {
			log.Fatalf("Error reading handshake response: %v", err)
		}

		if resp.Error != "" {
			log.Fatalf("Server rejected connection: %s", resp.Error)
		}

		log.Printf("Connected to server at %s. Waiting for incoming traffic...\n", serverAddr)

		session, err := yamux.Client(serverConn, nil)
		if err != nil {
			log.Fatalf("Error creating yamux client session: %v", err)
		}

		// Setup graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			<-sigChan
			log.Println("Received termination signal. Closing tunnel gracefully...")
			session.Close()
			serverConn.Close()
			os.Exit(0)
		}()

		for {
			stream, err := session.Accept()
			if err != nil {
				log.Fatalf("Yamux session closed: %v", err)
			}
			log.Println("New stream accepted from server.")

			go func(stream net.Conn) {
				defer stream.Close()

				localAppAddr := fmt.Sprintf("localhost:%d", localPort)
				localConn, err := net.Dial("tcp", localAppAddr)
				if err != nil {
					log.Printf("Error connecting to local application %s: %v", localAppAddr, err)
					return
				}
				defer localConn.Close()

				go func() {
					io.Copy(localConn, stream)
				}()

				io.Copy(stream, localConn)
				log.Println("Local stream closed.")
			}(stream)
		}
	},
}

func init() {
	rootCmd.AddCommand(tunnelCmd)

	tunnelCmd.Flags().StringVarP(&serverAddr, "server", "s", "localhost:9999", "Address of the Tempok server's control port")
	tunnelCmd.Flags().StringVar(&tunnelSecret, "secret", "", "Secret required to authenticate to server")
	tunnelCmd.Flags().StringVar(&expires, "expires", "1h", "Time to live for the tunnel (e.g. 30m, 1h)")
}

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/rpc"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/yamux"
	"github.com/spf13/cobra"
	"github.com/dimasandhk/tempok/internal/api"
	"github.com/dimasandhk/tempok/internal/state"
)

var controlPort int
var publicPort int
var secret string

var manager = state.NewManager()

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the Tempok server",
	Run: func(cmd *cobra.Command, args []string) {
		if secret == "" {
			log.Fatal("Server requires a --secret flag to start")
		}

		log.Printf("Starting Tempok server on control port %d...\n", controlPort)

		controlAddr := fmt.Sprintf(":%d", controlPort)
		controlListener, err := net.Listen("tcp", controlAddr)
		if err != nil {
			log.Fatalf("Error listening on control port %d: %v", controlPort, err)
		}
		defer controlListener.Close()

		for {
			clientConn, err := controlListener.Accept()
			if err != nil {
				log.Printf("Error accepting client connection: %v", err)
				continue
			}

			go handleClient(clientConn)
		}
	},
}

func handleClient(conn net.Conn) {
	// Handshake timeout
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	decoder := json.NewDecoder(conn)
	var req api.HandshakeRequest
	if err := decoder.Decode(&req); err != nil {
		log.Printf("Handshake failed: %v", err)
		conn.Close()
		return
	}
	conn.SetReadDeadline(time.Time{}) // Reset deadline

	encoder := json.NewEncoder(conn)
	if req.Secret != secret {
		encoder.Encode(api.HandshakeResponse{Error: "Invalid secret"})
		conn.Close()
		return
	}

	encoder.Encode(api.HandshakeResponse{Message: "Authenticated"})

	session, err := yamux.Server(conn, nil)
	if err != nil {
		log.Printf("Error creating yamux session: %v", err)
		conn.Close()
		return
	}

	if req.Type == "api" {
		handleAPIClient(session)
	} else if req.Type == "tunnel" {
		handleTunnelClient(session, conn)
	} else {
		log.Printf("Unknown client type: %s", req.Type)
		session.Close()
	}
}

func handleAPIClient(session *yamux.Session) {
	defer session.Close()
	stream, err := session.Accept()
	if err != nil {
		log.Printf("Error accepting API stream: %v", err)
		return
	}
	defer stream.Close()

	rpcServer := rpc.NewServer()
	rpcServer.Register(&API{manager: manager})
	rpcServer.ServeConn(stream)
}

func handleTunnelClient(session *yamux.Session, conn net.Conn) {
	// Start public listener for this tunnel
	publicAddr := fmt.Sprintf(":%d", publicPort)
	publicListener, err := net.Listen("tcp", publicAddr)
	if err != nil {
		log.Printf("Error listening on public port: %v", err)
		session.Close()
		return
	}

	actualPort := publicListener.Addr().(*net.TCPAddr).Port
	log.Printf("Tunnel created. Public port: %d\n", actualPort)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Default TTL is 1 hour if not specified, but let's assume we read it from handshake if we want.
	// For now, default 1 hour. We can update this later or have the client send it.
	tunnelID := uuid.New().String()
	manager.Add(tunnelID, 1*time.Hour, func() {
		publicListener.Close()
		session.Close()
		conn.Close()
		cancel()
	}, actualPort)

	go func() {
		<-ctx.Done()
		publicListener.Close()
		session.Close()
		conn.Close()
	}()

	for {
		publicConn, err := publicListener.Accept()
		if err != nil {
			if ctx.Err() == nil {
				log.Printf("Error accepting public connection on port %d: %v", actualPort, err)
			}
			break
		}

		go func(pConn net.Conn) {
			defer pConn.Close()

			stream, err := session.Open()
			if err != nil {
				log.Printf("Error opening yamux stream: %v\n", err)
				return
			}
			defer stream.Close()

			var wg sync.WaitGroup
			wg.Add(2)

			go func() {
				defer wg.Done()
				io.Copy(stream, pConn)
				stream.Close()
			}()

			go func() {
				defer wg.Done()
				io.Copy(pConn, stream)
			}()

			wg.Wait()
		}(publicConn)
	}
}

type API struct {
	manager *state.Manager
}

func (a *API) List(args *api.ListArgs, reply *api.ListReply) error {
	tunnels := a.manager.List()
	for _, t := range tunnels {
		reply.Tunnels = append(reply.Tunnels, api.TunnelInfo{
			ID:         t.ID,
			ExpiresAt:  t.ExpiresAt,
			IsForever:  t.IsForever,
			PublicPort: t.PublicPort,
		})
	}
	return nil
}

func (a *API) Extend(args *api.ExtendArgs, reply *api.ExtendReply) error {
	if err := a.manager.Extend(args.ID, args.Duration); err != nil {
		reply.Error = err.Error()
	}
	return nil
}

func (a *API) Persist(args *api.PersistArgs, reply *api.PersistReply) error {
	if err := a.manager.Persist(args.ID); err != nil {
		reply.Error = err.Error()
	}
	return nil
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().IntVarP(&controlPort, "control-port", "c", 9999, "Port for the local tempok client to connect to")
	serverCmd.Flags().IntVarP(&publicPort, "public-port", "p", 0, "Port for public internet traffic (default 0 for random)")
	serverCmd.Flags().StringVarP(&secret, "secret", "s", "", "Secret required to authenticate clients")
}

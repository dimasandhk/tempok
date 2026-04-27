package cmd

import (
	"encoding/json"
	"fmt"
	"net"
	"net/rpc"

	"github.com/hashicorp/yamux"
	"github.com/dimasandhk/tempok/internal/api"
)

func newRPCClient(serverAddr, secret string) (*rpc.Client, error) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return nil, fmt.Errorf("error connecting to server %s: %v", serverAddr, err)
	}

	encoder := json.NewEncoder(conn)
	req := api.HandshakeRequest{
		Secret: secret,
		Type:   "api",
	}
	if err := encoder.Encode(req); err != nil {
		conn.Close()
		return nil, fmt.Errorf("error sending handshake: %v", err)
	}

	decoder := json.NewDecoder(conn)
	var resp api.HandshakeResponse
	if err := decoder.Decode(&resp); err != nil {
		conn.Close()
		return nil, fmt.Errorf("error reading handshake response: %v", err)
	}

	if resp.Error != "" {
		conn.Close()
		return nil, fmt.Errorf("server rejected connection: %s", resp.Error)
	}

	session, err := yamux.Client(conn, nil)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("error creating yamux client session: %v", err)
	}

	stream, err := session.Open()
	if err != nil {
		session.Close()
		conn.Close()
		return nil, fmt.Errorf("error opening api stream: %v", err)
	}

	client := rpc.NewClient(stream)
	return client, nil
}

package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/dimasandhk/tempok/internal/api"
)

var extendCmd = &cobra.Command{
	Use:   "extend [tunnel_id] [duration]",
	Short: "Extend the TTL of a specific tunnel",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if tunnelSecret == "" {
			log.Fatal("extend requires a --secret flag to connect")
		}

		tunnelID := args[0]
		durationStr := args[1]
		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			log.Fatalf("Invalid duration: %s. Example: 1h, 30m", durationStr)
		}

		client, err := newRPCClient(serverAddr, tunnelSecret)
		if err != nil {
			log.Fatalf("Failed to connect to API: %v", err)
		}
		defer client.Close()

		req := api.ExtendArgs{
			ID:       tunnelID,
			Duration: duration,
		}
		var resp api.ExtendReply

		if err := client.Call("API.Extend", &req, &resp); err != nil {
			log.Fatalf("RPC error: %v", err)
		}

		if resp.Error != "" {
			log.Fatalf("Failed to extend tunnel: %s", resp.Error)
		}

		fmt.Printf("Successfully extended tunnel %s by %s\n", tunnelID, durationStr)
	},
}

func init() {
	rootCmd.AddCommand(extendCmd)
	extendCmd.Flags().StringVarP(&serverAddr, "server", "s", "localhost:9999", "Address of the Tempok server's control port")
	extendCmd.Flags().StringVar(&tunnelSecret, "secret", "", "Secret required to authenticate to server")
}

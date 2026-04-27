package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/dimasandhk/tempok/internal/api"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List active tunnels",
	Run: func(cmd *cobra.Command, args []string) {
		if tunnelSecret == "" {
			log.Fatal("list requires a --secret flag to connect")
		}

		client, err := newRPCClient(serverAddr, tunnelSecret)
		if err != nil {
			log.Fatalf("Failed to connect to API: %v", err)
		}
		defer client.Close()

		var req api.ListArgs
		var resp api.ListReply

		if err := client.Call("API.List", &req, &resp); err != nil {
			log.Fatalf("RPC error: %v", err)
		}

		fmt.Println("Active Tunnels:")
		fmt.Printf("%-36s %-10s %-15s %s\n", "ID", "PORT", "TTL", "STATUS")
		fmt.Println("--------------------------------------------------------------------------------")
		for _, t := range resp.Tunnels {
			status := "Active"
			ttl := "Forever"
			if !t.IsForever {
				timeRemaining := time.Until(t.ExpiresAt)
				if timeRemaining < 0 {
					status = "Expired"
					ttl = "0s"
				} else {
					ttl = timeRemaining.Round(time.Second).String()
				}
			}
			fmt.Printf("%-36s %-10d %-15s %s\n", t.ID, t.PublicPort, ttl, status)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&serverAddr, "server", "s", "localhost:9999", "Address of the Tempok server's control port")
	listCmd.Flags().StringVar(&tunnelSecret, "secret", "", "Secret required to authenticate to server")
}

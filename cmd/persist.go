package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/dimasandhk/tempok/internal/api"
)

var persistCmd = &cobra.Command{
	Use:   "persist [tunnel_id]",
	Short: "Make a specific tunnel persistent (no expiration)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if tunnelSecret == "" {
			log.Fatal("persist requires a --secret flag to connect")
		}

		tunnelID := args[0]

		client, err := newRPCClient(serverAddr, tunnelSecret)
		if err != nil {
			log.Fatalf("Failed to connect to API: %v", err)
		}
		defer client.Close()

		req := api.PersistArgs{
			ID: tunnelID,
		}
		var resp api.PersistReply

		if err := client.Call("API.Persist", &req, &resp); err != nil {
			log.Fatalf("RPC error: %v", err)
		}

		if resp.Error != "" {
			log.Fatalf("Failed to persist tunnel: %s", resp.Error)
		}

		fmt.Printf("Successfully set tunnel %s to persist forever\n", tunnelID)
	},
}

func init() {
	rootCmd.AddCommand(persistCmd)
	persistCmd.Flags().StringVarP(&serverAddr, "server", "s", "localhost:9999", "Address of the Tempok server's control port")
	persistCmd.Flags().StringVar(&tunnelSecret, "secret", "", "Secret required to authenticate to server")
}

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tempok",
	Short: "Tempok is a self hosted, lightweight reverse proxy and tunneling tool",
	Long:  `A fast and flexible reverse proxy tool with time-bound tunneling capabilities.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

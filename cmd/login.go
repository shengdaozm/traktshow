package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zm/traktshow/config"
	"github.com/zm/traktshow/trakt"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Trakt.tv",
	Long:  `Initiates the device authentication process with Trakt.tv to get an access token.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			// If config doesn't exist, create a new one
			cfg = &config.Config{}
		}

		client, err := trakt.NewClient(cfg)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		token, err := client.Authenticate()
		if err != nil {
			fmt.Println("Authentication failed:", err)
			return
		}

		cfg.Token = token
		if err := cfg.Save(); err != nil {
			fmt.Println("Failed to save config:", err)
			return
		}

		fmt.Println("Successfully authenticated and saved token!")
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

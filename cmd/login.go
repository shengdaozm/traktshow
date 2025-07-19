package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zm/traktshow/trakt"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Trakt.tv",
	Long:  `Initiates the device authentication process with Trakt.tv to get an access token.`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := trakt.NewClient()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if err := client.Authenticate(); err != nil {
			fmt.Println("Authentication failed:", err)
			return
		}

		fmt.Println("Successfully authenticated!")
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

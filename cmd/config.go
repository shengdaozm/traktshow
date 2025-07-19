package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	clientID     string
	clientSecret string
)

type Credentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure your Trakt.tv Client ID and Client Secret",
	Long:  `The config command allows you to set and store your Trakt.tv Client ID and Client Secret.`,
	Run: func(cmd *cobra.Command, args []string) {
		if clientID == "" || clientSecret == "" {
			fmt.Println("Both --client-id and --client-secret are required.")
			return
		}

		creds := Credentials{
			ClientID:     clientID,
			ClientSecret: clientSecret,
		}

		configPath := getConfigPath("")
		if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
			fmt.Printf("Error creating config directory: %s\n", err)
			return
		}

		data, err := json.MarshalIndent(creds, "", "  ")
		if err != nil {
			fmt.Printf("Error marshalling credentials to JSON: %s\n", err)
			return
		}

		if err := ioutil.WriteFile(configPath, data, 0600); err != nil {
			fmt.Printf("Error writing credentials to file: %s\n", err)
			return
		}

		fmt.Printf("Credentials saved to %s\nRun 'traktshow login' to authenticate.", configPath)
	},
}

func init() {
	configCmd.Flags().StringVar(&clientID, "client-id", "", "Your Trakt.tv Client ID")
	configCmd.Flags().StringVar(&clientSecret, "client-secret", "", "Your Trakt.tv Client Secret")
	rootCmd.AddCommand(configCmd)
}

func getConfigPath(filename string) string {
	if filename == "" {
		filename = "credentials.json"
	}
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting user home directory:", err)
		return ""
	}
	return filepath.Join(home, ".config", "trakt", filename)
}
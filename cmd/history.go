package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/zm/traktshow/config"
	"github.com/zm/traktshow/trakt"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Fetch your watch history",
	Long:  `Fetches and displays your watch history from Trakt.tv. Use --export to save all history to a file.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		client, err := trakt.NewClient(cfg)
		if err != nil {
			fmt.Println("Error creating Trakt client:", err)
			return
		}

		export, _ := cmd.Flags().GetBool("export")

		if export {
			history, err := client.GetAllHistory("me")
			if err != nil {
				fmt.Println("Error fetching all history:", err)
				return
			}

			if err := saveHistoryToFile(history); err != nil {
				fmt.Println("Error saving history to file:", err)
				return
			}

			fmt.Println("Successfully exported all history to ~/.cache/traktshow/history.json")
		} else {
			limit, _ := cmd.Flags().GetInt("limit")
			history, err := client.GetHistory("me", limit, 1)
			if err != nil {
				fmt.Println("Error fetching history:", err)
				return
			}

			fmt.Println("\n--- Your Recent Trakt.tv History ---")
			fmt.Println("-------------------------------------")
			for _, item := range history {
				formattedTime := item.WatchedAt.Format("2006-01-02 15:04:05")

				if item.Show.Title != "" && item.Episode.Title != "" {
					fmt.Printf("[%s] %-30s S%02dE%02d - %s\n", formattedTime, item.Show.Title, item.Episode.Season, item.Episode.Number, item.Episode.Title)
				} else if item.Show.Title != "" {
					fmt.Printf("[%s] %-30s (Movie)\n", formattedTime, item.Show.Title)
				} else {
					fmt.Printf("[%s] Unknown item\n", formattedTime)
				}
			}
			fmt.Println("-------------------------------------")
		}
	},
}

func saveHistoryToFile(history []trakt.HistoryItem) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	cacheDir := filepath.Join(home, ".cache", "traktshow")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}

	filePath := filepath.Join(cacheDir, "history.json")
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

func init() {
	historyCmd.Flags().Int("limit", 25, "Number of records to fetch")
	historyCmd.Flags().Bool("export", false, "Export all watch history to a JSON file")
	rootCmd.AddCommand(historyCmd)
}
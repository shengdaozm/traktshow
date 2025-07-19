package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zm/traktshow/trakt"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Fetch your watch history",
	Long:  `Fetches and displays your watch history from Trakt.tv.`,
	Run: func(cmd *cobra.Command, args []string) {
		limit, _ := cmd.Flags().GetInt("limit")

		client, err := trakt.NewClient()
		if err != nil {
			fmt.Println("Error creating Trakt client:", err)
			return
		}

		history, err := client.GetHistory("me", limit)
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
	},
}

func init() {
	historyCmd.Flags().Int("limit", 25, "Number of records to fetch")
	rootCmd.AddCommand(historyCmd)
}

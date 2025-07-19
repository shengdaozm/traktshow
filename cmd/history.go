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

		for _, item := range history {
			// Format the watched_at time
			formattedTime := item.WatchedAt.Format("2006-01-02 15:04:05")

			// Determine if it's a movie or an episode
			if item.Show.Title != "" && item.Episode.Title != "" {
				// It's an episode
				fmt.Printf("[%s] %-30s S%02dE%02d - %s\n", formattedTime, item.Show.Title, item.Episode.Season, item.Episode.Number, item.Episode.Title)
			} else if item.Show.Title != "" { // Assuming if Episode.Title is empty, it's a movie
				// It's a movie (Trakt API might return movies under 'show' field with no episode info)
				fmt.Printf("[%s] %-30s (Movie)\n", formattedTime, item.Show.Title)
			} else {
				// Fallback for unexpected data structure
				fmt.Printf("[%s] Unknown item\n", formattedTime)
			}
		}
	},
}

func init() {
	historyCmd.Flags().Int("limit", 25, "Number of records to fetch")
	rootCmd.AddCommand(historyCmd)
}

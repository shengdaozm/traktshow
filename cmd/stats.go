package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zm/traktshow/trakt"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Fetch your stats",
	Long:  `Fetches and displays your statistics from Trakt.tv.`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := trakt.NewClient()
		if err != nil {
			fmt.Println("Error creating Trakt client:", err)
			return
		}

		stats, err := client.GetStats("me")
		if err != nil {
			fmt.Println("Error fetching stats:", err)
			return
		}

		fmt.Printf("Movies watched: %d\n", stats.Movies.Watched)
		fmt.Printf("Shows watched: %d\n", stats.Shows.Watched)
		fmt.Printf("Episodes watched: %d\n", stats.Episodes.Watched)
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}


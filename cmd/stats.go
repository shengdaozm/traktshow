package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zm/traktshow/config"
	"github.com/zm/traktshow/trakt"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Fetch your stats",
	Long:  `Fetches and displays your statistics from Trakt.tv.`,
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

		stats, err := client.GetStats("me")
		if err != nil {
			fmt.Println("Error fetching stats:", err)
			return
		}

		fmt.Println("--- Your Trakt.tv Statistics ---")
		fmt.Println("-------------------------------------")
		fmt.Printf("Movies watched: %d\n", stats.Movies.Watched)
		fmt.Printf("Shows watched: %d\n", stats.Shows.Watched)
		fmt.Printf("Episodes watched: %d\n", stats.Episodes.Watched)
		fmt.Println("-------------------------------------")
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}


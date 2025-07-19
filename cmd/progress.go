package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zm/traktshow/trakt"
)

var progressCmd = &cobra.Command{
	Use:   "progress",
	Short: "Fetch your watch progress",
	Long:  `Fetches and displays your watch progress from Trakt.tv.`,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := trakt.NewClient()
		if err != nil {
			fmt.Println("Error creating Trakt client:", err)
			return
		}

		progress, err := client.GetWatched("me")
		if err != nil {
			fmt.Println("Error fetching progress:", err)
			return
		}

		fmt.Println("\n--- Your Trakt.tv Watch Progress ---")
		fmt.Println("-------------------------------------")
		for _, item := range progress {
			fmt.Printf("%-30s Watched: %d/%d episodes\n", item.Show.Title, item.Plays, item.Show.AiredEpisodes)
		}
		fmt.Println("-------------------------------------")
	},
}

func init() {
	rootCmd.AddCommand(progressCmd)
}


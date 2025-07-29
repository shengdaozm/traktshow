package cmd

import (
	"fmt"
	"os"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/zm/traktshow/config"
	"github.com/zm/traktshow/trakt"
)

var progressCmd = &cobra.Command{
	Use:   "progress",
	Short: "Fetch your watch progress",
	Long:  `Fetches and displays your watch progress for shows from Trakt.tv. Shows that have been fully watched are not displayed.`,
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

		progress, err := client.GetWatched("me")
		if err != nil {
			fmt.Println("Error fetching progress:", err)
			return
		}

		fmt.Println("\n--- Your Trakt.tv Watch Progress ---")
		for _, item := range progress {
			if item.Plays < item.Show.AiredEpisodes {
				description := fmt.Sprintf("%-40s", item.Show.Title)
				bar := progressbar.NewOptions(item.Show.AiredEpisodes,
					progressbar.OptionSetDescription(description),
					progressbar.OptionSetWriter(os.Stderr),
					progressbar.OptionShowCount(),
					progressbar.OptionSetWidth(30),
					progressbar.OptionOnCompletion(func() {
						fmt.Fprint(os.Stderr, "\n")
					}),
				)
				bar.Set(item.Plays)
				bar.Finish() // Ensure newline after each bar
			}
		}
		fmt.Println("-------------------------------------")
	},
}

func init() {
	rootCmd.AddCommand(progressCmd)
}


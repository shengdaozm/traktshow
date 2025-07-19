package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "traktshow",
	Short: "A CLI tool to interact with Trakt.tv",
	Long:  `traktshow is a command-line interface for Trakt.tv, allowing you to view your watch history, progress, and stats.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

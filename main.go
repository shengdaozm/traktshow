package main

import (
	"github.com/spf13/cobra"
	"github.com/zm/traktshow/cmd"
)

var rootCmd = &cobra.Command{
	Use:   "traktshow",
	Short: "A CLI tool to interact with Trakt.tv",
	Long:  `traktshow is a command-line interface for Trakt.tv, allowing you to view your watch history, progress, and stats.`,
}

func main() {
	cmd.Execute()
}

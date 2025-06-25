package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display version, commit, and build date information for my-day.`,
	Run: func(cmd *cobra.Command, args []string) {
		showVersion()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func showVersion() {
	color.Cyan("my-day - DevOps Daily Standup Report Generator")
	fmt.Println()
	color.White("Version: %s", version)
	color.White("Commit:  %s", commit)
	color.White("Date:    %s", date)
}
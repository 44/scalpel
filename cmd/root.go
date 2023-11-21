package cmd

import (
	"os"
	"github.com/spf13/cobra"
	"fmt"
)

var rootCmd = &cobra.Command{
	Use:   "scalpel",
	Short: "toolset for dealing with logs",
	PersistentPreRun: func(cmd *cobra.Command, arg []string) {
		verbose, _ := cmd.Flags().GetCount("verbose")
		fmt.Println("Configure verbosity here", verbose)
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().CountP("verbose", "v", "verbose output")
}

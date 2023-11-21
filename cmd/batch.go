package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/44/scalpel/internal/batch"
	log "github.com/sirupsen/logrus"
)

var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "Extract log files from batches",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("batch called")
		for _, file := range args {
			fmt.Println("Processing", file)
			dest, err := cmd.Flags().GetString("dest")
			if err != nil {
				log.Errorf("Error getting destination: %v", err)
				return
			}
			force, err := cmd.Flags().GetBool("force")
			if err != nil {
				log.Errorf("Error getting force: %v", err)
				return
			}
			batch.FindAndExtractBatches(args, dest, force)
		}
	},
}

func init() {
	rootCmd.AddCommand(batchCmd)
	batchCmd.Flags().StringP("dest", "d", ".", "directory to place extracted files to")
	batchCmd.Flags().StringSliceP("match", "m", []string{}, "regex to match files to extract")
	batchCmd.Flags().BoolP("force", "f", false, "Force overwrite of existing files")
	batchCmd.Flags().BoolP("unpack", "z", false, "Unpack gzipped logs")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// batchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// batchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

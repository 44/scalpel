package cmd

import (
	"github.com/spf13/cobra"
	"github.com/44/scalpel/internal/batch"
	log "github.com/sirupsen/logrus"
)

var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "Extract log files from batches",
	Run: func(cmd *cobra.Command, args []string) {
		execute(cmd, args)
	},
}

func execute(cmd *cobra.Command, args []string) {
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

func init() {
	rootCmd.AddCommand(batchCmd)
	batchCmd.Flags().StringP("dest", "d", ".", "directory to place extracted files to")
	batchCmd.Flags().StringSliceP("match", "m", []string{}, "regex to match files to extract")
	batchCmd.Flags().BoolP("force", "f", false, "Force overwrite of existing files")
	batchCmd.Flags().BoolP("unpack", "z", false, "Unpack gzipped logs")
}

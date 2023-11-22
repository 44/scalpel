package cmd

import (
	"github.com/spf13/cobra"
	"github.com/44/scalpel/internal/batch"
)

var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "Extract log files from batches",
	Run: func(cmd *cobra.Command, args []string) {
		execute(cmd, args)
	},
}

func execute(cmd *cobra.Command, args []string) {
	opts := batch.Options{}
	opts.Dest, _ = cmd.Flags().GetString("dest")
	opts.Match, _ = cmd.Flags().GetStringSlice("match")
	opts.Unpack, _ = cmd.Flags().GetBool("unpack")
	opts.Test, _ = cmd.Flags().GetBool("test")
	opts.Long, _ = cmd.Flags().GetBool("long")
	opts.Force, _ = cmd.Flags().GetBool("force")
	batch.FindAndExtractBatches(args, opts)
}

func init() {
	rootCmd.AddCommand(batchCmd)
	batchCmd.Flags().StringP("dest", "d", ".", "directory to place extracted files to")
	batchCmd.Flags().StringSliceP("match", "m", []string{}, "regex to match files to extract")
	batchCmd.Flags().BoolP("force", "f", false, "Force overwrite of existing files")
	batchCmd.Flags().BoolP("unpack", "z", false, "Unpack gzipped logs")
	batchCmd.Flags().BoolP("test", "t", false, "Do not extract logs from batches, but rather print log files included")
	batchCmd.Flags().BoolP("long", "l", false, "Show size of the log files")
}

package cmd

import (
	"os"
	"github.com/spf13/cobra"
	log "github.com/sirupsen/logrus"
	"fmt"
)

var rootCmd = &cobra.Command{
	Use:   "scalpel",
	Short: "toolset for dealing with logs",
	PersistentPreRun: func(cmd *cobra.Command, arg []string) {
		verbose, _ := cmd.Flags().GetCount("verbose")
		switch verbose {
		case 0:
			log.SetLevel(log.ErrorLevel)
		case 1:
			log.SetLevel(log.WarnLevel)
		case 2:
			log.SetLevel(log.InfoLevel)
		case 3:
			log.SetLevel(log.DebugLevel)
		case 4:
			log.SetLevel(log.TraceLevel)
		}
		log.SetFormatter(&logFormatter{ name: "scalpel" })
	},
}

type logFormatter struct {
	name string
}

func levelToString(l log.Level) string {
	switch l {
	case log.PanicLevel, log.FatalLevel, log.ErrorLevel:
		return "ERR"
	case log.WarnLevel:
		return "WRN"
	case log.InfoLevel:
		return "INF"
	case log.DebugLevel, log.TraceLevel:
		return "DBG"
	}
	return "TRC"
}

func (self *logFormatter) Format(e *log.Entry) ([]byte, error) {
	msg := fmt.Sprintf("%3s [%-10s] %s\n", levelToString(e.Level), self.name, e.Message)
	return []byte(msg), nil
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

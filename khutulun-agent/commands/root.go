package commands

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/tliron/kutil/logging"
	"github.com/tliron/kutil/logging/simple"
	"github.com/tliron/kutil/terminal"
	"github.com/tliron/kutil/util"
)

var logTo string
var verbose int
var colorize string
var simpleLog bool

func init() {
	rootCommand.PersistentFlags().BoolVarP(&terminal.Quiet, "quiet", "q", false, "suppress output")
	rootCommand.PersistentFlags().StringVarP(&logTo, "log", "l", "", "log to file (defaults to stderr)")
	rootCommand.PersistentFlags().CountVarP(&verbose, "verbose", "v", "add a log verbosity level (can be used twice)")
	rootCommand.PersistentFlags().StringVarP(&colorize, "colorize", "z", "true", "colorize output (boolean or \"force\")")
	rootCommand.PersistentFlags().BoolVarP(&simpleLog, "simple", "s", false, "simple log")
}

var rootCommand = &cobra.Command{
	Use:   toolName,
	Short: "Khutulun agent",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		err := terminal.ProcessColorizeFlag(colorize)
		util.FailOnError(err)
		if simpleLog {
			backend := simple.NewBackend()
			backend.Format = SimpleFormat
			logging.SetBackend(backend)
		}
		if logTo == "" {
			if terminal.Quiet {
				verbose = -4
			}
			logging.Configure(verbose, nil)
		} else {
			logging.Configure(verbose, &logTo)
		}
	},
}

func Execute() {
	err := rootCommand.Execute()
	util.FailOnError(err)
}

// simple.FormatFunc signature
func SimpleFormat(message string, id []string, level logging.Level, colorize bool) string {
	var builder strings.Builder

	simple.FormatLevel(&builder, level, false)
	builder.WriteRune(' ')
	simple.FormatID(&builder, id)
	builder.WriteRune(' ')
	builder.WriteString(message)

	return builder.String()
}

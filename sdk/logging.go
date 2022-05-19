package sdk

import (
	"strings"

	"github.com/tliron/kutil/logging"
	"github.com/tliron/kutil/logging/simple"
)

func ConfigurePluginLogging(verbosity int) {
	backend := simple.NewBackend()
	backend.Configure(verbosity, nil)
	backend.Format = format
	logging.SetBackend(backend)
}

func format(message string, id []string, level logging.Level, colorize bool) string {
	var builder strings.Builder

	simple.FormatLevel(&builder, level, true)
	builder.WriteRune(' ')
	simple.FormatID(&builder, id)
	builder.WriteRune(' ')
	builder.WriteString(message)

	return builder.String()
}

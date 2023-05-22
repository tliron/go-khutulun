package sdk

import (
	"strings"

	"github.com/tliron/commonlog"
	"github.com/tliron/commonlog/simple"
)

func ConfigureDelegateLogging(verbosity int) {
	backend := simple.NewBackend()
	backend.Configure(verbosity, nil)
	backend.Format = format
	commonlog.SetBackend(backend)
}

func format(message string, id []string, level commonlog.Level, colorize bool) string {
	var builder strings.Builder

	simple.FormatLevel(&builder, level, true)
	builder.WriteRune(' ')
	simple.FormatID(&builder, id)
	builder.WriteRune(' ')
	builder.WriteString(message)

	return builder.String()
}

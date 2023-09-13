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

// ([simple.FormatFunc] signature)
func format(message *commonlog.UnstructuredMessage, name []string, level commonlog.Level, colorize bool) string {
	var builder strings.Builder

	simple.FormatLevel(&builder, level, true)
	builder.WriteRune(' ')
	simple.FormatName(&builder, name)
	builder.WriteRune(' ')
	builder.WriteString(message.Message)

	return builder.String()
}

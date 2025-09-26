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
func format(message *commonlog.LinearMessage, name []string, level commonlog.Level, colorize bool) string {
	var builder strings.Builder

	builder.WriteString(simple.FormatLevel(level, true))
	builder.WriteRune(' ')
	builder.WriteString(strings.Join(name, "."))
	builder.WriteRune(' ')
	builder.WriteString(message.Message)

	return builder.String()
}

package commands

import (
	"github.com/tliron/commonlog"
	"github.com/tliron/go-transcribe"
)

const toolName = "khutulun"

var log = commonlog.GetLogger(toolName)

var clusterName string
var pseudoTerminal bool
var unpack string

func Transcriber() *transcribe.Transcriber {
	return &transcribe.Transcriber{
		Strict:      strict,
		Format:      format,
		ForTerminal: pretty,
		Base64:      base64,
	}
}

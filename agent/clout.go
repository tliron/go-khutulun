package agent

import (
	"io"
	"os"
	"strings"

	"github.com/tliron/commonlog"
	problemspkg "github.com/tliron/go-kutil/problems"
	cloutpkg "github.com/tliron/go-puccini/clout"
	"github.com/tliron/go-puccini/clout/js"
	"github.com/tliron/go-transcribe"
)

func (self *Agent) CoerceClout(clout *cloutpkg.Clout, copy_ bool) (*cloutpkg.Clout, error) {
	coercedClout := clout
	if copy_ {
		var err error
		if coercedClout, err = clout.Copy(); err != nil {
			return nil, err
		}
	}
	problems := problemspkg.NewProblems(nil)
	execContext := js.ExecContext{
		Clout:      coercedClout,
		Problems:   problems,
		URLContext: self.urlContext,
		History:    true,
		Format:     "yaml",
		Strict:     false,
		Pretty:     false,
	}
	execContext.Coerce()
	return coercedClout, problems.ToError(true)
}

func (self *Agent) OpenFile(path string, coerceClout bool) (io.ReadCloser, error) {
	if coerceClout {
		if file, err := os.Open(path); err == nil {
			defer commonlog.CallAndLogError(file.Close, "file close", log)
			if clout, err := cloutpkg.Read(file, "yaml"); err == nil {
				if clout, err = self.CoerceClout(clout, false); err == nil {
					if code, err := transcribe.NewTranscriber().SetIndentSpaces(2).StringifyYAML(clout); err == nil {
						return io.NopCloser(strings.NewReader(code)), nil
					} else {
						return nil, err
					}
				} else {
					return nil, err
				}
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		return os.Open(path)
	}
}

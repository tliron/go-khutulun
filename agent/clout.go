package agent

import (
	"io"
	"os"
	"strings"

	"github.com/tliron/kutil/logging"
	problemspkg "github.com/tliron/kutil/problems"
	"github.com/tliron/kutil/transcribe"
	cloutpkg "github.com/tliron/puccini/clout"
	"github.com/tliron/puccini/clout/js"
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
	js.Coerce(coercedClout, problems, self.urlContext, true, "yaml", false, true)
	return coercedClout, problems.ToError(true)
}

func (self *Agent) OpenFile(path string, coerceClout bool) (io.ReadCloser, error) {
	if coerceClout {
		if file, err := os.Open(path); err == nil {
			defer logging.CallAndLogError(file.Close, "file close", log)
			if clout, err := cloutpkg.Read(file, "yaml"); err == nil {
				if clout, err = self.CoerceClout(clout, false); err == nil {
					if code, err := transcribe.EncodeYAML(clout, "  ", false); err == nil {
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

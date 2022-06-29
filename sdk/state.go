package sdk

import (
	"io"

	"github.com/danjacques/gofslock/fslock"
	"github.com/tliron/kutil/logging"
)

//
// State
//

type State struct {
	RootDir string
}

func NewState(rootDir string) *State {
	return &State{rootDir}
}

//
// LockedReadCloser
//

type LockedReadCloser struct {
	readCloser io.ReadCloser
	lock       fslock.Handle
}

// io.Reader interface
func (self *LockedReadCloser) Read(p []byte) (n int, err error) {
	return self.readCloser.Read(p)
}

// io.Closer interface
func (self *LockedReadCloser) Close() error {
	logging.CallAndLogError(self.readCloser.Close, "close", stateLog)
	return self.lock.Unlock()
}

//
// LockedWriteCloser
//

type LockedWriteCloser struct {
	writeCloser io.WriteCloser
	lock        fslock.Handle
}

// io.Writer interface
func (self *LockedWriteCloser) Write(p []byte) (n int, err error) {
	return self.writeCloser.Write(p)
}

// io.Closer interface
func (self *LockedWriteCloser) Close() error {
	logging.CallAndLogError(self.writeCloser.Close, "close", stateLog)
	return self.lock.Unlock()
}

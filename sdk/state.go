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
	ReadCloser io.ReadCloser
	Lock       fslock.Handle
}

// io.Reader interface
func (self *LockedReadCloser) Read(p []byte) (n int, err error) {
	return self.ReadCloser.Read(p)
}

// io.Closer interface
func (self *LockedReadCloser) Close() error {
	logging.CallAndLogError(self.ReadCloser.Close, "close", stateLog)
	return self.Lock.Unlock()
}

//
// LockedWriteCloser
//

type LockedWriteCloser struct {
	WriteCloser io.WriteCloser
	Lock        fslock.Handle
}

// io.Writer interface
func (self *LockedWriteCloser) Write(p []byte) (n int, err error) {
	return self.WriteCloser.Write(p)
}

// io.Closer interface
func (self *LockedWriteCloser) Close() error {
	logging.CallAndLogError(self.WriteCloser.Close, "close", stateLog)
	return self.Lock.Unlock()
}

package sdk

import (
	contextpkg "context"
	"errors"
	"fmt"
	"io"
	execpkg "os/exec"

	"github.com/tliron/commonlog"
	"github.com/tliron/go-khutulun/api"
	"github.com/tliron/go-kutil/exec"
	"github.com/tliron/go-kutil/util"
	"google.golang.org/grpc/codes"
	statuspkg "google.golang.org/grpc/status"
)

func NewCommand(start *api.Interaction_Start, log commonlog.Logger) *exec.Command {
	command := exec.NewCommand()

	if start.PseudoTerminal {
		command.PseudoTerminal = new(exec.Size)
		if start.InitialSize != nil {
			log.Debugf("pseudo-terminal size: %d, %d", start.InitialSize.Width, start.InitialSize.Height)
			command.PseudoTerminal.Width = uint(start.InitialSize.Width)
			command.PseudoTerminal.Height = uint(start.InitialSize.Height)
		}
	}

	cmd := start.Command
	if len(cmd) == 0 {
		// Default to bash
		cmd = []string{"/bin/bash"}
		if command.PseudoTerminal != nil {
			// We need to force interactive mode for bash
			cmd = append(cmd, "-i")
		}
	}

	command.Name = cmd[0]
	if len(cmd) > 1 {
		command.Args = cmd[1:]
	}
	command.Environment = start.Environment

	return command
}

func StartCommand(command *exec.Command, server GRPCInteractor, log commonlog.Logger) error {
	context, cancel := contextpkg.WithCancel(contextpkg.Background())
	defer cancel()

	log.Infof("starting command: %s %s", command.Name, util.JoinQuote(command.Args, " "))

	process, err := command.Start(context)
	if err != nil {
		return statuspkg.Error(codes.Aborted, err.Error())
	}
	defer process.Close()

	// Listen to stdout and stderr
	go func() {
		for {
			select {
			case buffer, ok := <-process.Stdout:
				if !ok {
					log.Debug("stdout closed")
					return
				}
				log.Debugf("stdout: %q", buffer)
				server.Send(&api.Interaction{
					Stream: api.Interaction_STDOUT,
					Bytes:  buffer,
				})

			case buffer, ok := <-process.Stderr:
				if !ok {
					log.Debug("stderr closed")
					return
				}
				log.Debugf("stderr: %q", buffer)
				server.Send(&api.Interaction{
					Stream: api.Interaction_STDERR,
					Bytes:  buffer,
				})
			}
		}
	}()

	// Listen to client
	go func() {
		for {
			if interaction, err := server.Recv(); err == nil {
				if interaction.Start != nil {
					command.Stop(errors.New("received more than one message with \"start\""))
					return
				}

				switch interaction.Stream {
				case api.Interaction_STDIN:
					if interaction.Bytes != nil {
						log.Debugf("stdin: %q", interaction.Bytes)
						process.Stdin(interaction.Bytes)
					}

				case api.Interaction_SIZE:
					if interaction.Size != nil {
						log.Debugf("size: %d, %d", interaction.Size.Width, interaction.Size.Height)
						process.Resize(uint(interaction.Size.Width), uint(interaction.Size.Height))
					}

				default:
					command.Stop(fmt.Errorf("unsupported stream: %d", interaction.Stream))
					return
				}
			} else {
				if err == io.EOF {
					log.Info("client closed")
					err = nil
				} else {
					if statuspkg.Code(err) == codes.Canceled {
						// We're OK with canceling
						log.Infof("client canceled")
						err = nil
					}
				}
				command.Stop(err)
				return
			}
		}
	}()

	// Wait until done
	err = command.Wait()
	log.Info("interaction ended")
	if err == nil {
		return nil
	} else if err_, ok := err.(*execpkg.ExitError); ok {
		status := statuspkg.New(codes.Aborted, "execution error")
		if status_, err := status.WithDetails(&api.InteractionErrorDetails{
			ExitCode: int32(err_.ExitCode()),
			Stderr:   err_.Stderr,
		}); err == nil {
			status = status_
		} else {
			panic(err)
		}
		return status.Err()
	} else if _, ok := statuspkg.FromError(err); ok {
		return err
	} else {
		return statuspkg.Error(codes.Aborted, err.Error())
	}
}

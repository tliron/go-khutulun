package util

import (
	"errors"
	"fmt"
	"io"

	"github.com/tliron/khutulun/api"
	"github.com/tliron/kutil/exec"
	"github.com/tliron/kutil/logging"
	"google.golang.org/grpc/codes"
	statuspkg "google.golang.org/grpc/status"
)

func NewCommand(interaction *api.Interaction, log logging.Logger) *exec.Command {
	command := exec.NewCommand()

	if interaction.Start.PseudoTerminal {
		command.PseudoTerminal = new(exec.Size)
		if interaction.Stream == api.Interaction_SIZE {
			log.Debugf("pseudo-terminal size: %d, %d", interaction.Width, interaction.Height)
			command.PseudoTerminal.Width = uint(interaction.Width)
			command.PseudoTerminal.Height = uint(interaction.Height)
		}
	}

	cmd := interaction.Start.Command
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
	command.Environment = interaction.Start.Environment

	return command
}

func StartCommand(command *exec.Command, server Interactor, log logging.Logger) error {
	process, err := command.Start()
	if err != nil {
		return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
	}
	defer process.Close()

	// Listen to stdout and stderr
	go func() {
		for {
			select {
			case buffer := <-process.Stdout:
				if buffer == nil {
					log.Debug("stdout closed")
					return
				}
				log.Debugf("stdout: %q", buffer)
				server.Send(&api.Interaction{
					Stream: api.Interaction_STDOUT,
					Bytes:  buffer,
				})

			case buffer := <-process.Stderr:
				if buffer == nil {
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
					command.Stop(errors.New("received more than one \"start\" message"))
					return
				}

				switch interaction.Stream {
				case api.Interaction_STDIN:
					log.Debugf("stdin: %q", interaction.Bytes)
					process.Stdin(interaction.Bytes)

				case api.Interaction_SIZE:
					log.Debugf("size: %d, %d", interaction.Width, interaction.Height)
					process.Resize(uint(interaction.Width), uint(interaction.Height))

				default:
					command.Stop(fmt.Errorf("unsupported stream: %d", interaction.Stream))
					return
				}
			} else {
				if err == io.EOF {
					log.Info("client closed")
					err = nil
				} else {
					if status, ok := statuspkg.FromError(err); ok {
						if status.Code() == codes.Canceled {
							// We're OK with canceling
							log.Infof("client canceled")
							err = nil
						}
					}
				}
				process.Kill()
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
	} else {
		return statuspkg.Errorf(codes.Aborted, "%s", err.Error())
	}
}

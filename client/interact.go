package client

import (
	"fmt"
	"io"

	"github.com/tliron/khutulun/api"
	"github.com/tliron/khutulun/util"
	"github.com/tliron/kutil/exec"
)

func (self *Client) Interact(identifier []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, terminal *exec.Terminal, environment map[string]string, command ...string) error {
	context, cancel := self.newContextWithCancel()
	defer cancel()

	if client, err := self.client.Interact(context); err == nil {
		start := api.Interaction_Start{
			Identifier:  identifier,
			Command:     command,
			Environment: environment,
		}

		interaction := api.Interaction{
			Start: &start,
		}

		if terminal != nil {
			start.PseudoTerminal = true
			if terminal.InitialSize != nil {
				start.InitialSize = &api.Interaction_Size{
					Width:  uint32(terminal.InitialSize.Width),
					Height: uint32(terminal.InitialSize.Height),
				}
			}

			go func() {
				for size := range terminal.Resize {
					if err := client.Send(&api.Interaction{
						Stream: api.Interaction_SIZE,
						Size: &api.Interaction_Size{
							Width:  uint32(size.Width),
							Height: uint32(size.Height),
						},
					}); err != nil {
						log.Errorf("client send: %s", err.Error())
						return
					}
				}
				log.Info("closed resize")
			}()
		}

		if err := client.Send(&interaction); err != nil {
			return util.UnpackError(err)
		}

		// Read and send stdin
		go func() {
			var buffer []byte = make([]byte, 1)
			for {
				if _, err := stdin.Read(buffer); err == nil {
					if err := client.Send(&api.Interaction{
						Stream: api.Interaction_STDIN,
						Bytes:  buffer,
					}); err != nil {
						log.Errorf("client send: %s", err.Error())
						return
					}
				} else {
					if err != io.EOF {
						log.Errorf("stdin read: %s", err.Error())
					}
					return
				}
			}
		}()

		for {
			if interaction, err := client.Recv(); err == nil {
				switch interaction.Stream {
				case api.Interaction_STDOUT:
					if _, err := stdout.Write(interaction.Bytes); err != nil {
						return err
					}

				case api.Interaction_STDERR:
					if _, err := stderr.Write(interaction.Bytes); err != nil {
						return err
					}

				default:
					return fmt.Errorf("unsupported stream: %d", interaction.Stream)
				}
			} else {
				if err == io.EOF {
					break
				} else {
					return util.UnpackError(err)
				}
			}
		}

		return nil
	} else {
		return util.UnpackError(err)
	}
}

func (self *Client) InteractRelay(server api.Conductor_InteractServer, start *api.Interaction_Start) error {
	context, cancel := self.newContextWithCancel()
	defer cancel()

	if client, err := self.client.Interact(context); err == nil {
		return util.InteractRelay(server, client, start, log)
	} else {
		return util.UnpackError(err)
	}
}

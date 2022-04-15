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
				interaction.Stream = api.Interaction_SIZE
				interaction.Width = uint32(terminal.InitialSize.Width)
				interaction.Height = uint32(terminal.InitialSize.Height)
			}

			go func() {
				for size := range terminal.Resize {
					if err := client.Send(&api.Interaction{
						Stream: api.Interaction_SIZE,
						Width:  uint32(size.Width),
						Height: uint32(size.Height),
					}); err != nil {
						log.Errorf("client send: %s", err.Error())
						return
					}
				}
				log.Info("closed resize")
			}()
		}

		if err := client.Send(&interaction); err != nil {
			return err
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
			interaction, err := client.Recv()
			if err != nil {
				if err == io.EOF {
					break
				} else {
					return err
				}
			}

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
		}

		return nil
	} else {
		return err
	}
}

func (self *Client) InteractRelay(server api.Conductor_InteractServer, first *api.Interaction) error {
	context, cancel := self.newContextWithCancel()
	defer cancel()

	if client, err := self.client.Interact(context); err == nil {
		return util.InteractRelay(server, client, first, log)
	} else {
		return err
	}
}

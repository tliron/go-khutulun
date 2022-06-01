package dashboard

import (
	"time"

	clientpkg "github.com/tliron/khutulun/client"
)

func Dashboard(client *clientpkg.Client) error {
	application := NewApplication(client, 1*time.Second)
	err := application.application.Run()
	if application.ticker != nil {
		application.ticker.Stop()
	}
	return err
}

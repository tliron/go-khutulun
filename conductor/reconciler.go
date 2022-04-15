package conductor

import (
	"time"

	"github.com/tliron/khutulun/plugin"
	cloutpkg "github.com/tliron/puccini/clout"
)

const FREQUENCY = 10 * time.Second

//
// Reconciler
//

type Reconciler struct {
	ticker *Ticker
}

func NewReconciler(conductor *Conductor) *Reconciler {
	var self = Reconciler{
		ticker: NewTicker(FREQUENCY, conductor.Reconcile, reconcilerLog),
	}
	return &self
}

func (self *Reconciler) Start() {
	self.ticker.Start()
}

func (self *Reconciler) Stop() {
	self.ticker.Stop()
}

func (self *Conductor) Reconcile() {
	self.lock.Lock()
	defer self.lock.Unlock()

	if artifacts, err := self.ListArtifacts("", "clout"); err == nil {
		for _, artifact := range artifacts {
			if clout, err := self.GetClout(artifact.Namespace, artifact.Name, true); err == nil {
				self.reconcileRunnables(clout)
			} else {
				reconcilerLog.Errorf("%s", err.Error())
			}
		}
	} else {
		reconcilerLog.Errorf("%s", err.Error())
	}
}

func (self *Conductor) reconcileRunnables(clout *cloutpkg.Clout) {
	containers := self.getResources(clout, "runnable")
	if len(containers) == 0 {
		return
	}

	name := "runnable.podman"
	command := self.getArtifactFile("common", "plugin", name)

	go func() {
		client := plugin.NewRunnableClient(name, command)
		defer client.Close()

		if runnable, err := client.Runnable(); err == nil {
			for _, container := range containers {
				if err := runnable.Instantiate(container); err != nil {
					log.Errorf("instantiate error: %s", err.Error())
				}
			}
		} else {
			log.Errorf("plugin error: %s", err.Error())
		}
	}()
}

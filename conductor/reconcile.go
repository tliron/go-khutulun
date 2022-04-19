package conductor

import (
	"github.com/tliron/khutulun/plugin"
	cloutpkg "github.com/tliron/puccini/clout"
)

func (self *Conductor) Reconcile() {
	if bundles, err := self.ListBundles("", "clout"); err == nil {
		for _, bundle := range bundles {
			if clout, err := self.GetClout(bundle.Namespace, bundle.Name, true); err == nil {
				self.reconcileRunnables(clout)
			} else {
				reconcileLog.Errorf("%s", err.Error())
			}
		}
	} else {
		reconcileLog.Errorf("%s", err.Error())
	}
}

func (self *Conductor) reconcileRunnables(clout *cloutpkg.Clout) {
	containers := self.getResources(clout, "runnable")
	if len(containers) == 0 {
		return
	}

	go func() {
		var runnable plugin.Runnable
		for _, container := range containers {
			instantiate := true
			if self.cluster != nil {
				if self.cluster.cluster.LocalNode().Name != container.Host {
					instantiate = false
				}
			}

			if instantiate {
				if runnable == nil {
					name := "runnable.podman"
					command := self.getBundleMainFile("common", "plugin", name)
					client := plugin.NewRunnableClient(name, command)
					defer client.Close()
					var err error
					if runnable, err = client.Runnable(); err != nil {
						log.Errorf("plugin: %s", err.Error())
						return
					}
				}

				if err := runnable.Instantiate(container); err != nil {
					log.Errorf("instantiate: %s", err.Error())
				}
			}
		}
	}()
}

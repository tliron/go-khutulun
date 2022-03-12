package conductor

import (
	"time"

	"github.com/tliron/khutulun/plugin"
	"github.com/tliron/kutil/ard"
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

func (self *Conductor) getResources(clout *cloutpkg.Clout, type_ string) []Container {
	var containers []Container
	for _, vertex := range clout.Vertexes {
		if capabilities, ok := ard.NewNode(vertex.Properties).Get("capabilities").StringMap(false); ok {
			for _, capability := range capabilities {
				if types, ok := ard.NewNode(capability).Get("types").StringMap(false); ok {
					if _, ok := types["cloud.puccini.khutulun::Container"]; ok {
						var container Container
						if container.name, ok = ard.NewNode(capability).Get("properties").Get("image").Get("name").String(false); !ok {
							container.name, _ = ard.NewNode(vertex.Properties).Get("name").String(false)
						}
						container.source, _ = ard.NewNode(capability).Get("properties").Get("image").Get("source").String(false)
						container.createArguments, _ = ard.NewNode(capability).Get("properties").Get("image").Get("create-arguments").StringList(false)
						containers = append(containers, container)
					}
				}
			}
		}
	}
	return containers
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

		if run, err := client.Runnable(); err == nil {
			for _, container := range containers {
				if err := run.Instantiate(container.ToConfig()); err != nil {
					log.Errorf("instantiate error: %s", err.Error())
					return
				}
			}
		} else {
			log.Errorf("plugin error: %s", err.Error())
		}
	}()
}

type Container struct {
	name            string
	source          string
	createArguments []string
	ports           []Port
}

type Port struct {
	external int64
	internal int64
}

func (self Container) ToConfig() ard.StringMap {
	return ard.StringMap{
		"name":            self.name,
		"source":          self.source,
		"createArguments": self.createArguments,
	}
}

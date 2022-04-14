package conductor

//
// Server
//

type Server struct {
	conductor  *Conductor
	grpc       *GRPC
	cluster    *Cluster
	http       *HTTP
	reconciler *Reconciler
}

func NewServer(conductor *Conductor) *Server {
	return &Server{conductor: conductor}
}

func (self *Server) Start(cluster bool, grpc bool, http bool, reconciler bool) error {
	if cluster {
		self.cluster = NewCluster()
		if err := self.cluster.Start(); err != nil {
			return err
		}
	}

	if grpc {
		self.grpc = NewGRPC(self.conductor, self.cluster)
		if err := self.grpc.Start(); err != nil {
			if self.cluster != nil {
				self.cluster.Stop()
			}
			return err
		}
	}

	if http {
		var err error
		if self.http, err = NewHTTP(self.conductor); err == nil {
			if err := self.http.Start(); err != nil {
				if self.grpc != nil {
					self.grpc.Stop()
				}
				if self.cluster != nil {
					self.cluster.Stop()
				}
				return err
			}
		} else {
			return err
		}
	}

	if reconciler {
		self.reconciler = NewReconciler(self.conductor)
		self.reconciler.Start()
	}

	return nil
}

func (self *Server) Stop() error {
	var err error

	if self.reconciler != nil {
		self.reconciler.Stop()
	}

	if self.http != nil {
		if err_ := self.http.Stop(); err_ != nil {
			err = err_
		}
	}

	if self.grpc != nil {
		self.grpc.Stop()
	}

	if self.cluster != nil {
		if err_ := self.cluster.Stop(); err_ != nil {
			err = err_
		}
	}

	return err
}

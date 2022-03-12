package conductor

import (
	contextpkg "context"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/tliron/kutil/logging/sink"
	"google.golang.org/grpc"

	"github.com/tliron/khutulun/api"
	_ "github.com/tliron/khutulun/web"
)

//
// Server
//

type Server struct {
	conductor    *Conductor
	grpcServer   *grpc.Server
	httpServer   *http.Server
	reconciler   *Reconciler
	cluster      *memberlist.Memberlist
	clusterQueue *memberlist.TransmitLimitedQueue
}

func NewServer(conductor *Conductor) *Server {
	return &Server{conductor: conductor}
}

func (self *Server) Start(grpc bool, cluster bool, http bool, reconciler bool) error {
	if grpc {
		if err := self.startGrpc(); err != nil {
			return err
		}
	}

	if cluster {
		if err := self.startCluster(); err != nil {
			return err
		}
	}

	if http {
		if err := self.startHttp(); err != nil {
			return err
		}
	}

	if reconciler {
		self.startReconciler()
	}

	return nil
}

func (self *Server) Stop() error {
	var err error

	self.stopReconciler()

	if err_ := self.stopHttp(); err_ != nil {
		err = err_
	}

	if err_ := self.stopCluster(); err_ != nil {
		err = err_
	}

	self.stopGrpc()

	return err
}

func (self *Server) startCluster() error {
	config := memberlist.DefaultLocalConfig()
	config.Name, _ = os.Hostname()
	config.Delegate = self
	config.Events = sink.NewMemberlistEventLog(clusterLog)

	clusterLog.Notice("starting memberlist")
	var err error
	self.cluster, err = memberlist.Create(config)
	return err
}

func (self *Server) stopCluster() error {
	if self.cluster == nil {
		return nil
	}

	err := self.cluster.Leave(time.Second * 5)
	self.cluster.Shutdown()
	return err
}

func (self *Server) startGrpc() error {
	self.grpcServer = grpc.NewServer()
	api.RegisterConductorServer(self.grpcServer, NewGRPC(self.conductor))

	if listener, err := net.Listen("tcp", ":8181"); err == nil {
		grpcLog.Noticef("starting server on: %s", listener.Addr().String())
		go func() {
			self.grpcServer.Serve(listener)
		}()
		return nil
	} else {
		return err
	}
}

func (self *Server) stopGrpc() {
	if self.grpcServer == nil {
		return
	}

	self.grpcServer.Stop()
}

func (self *Server) startHttp() error {
	if http_, err := NewHTTP(self.conductor); err == nil {
		self.httpServer = &http.Server{
			Handler: http_.handler,
		}
	} else {
		return err
	}

	if listener, err := net.Listen("tcp", ":8182"); err == nil {
		httpLog.Noticef("starting server on: %s", listener.Addr().String())
		go func() {
			self.httpServer.Serve(listener)
		}()
		return nil
	} else {
		return err
	}
}

func (self *Server) stopHttp() error {
	if self.httpServer == nil {
		return nil
	}

	context, cancel := contextpkg.WithTimeout(contextpkg.Background(), 5*time.Second)
	defer cancel()

	return self.httpServer.Shutdown(context)
}

func (self *Server) startReconciler() {
	self.reconciler = NewReconciler(self.conductor)
	self.reconciler.Start()
}

func (self *Server) stopReconciler() {
	if self.reconciler == nil {
		return
	}

	self.reconciler.Stop()
}

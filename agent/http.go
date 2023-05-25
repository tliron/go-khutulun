package agent

import (
	contextpkg "context"
	"net"
	"net/http"
	"time"

	fspkg "github.com/rakyll/statik/fs"
	_ "github.com/tliron/khutulun/web"
	"github.com/tliron/kutil/transcribe"
	"github.com/tliron/kutil/util"
)

//
// HTTP
//

type HTTP struct {
	Protocol string
	Address  string
	Port     int

	httpServer *http.Server
	mux        *http.ServeMux
	agent      *Agent
}

func NewHTTP(agent *Agent, protocol string, address string, port int) (*HTTP, error) {
	self := HTTP{
		Protocol: protocol,
		Address:  address,
		Port:     port,
		mux:      http.NewServeMux(),
		agent:    agent,
	}

	if filesystem, err := fspkg.New(); err == nil {
		self.mux.Handle("/", http.FileServer(filesystem))
	} else {
		return nil, err
	}

	self.mux.HandleFunc("/api/namespace/list", self.listNamespaces)
	self.mux.HandleFunc("/api/package/list", self.listPackages)
	self.mux.HandleFunc("/api/resource/list", self.listResources)
	self.mux.HandleFunc("/api/host/list", self.listHosts)

	self.httpServer = &http.Server{
		Handler: self.mux,
	}

	return &self, nil
}

func (self *HTTP) Start() error {
	if listener, err := net.Listen(self.Protocol, util.JoinIPAddressPort(self.Address, self.Port)); err == nil {
		httpLog.Noticef("starting server on %s", listener.Addr().String())
		go func() {
			if err := self.httpServer.Serve(listener); err != nil {
				if err == http.ErrServerClosed {
					httpLog.Info("server closed")
				} else {
					httpLog.Errorf("%s", err.Error())
				}
			}
		}()
		return nil
	} else {
		return err
	}
}

func (self *HTTP) Stop() error {
	context, cancel := contextpkg.WithTimeout(contextpkg.Background(), 5*time.Second)
	defer cancel()

	return self.httpServer.Shutdown(context)
}

func (self *HTTP) listNamespaces(writer http.ResponseWriter, request *http.Request) {
	if namespaces, err := self.agent.state.ListNamespaces(); err == nil {
		transcribe.WriteJSON(namespaces, writer, "")
	} else {
		writer.WriteHeader(500)
	}
}

func (self *HTTP) listPackages(writer http.ResponseWriter, request *http.Request) {
	namespace := request.URL.Query().Get("namespace")
	type_ := request.URL.Query().Get("type")
	if type_ != "" {
		if identifiers, err := self.agent.state.ListPackages(namespace, type_); err == nil {
			transcribe.WriteJSON(identifiers, writer, "")
		} else {
			writer.WriteHeader(500)
		}
	} else {
		writer.WriteHeader(400)
	}
}

func (self *HTTP) listResources(writer http.ResponseWriter, request *http.Request) {
	namespace := request.URL.Query().Get("namespace")
	service := request.URL.Query().Get("service")
	type_ := request.URL.Query().Get("type")
	if type_ != "" {
		if resources, err := self.agent.ListResources(contextpkg.TODO(), namespace, service, type_); err == nil {
			transcribe.WriteJSON(resources, writer, "")
		} else {
			writer.WriteHeader(500)
		}
	} else {
		writer.WriteHeader(400)
	}
}

func (self *HTTP) listHosts(writer http.ResponseWriter, request *http.Request) {
	if self.agent.gossip != nil {
		transcribe.WriteJSON(self.agent.gossip.ListHosts(), writer, "")
	} else {
		writer.WriteHeader(500)
	}
}

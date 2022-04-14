package conductor

import (
	contextpkg "context"
	"net"
	"net/http"
	"time"

	fspkg "github.com/rakyll/statik/fs"
	_ "github.com/tliron/khutulun/web"
	"github.com/tliron/kutil/format"
)

//
// HTTP
//

type HTTP struct {
	conductor  *Conductor
	httpServer *http.Server
	handler    *http.ServeMux
}

func NewHTTP(conductor *Conductor) (*HTTP, error) {
	self := HTTP{
		conductor: conductor,
		handler:   http.NewServeMux(),
	}

	if fs, err := fspkg.New(); err == nil {
		self.handler.Handle("/", http.FileServer(fs))
	} else {
		return nil, err
	}

	self.handler.HandleFunc("/api/namespace/list", self.listNamespaces)
	self.handler.HandleFunc("/api/artifact/list", self.listArtifacts)
	self.handler.HandleFunc("/api/resource/list", self.listResources)

	self.httpServer = &http.Server{
		Handler: self.handler,
	}

	return &self, nil
}

func (self *HTTP) Start() error {
	if listener, err := net.Listen("tcp", ":8182"); err == nil {
		httpLog.Noticef("starting server on: %s", listener.Addr().String())
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
	if namespaces, err := self.conductor.ListNamespaces(); err == nil {
		format.WriteJSON(namespaces, writer, "")
	} else {
		writer.WriteHeader(500)
	}
}

func (self *HTTP) listArtifacts(writer http.ResponseWriter, request *http.Request) {
	namespace := request.URL.Query().Get("namespace")
	type_ := request.URL.Query().Get("type")
	if type_ != "" {
		if artifacts, err := self.conductor.ListArtifacts(namespace, type_); err == nil {
			format.WriteJSON(artifacts, writer, "")
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
		if resources, err := self.conductor.ListResources(namespace, service, type_); err == nil {
			format.WriteJSON(resources, writer, "")
		} else {
			writer.WriteHeader(500)
		}
	} else {
		writer.WriteHeader(400)
	}
}

package conductor

import (
	contextpkg "context"
	"fmt"
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
	port       int
	httpServer *http.Server
	mux        *http.ServeMux
	conductor  *Conductor
}

func NewHTTP(conductor *Conductor) (*HTTP, error) {
	self := HTTP{
		port:      8182,
		mux:       http.NewServeMux(),
		conductor: conductor,
	}

	if fs, err := fspkg.New(); err == nil {
		self.mux.Handle("/", http.FileServer(fs))
	} else {
		return nil, err
	}

	self.mux.HandleFunc("/api/namespace/list", self.listNamespaces)
	self.mux.HandleFunc("/api/bundle/list", self.listBundles)
	self.mux.HandleFunc("/api/resource/list", self.listResources)

	self.httpServer = &http.Server{
		Handler: self.mux,
	}

	return &self, nil
}

func (self *HTTP) Start() error {
	if listener, err := net.Listen("tcp", fmt.Sprintf(":%d", self.port)); err == nil {
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

func (self *HTTP) listBundles(writer http.ResponseWriter, request *http.Request) {
	namespace := request.URL.Query().Get("namespace")
	type_ := request.URL.Query().Get("type")
	if type_ != "" {
		if bundles, err := self.conductor.ListBundles(namespace, type_); err == nil {
			format.WriteJSON(bundles, writer, "")
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

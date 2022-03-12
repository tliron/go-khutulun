package conductor

import (
	"net/http"

	fspkg "github.com/rakyll/statik/fs"
	"github.com/tliron/kutil/format"
)

//
// HTTP
//

type HTTP struct {
	conductor *Conductor
	handler   *http.ServeMux
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

	return &self, nil
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

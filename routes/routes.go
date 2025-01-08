package routes

import (
	"net/http"

	"github.com/Falokut/storage_service/controller"

	"github.com/Falokut/go-kit/http/endpoint"
	"github.com/Falokut/go-kit/http/router"
)

type Router struct {
	Files controller.Files
}

func (r Router) InitRoutes(wrapper endpoint.Wrapper) *router.Router {
	mux := router.New()
	for _, desc := range endpointDescriptors(r) {
		mux.Handler(desc.Method, desc.Path, wrapper.Endpoint(desc.Handler))
	}

	return mux
}

type EndpointDescriptor struct {
	Method  string
	Path    string
	Handler any
}

func endpointDescriptors(r Router) []EndpointDescriptor {
	return []EndpointDescriptor{
		{
			Method:  http.MethodPost,
			Path:    "/file/:category",
			Handler: r.Files.UploadFile,
		},
		{
			Method:  http.MethodGet,
			Path:    "/file/:category/:filename",
			Handler: r.Files.GetFile,
		},
		{
			Method:  http.MethodDelete,
			Path:    "/file/:category/:filename",
			Handler: r.Files.DeleteFile,
		},
		{
			Method:  http.MethodGet,
			Path:    "/file/:category/:filename/exist",
			Handler: r.Files.IsFileExist,
		},
	}
}

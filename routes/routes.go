package routes

import (
	"net/http"

	"storage-service/controller"

	"github.com/Falokut/go-kit/cluster"
	"github.com/Falokut/go-kit/http/endpoint"
	"github.com/Falokut/go-kit/http/router"
)

type Router struct {
	Files controller.Files
}

func (r Router) Handler(wrapper endpoint.Wrapper) *router.Router {
	mux := router.New()
	for _, desc := range EndpointDescriptors(r) {
		mux.Handler(desc.HttpMethod, desc.Path, wrapper.Endpoint(desc.Handler))
	}

	return mux
}

func EndpointDescriptors(r Router) []cluster.EndpointDescriptor {
	return []cluster.EndpointDescriptor{
		{
			HttpMethod: http.MethodPost,
			Path:       "/file/:category",
			Handler:    r.Files.UploadFile,
		},
		{
			HttpMethod: http.MethodPost,
			Path:       "/file/:category/:filename",
			Handler:    r.Files.UploadFile,
		},
		{
			HttpMethod: http.MethodGet,
			Path:       "/file/:category/:filename",
			Handler:    r.Files.GetFile,
		},
		{
			HttpMethod: http.MethodDelete,
			Path:       "/file/:category/:filename",
			Handler:    r.Files.DeleteFile,
		},
		{
			HttpMethod: http.MethodGet,
			Path:       "/file/:category/:filename/exist",
			Handler:    r.Files.IsFileExist,
		},
		{
			HttpMethod: http.MethodPost,
			Path:       "/file/:category/:filename/commit",
			Handler:    r.Files.Commit,
		},
		{
			HttpMethod: http.MethodPost,
			Path:       "/file/:category/:filename/rollback",
			Handler:    r.Files.Rollback,
		},
	}
}

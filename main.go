package main

import (
	"storage-service/assembly"
	"storage-service/conf"
	"storage-service/routes"

	"github.com/Falokut/go-kit/bootstrap"
	"github.com/Falokut/go-kit/shutdown"
)

var (
	version = "1.0.0"
)

// @title						storage-service
// @version					1.0.0
// @description				Сервис для заказа еды
// @BasePath					/api/storage-service
//
// @securityDefinitions.apikey	Bearer
// @in							header
// @name						Authorization
// @description				Type "Bearer" followed by a space and JWT token.
//
//go:generate swag init --parseDependency
//go:generate rm -f docs/swagger.json docs/docs.go
func main() {
	boot := bootstrap.New(version, conf.Remote{}, routes.EndpointDescriptors(routes.Router{}))
	app := boot.App
	logger := app.Logger()

	assembly, err := assembly.New(boot)
	if err != nil {
		boot.Fatal(err)
	}

	app.AddRunners(assembly.Runners()...)
	app.AddClosers(assembly.Closers()...)

	shutdown.On(func() {
		logger.Info(app.Context(), "starting shutdown")
		app.Shutdown()
		logger.Info(app.Context(), "shutdown completed")
	})

	err = app.Run()
	if err != nil {
		boot.Fatal(err)
	}
}

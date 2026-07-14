package main

import (
	"log"
	"os"

	"github.com/webdevtedxuniversitasairlangga/middlewares"
	"github.com/webdevtedxuniversitasairlangga/modules/auth"
	"github.com/webdevtedxuniversitasairlangga/modules/todo"
	"github.com/webdevtedxuniversitasairlangga/providers"
	"github.com/common-nighthawk/go-figure"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func run(server *gin.Engine) {
	port := os.Getenv("GOLANG_PORT")
	if port == "" {
		port = "8888"
	}

	var serve string
	if os.Getenv("APP_ENV") == "localhost" {
		serve = "0.0.0.0:" + port
	} else {
		serve = ":" + port
	}

	myFigure := figure.NewColorFigure("TEDxUniversitasAirlangga", "doom", "red", true)
	myFigure.Print()

	err := server.Run(serve)
	if err != nil {
		log.Fatalf("error running server: %s", err)
	}
}

func main() {
	var (
		injector = do.New()
	)

	providers.RegisterDependencies(injector)

	server := gin.Default()
	server.Use(middlewares.CORSMiddleware())

	v1 := server.Group("api/v1")
	{
		auth.RegisterRoutes(v1, injector)
		todo.RegisterRoutes(v1, injector)
	}

	run(server)
}
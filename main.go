package main

import (
	"fmt"
	"net/http"

	"github.com/fossyy/filekeeper/middleware"
	"github.com/fossyy/filekeeper/routes"
	"github.com/fossyy/filekeeper/utils"
)

func main() {
	serverAddr := fmt.Sprintf("%s:%s", utils.Getenv("SERVER_HOST"), utils.Getenv("SERVER_PORT"))
	server := http.Server{
		Addr:    serverAddr,
		Handler: middleware.Handler(routes.SetupRoutes()),
	}

	fmt.Printf("Listening on http://%s\n", serverAddr)
	err := server.ListenAndServe()
	if err != nil {
		return
	}
}

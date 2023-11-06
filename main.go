package main

import (
	"log"
	"net/http"

	"github.com/rosricard/ribbitDeviceManager/api"
	"github.com/rosricard/ribbitDeviceManager/db"
)

func main() {
	db.ConnectDatabase()

	r := api.SetupRouter()

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal("Failed to start the server:", err)
	}
}

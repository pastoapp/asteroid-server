package main

import (
	odb "asteroid-api/internal/orbitdb"
	"asteroid-api/internal/routes"
	"context"
	"github.com/gin-gonic/gin"
	"log"
)

// default settings
var (
	ipfsURL    = "http://localhost:5001"
	orbitDbDir = "./data/orbitdb"
)

// main is the entry point of the program
func main() {
	// main database context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// create a new orbitdb instance
	cancelODB, err := odb.InitializeOrbitDB(ipfsURL, orbitDbDir)
	defer cancelODB() // cancel the orbitdb context

	// ODB in PoC uses only one database: "default"
	defaultDB, err := odb.OpenDatabase(ctx, "default")
	if err != nil {
		log.Print(err)
	}

	// gin server
	r := gin.Default()

	// /ping endpoint
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// Initialise the auth middleware
	//   protects the /notes endpoint
	routes.InitAuth(r, defaultDB)

	// Initialise User Route Module
	routes.InitUsers(r, defaultDB)

	// run on port 3000
	err = r.Run(":3000")

	// print errors, if there are any with the webserver
	if err != nil {
		log.Panicf("Error starting server: %v", err)
	}
}

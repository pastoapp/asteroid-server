package main

import (
	"context"
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/itsjamie/gin-cors"
	odb "gitlab.gwdg.de/v.mattfeld/asteroid-server/internal/orbitdb"
	"gitlab.gwdg.de/v.mattfeld/asteroid-server/internal/routes"

	"log"
	"os"
	"time"
)

// default settings
var (
	ipfsURL    string
	orbitDbDir string
)

// parse cli flags
func init() {
	flag.StringVar(&ipfsURL, "ipfs-url", "http://localhost:5001", "IPFS URL")
	flag.StringVar(&orbitDbDir, "orbitdb-dir", "./data/orbitdb", "OrbitDB directory")
}

// main is the entry point of the program
func main() {
	// parse cli flags
	flag.Parse()

	// verify orbitdb dir exists
	if _, err := os.Stat(orbitDbDir); os.IsNotExist(err) {
		log.Printf("OrbitDB directory does not exist: %v\n", err)
		// create orbitdb dir
		err = os.MkdirAll(orbitDbDir, 0755)
		if err != nil {
			log.Panicf("Error creating OrbitDB directory: %v\n", err)
			return
		}
	}

	// main database context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log.Println("IPFS URL:", ipfsURL)
	log.Println("OrbitDB directory:", orbitDbDir)
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

	// cors
	// In PoC, we set the Cross-Origin policies to allow all.

	r.Use(cors.Middleware(cors.Config{
		Origins:         "*",
		Methods:         "GET, PUT, POST, DELETE",
		RequestHeaders:  "Origin, Authorization, Content-Type",
		ExposedHeaders:  "",
		MaxAge:          50 * time.Second,
		Credentials:     false,
		ValidateHeaders: false,
	}))

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

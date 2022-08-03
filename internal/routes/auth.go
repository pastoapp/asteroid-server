package routes

import (
	jwt2 "asteroid-api/internal/jwt"
	"asteroid-api/internal/orbitdb"
	"github.com/gin-gonic/gin"
	"log"
)

// init auth middleware module
func init() {
	log.SetPrefix("[routes/auth] ")
}

// InitAuth takes the current gin-instance and ODB to create the corresponding protected routes
func InitAuth(router *gin.Engine, db *orbitdb.Database) {
	authMiddleware, err := jwt2.AsteroidJWTMiddleware()
	if err != nil {
		log.Fatal("Error creating auth middleware")
		return
	}

	// init auth middleware
	err = authMiddleware.MiddlewareInit()

	if err != nil {
		log.Fatal("Error initializing auth middleware" + err.Error())
		return
	}

	// Auth management
	router.POST("/login", authMiddleware.LoginHandler)
	router.GET("/refresh_token", authMiddleware.RefreshHandler)

	// attach protected routes
	auth := router.Group("/notes")
	auth.Use(authMiddleware.MiddlewareFunc())
	{
		// protecting the /notes endpoint
		notes := Notes{
			DB:     db,
			RGroup: auth,
		}
		auth.POST("/", notes.Create)
		auth.GET("/:id", notes.Find)
	}
}

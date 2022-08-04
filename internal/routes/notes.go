package routes

import (
	"github.com/docker/distribution/uuid"
	"github.com/gin-gonic/gin"
	jwt2 "gitlab.gwdg.de/v.mattfeld/asteroid-server/internal/jwt"
	"gitlab.gwdg.de/v.mattfeld/asteroid-server/internal/middleware/note"
	"gitlab.gwdg.de/v.mattfeld/asteroid-server/internal/orbitdb"
	"net/http"
)

// Notes is a reference to the notes database
type Notes struct {
	DB     *orbitdb.Database
	RGroup *gin.RouterGroup
}

// createReq is the request body for creating a new note
type createReq struct {
	Note string `json:"note" binding:"required"`
}

// Create uses the request body to create a new note on authenticated routes
func (n Notes) Create(c *gin.Context) {
	// get user from JWT
	user := getUserFromJWT(c)
	if user == nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "unable to find user in token"})
		return
	}

	// get request body
	var body createReq
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// convert uid to uuid
	uid, err := uuid.Parse(user.ID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// create note
	newNote, err := note.NewNote(body.Note, uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// response
	c.JSON(http.StatusOK, gin.H{
		"id":   newNote.ID.String(),
		"uid":  newNote.UID.String(),
		"note": newNote.Data, // Optional for production: Add encryption
	})
}

// getUserFromJWT gets the user from the JWT. It's a helper function.
func getUserFromJWT(context *gin.Context) *jwt2.User {
	// get user from JWT
	tokenUser, exists := context.Get(jwt2.IdentityKey)

	if !exists {
		context.JSON(http.StatusBadRequest, gin.H{"message": "unable to find user in token"})
		return nil
	}

	return tokenUser.(*jwt2.User)
}

// Find returns a note by id on authenticated routes.
func (n Notes) Find(context *gin.Context) {
	// get user from JWT
	user := getUserFromJWT(context)

	if user == nil {
		context.JSON(http.StatusBadRequest, gin.H{"message": "unable to find user in token"})
		return
	}

	// get note id from url
	id := context.Param("id")

	if id == "" {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": "id is required",
		})
		return
	}

	// parse node id
	noteID, err := uuid.Parse(id)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// note result from database
	find, err := note.GetNote(noteID)
	// note owner ID
	nodeUID := find.UID.String()

	// check if user is the owner of the note
	if nodeUID != user.ID {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": "user does not own note",
		})
		return
	}

	// respond
	context.JSON(http.StatusOK, gin.H{
		"id":   find.ID.String(),
		"uid":  nodeUID,
		"note": find.Data,
	})
}

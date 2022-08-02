package routes

import (
	"asteroid-api/internal/middleware/note"
	"asteroid-api/internal/orbitdb"
	"github.com/docker/distribution/uuid"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Notes struct {
	DB     *orbitdb.Database
	RGroup *gin.RouterGroup
}

var notes Notes

type createReq struct {
	UID  string `json:"uid" binding:"required"`
	Note string `json:"note" binding:"required"`
}

func InitNotes(router *gin.Engine, db *orbitdb.Database) *Notes {
	group := router.Group("/notes")
	notes := Notes{
		DB:     db,
		RGroup: group,
	}
	group.POST("/", notes.Create)
	group.GET("/:id", notes.Find)

	return &notes
}

func (n Notes) Create(c *gin.Context) {
	var body createReq
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uid, err := uuid.Parse(body.UID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newNote, err := note.NewNote(body.Note, uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":   newNote.ID.String(),
		"uid":  newNote.UID.String(),
		"note": newNote.Data,
	})
}

func (n Notes) Find(context *gin.Context) {
	id := context.Param("id")

	if id == "" {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": "id is required",
		})
		return
	}

	noteID, err := uuid.Parse(id)

	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	find, err := note.GetNote(noteID)

	context.JSON(http.StatusOK, gin.H{
		"id":   find.ID.String(),
		"uid":  find.UID.String(),
		"note": find.Data,
	})
}

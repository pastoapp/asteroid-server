package routes

import (
	"github.com/gin-gonic/gin"
	"gitlab.gwdg.de/v.mattfeld/asteroid-server/internal/middleware/user"
	"gitlab.gwdg.de/v.mattfeld/asteroid-server/internal/orbitdb"
	"log"
	"net/http"
	"os"
	"strings"
)

// init runs on module initialization
func init() {
	log.SetPrefix("[routes/users] ")
}

// Users is the route module struct
type Users struct {
	DB     *orbitdb.Database
	RGroup *gin.RouterGroup
}

var users Users

// InitUsers takes the current gin-instance and ODB to create the corresponding routes
func InitUsers(router *gin.Engine, db *orbitdb.Database) *Users {
	group := router.Group("/users")
	users = Users{
		DB:     db,
		RGroup: group,
	}
	group.POST("/", users.Create)
	group.GET("/:id", users.Find)

	return &users
}

// Find is a GET endpoint, finding a user with a corresponding id/path.
func (u Users) Find(context *gin.Context) {

	id := context.Param("id")

	if id == "" {
		context.JSON(400, gin.H{
			"error": "id is required",
		})
		return
	}

	find, err := user.Find(id)

	if err != nil {
		context.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	context.JSON(200, u.response(&find))
}

// Create is a POST endpoint at /users/ taking a public key in PEM
// via Form-File-Upload and returning a new users object.
func (u Users) Create(c *gin.Context) {
	// create a temporary directory to read the file from
	tmpDir := os.TempDir()
	file, err := c.FormFile("file")
	if err != nil {
		log.Fatal(err)
	}

	log.Println(file.Filename)

	// saving the uploaded file in
	if err := c.SaveUploadedFile(file, tmpDir+file.Filename); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Unable to save the file",
		})
		return
	}

	// read keyfile
	rawFileContents, err := os.ReadFile(tmpDir + file.Filename)
	fileContents := string(rawFileContents)

	// fix line endings
	fileContents = strings.ReplaceAll(fileContents, "\r", "")

	// validate the public key by checking on some attributes
	// ends with .pem
	if !strings.HasSuffix(file.Filename, ".pem") {
		c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{
			"message": "filename must end in .pem",
		})
		return
	}

	// contains BEGIN RSA PUBLIC KEY
	if !strings.HasPrefix(fileContents, "-----BEGIN RSA PUBLIC KEY-----") {
		c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{
			"message": "Malformed file",
		})
		return
	}

	// create user
	newUser, err := user.NewUser(fileContents, false)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	// remove file from tempdir
	err = os.Remove(tmpDir + file.Filename)
	if err != nil {
		log.Println("Cannot remove TempDir while creating user")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "User creation error on server. Please try again later",
		})
		return
	}

	// response with full user object
	c.JSON(200, u.response(&newUser))
}

// response is an object, returning a JSON-parsed version of the user.User object.
func (_ Users) response(u *user.User) gin.H {
	return gin.H{
		"_id":       u.ID.String(),
		"publicKey": u.PublicKey,
		"nonce":     u.Nonce,
		"createdAt": u.CreatedAt,
		"updatedAt": u.UpdatedAt,
		"notes":     u.Notes,
	}
}

// For production, add more CRUD methods here...

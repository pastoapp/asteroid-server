package jwt

import (
	"encoding/base64"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"gitlab.gwdg.de/v.mattfeld/asteroid-server/internal/middleware/user"
	"log"
	"time"
)

func init() {
	log.SetPrefix("[jwt/jwt] ")
}

// Login binds JSON input to this struct
type Login struct {
	ID        string `json:"id" form:"id" binding:"required"`
	Signature string `json:"signature" form:"signature" binding:"required"`
}

// User is the user struct for the JWT
type User struct {
	ID        string
	Notes     string
	PublicKey string
}

// Authenticator is a function that takes a context and returns an identity and/or an error
func Authenticator(c *gin.Context) (interface{}, error) {
	var login Login

	// bind the JSON input to the Login struct
	if err := c.ShouldBindJSON(&login); err != nil {
		return "", jwt.ErrMissingLoginValues
	}

	uid := login.ID

	// decode the signature
	sgntr, err := base64.StdEncoding.DecodeString(login.Signature)

	if uid == "" {
		return "", jwt.ErrFailedAuthentication
	}
	if err != nil {
		log.Println(err)
		return "", jwt.ErrFailedAuthentication
	}

	log.Printf("Authenticating user: %s\n", uid)

	// find the user
	usr, err := user.Find(uid)
	if err != nil {
		log.Println("Error finding user:", err)
		return nil, err
	}

	// check signature against user's nonce
	err = usr.VerifyUser(sgntr)

	// update nonce
	err = usr.RefreshNonce()
	if err != nil {
		log.Println("Error refreshing nonce:", err)
		return nil, err
	}

	// if no error, return the user
	return &User{
		ID:        usr.ID.String(),
		Notes:     usr.Notes,
		PublicKey: usr.PublicKey,
	}, nil
}

// IdentityKey is the key used to store the identity key in the GinJWTMiddleware.
var IdentityKey = "_id"

// AsteroidJWTMiddleware is the middleware for the JWT
func AsteroidJWTMiddleware() (*jwt.GinJWTMiddleware, error) {
	secret := []byte("secret key") // TODO: change to env variable for production!
	return jwt.New(&jwt.GinJWTMiddleware{
		Realm:      "main",
		Key:        secret,
		Timeout:    time.Hour * 24 * 7,
		MaxRefresh: time.Hour * 24 * 7,
		Authenticator: func(c *gin.Context) (interface{}, error) {
			return Authenticator(c)
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			// Production: if the user is an admin etc., return false
			log.Println("Authorizer called...")
			return true
		},
		IdentityKey: IdentityKey,
		IdentityHandler: func(context *gin.Context) interface{} {
			claims := jwt.ExtractClaims(context)
			return &User{
				ID: claims[IdentityKey].(string),
			}
		},
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			// JWT Payload
			if v, ok := data.(*User); ok {
				return jwt.MapClaims{
					IdentityKey: v.ID,
				}
			}
			return jwt.MapClaims{}
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		TimeFunc:   time.Now,
		CookieName: "Asteroid-JWT",
		// Add more cookie security settings for production ...
	})
}

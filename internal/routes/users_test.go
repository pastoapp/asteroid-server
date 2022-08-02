package routes

import (
	"asteroid-api/internal/middleware/user"
	"asteroid-api/internal/orbitdb"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"github.com/gin-gonic/gin"

	"net/http"
	"net/http/httptest"
	"testing"
)

func performRequest(r http.Handler, method, path string, body gin.H) *httptest.ResponseRecorder {
	bd, err := json.Marshal(body)

	if err != nil {
		panic(err)
	}

	var req *http.Request

	if method == "GET" {
		req, _ = http.NewRequest(method, path, nil)
	} else {
		req, _ = http.NewRequest(method, path, bytes.NewBuffer(bd))
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func setupRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"hi": "mom",
		})
	})

	return r
}

func TestUserRoutes(t *testing.T) {
	// gin server
	r := setupRouter()

	privateK, _ := rsa.GenerateKey(rand.Reader, 4096)
	pubkBytes := x509.MarshalPKCS1PublicKey(&privateK.PublicKey)
	pubkPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubkBytes,
	})

	publicKey := string(pubkPEM)

	cancelFunc, err := orbitdb.InitializeOrbitDB("http://localhost:5001", t.TempDir())
	if err != nil {
		t.Fatalf("Error initializing OrbitDB: %v", err)
	}
	defer cancelFunc()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	userDB, err := orbitdb.OpenDatabase(ctx, "users")

	if err != nil {
		t.Fatalf("Error opening database: %v", err)
	}

	InitUsers(r, userDB)

	t.Run("should test the /ping endpoint", func(t *testing.T) {
		w := performRequest(r, "GET", "/ping", nil)

		if w.Code != http.StatusOK {
			t.Errorf("Expected response code to be %d, but was %d", http.StatusOK, w.Code)
		}
	})

	t.Run("should create a user on /", func(t *testing.T) {
		w := performRequest(r, "POST", "/users/", gin.H{"publicKey": publicKey})

		if w.Code != http.StatusOK {
			t.Errorf("Expected response code to be %d, but was %d. %v\n", http.StatusOK, w.Code, w.Body)
		}

		t.Log(w.Body)
	})

	t.Run("should get a user on /:id", func(t *testing.T) {
		usr, err := user.NewUser(publicKey, false)
		if err != nil {
			t.Fatalf("Error creating user: %v", err)
		}
		w := performRequest(r, "GET", "/users/"+usr.ID.String(), nil)
		if w.Code != http.StatusOK {
			t.Errorf("Expected response code to be %d, but was %d. %v\n", http.StatusOK, w.Code, w.Body)
		}
		t.Log(w.Body)
	})

}

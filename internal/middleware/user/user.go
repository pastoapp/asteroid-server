package user

import (
	"asteroid-api/internal/orbitdb"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"github.com/docker/distribution/uuid"
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

// User entity, holding a note in the OrbitDB.
type User struct {
	ID        uuid.UUID
	PublicKey string
	Nonce     string
	IsAdmin   bool
	CreatedAt int64
	UpdatedAt int64
	Notes     string
}

// init runs at module initialization.
func init() {
	log.SetPrefix("[middleware/user/user] ")
}

// NewUser creates a new user entry in the ODB
func NewUser(publicKey string, isAdmin bool) (User, error) {
	nonce, err := GenerateNonce()
	if err != nil {
		log.Println("Failed to generate Nonce")
		return User{}, err
	}

	user := User{
		ID:        uuid.Generate(),
		PublicKey: publicKey,
		// TODO: REGENERATE NONCE EVERY TIME AN AUTH SUCCESSFULLY HAPPENS
		// base64 encoded nonce
		Nonce:     nonce,
		IsAdmin:   isAdmin,
		CreatedAt: time.Now().UTC().Unix(),
		UpdatedAt: time.Now().UTC().Unix(),
		Notes:     "",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := orbitdb.OpenDatabase(ctx, "default")

	if err != nil {
		log.Println("Could not open user database")
		return User{}, err
	}

	defer func(db *orbitdb.Database) {
		err := db.Close()
		if err != nil {
			log.Printf("Could not close user database %v\n", user)
		}
	}(db)

	resp, err := db.Create(gin.H{
		"id":        user.ID.String(),
		"publicKey": user.PublicKey,
		"nonce":     user.Nonce,
		"isAdmin":   user.IsAdmin,
		"createdAt": user.CreatedAt,
		"updatedAt": user.UpdatedAt,
	}, nil)

	if err != nil {
		log.Println("Could not create user")
		return User{}, err
	}

	_id := resp["_id"].(string)

	newID, err := uuid.Parse(_id)

	if err != nil {
		log.Println("Could not parse UUID")
		return User{}, err
	}

	return User{
		ID:        newID,
		PublicKey: user.PublicKey,
		Nonce:     user.Nonce,
		IsAdmin:   user.IsAdmin,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// GenerateNonce generates a SHA256 with the size of 64 bits. The user signs this nonce in order
// to get authenticated
func GenerateNonce() (string, error) {
	key := [64]byte{}
	_, err := rand.Read(key[:])
	if err != nil {
		log.Println("Failed to generate random key")
		return "", err
	}

	msgHash := sha256.New()
	_, err = msgHash.Write(key[:])
	if err != nil {
		log.Println("Failed to hash key")
		return "", err
	}
	return base64.StdEncoding.EncodeToString(msgHash.Sum(nil)), nil
}

// Login
func (u User) Login() (string, error) {
	// TODO: implement
	// TODO: return JWT
	return "", fmt.Errorf("not implemented")
}

// RefreshNonce updates the user nonce
func (u User) RefreshNonce() error {
	nonce, err := GenerateNonce()
	if err != nil {
		log.Println("Failed to generate Nonce")
		return err
	}
	u.Nonce = nonce
	return nil
}

func (u User) VerifyUser(signature []byte) error {

	block, _ := pem.Decode([]byte(u.PublicKey))
	if block == nil {
		return fmt.Errorf("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKCS1PublicKey(block.Bytes)

	if err != nil {
		return fmt.Errorf("failed to parse DER encoded public key: %s\n", err.Error())
	}

	nonce, err := base64.StdEncoding.DecodeString(u.Nonce)

	if err != nil {
		return fmt.Errorf("failed to decode nonce: %s\n", err.Error())
	}

	return rsa.VerifyPSS(pub, crypto.SHA256, nonce, signature, nil)
}

// Find finds a user with the corresponding user id.
func Find(key string) (User, error) {
	// use a basic context. For production, consider using context.WithTimeout
	ctx := context.Background()

	// chose the database to operate from
	db, err := orbitdb.OpenDatabase(ctx, "default")

	if err != nil {
		log.Println("Unable to open the default database")
		return User{}, err
	}

	// open the document database
	docs, err := orbitdb.Client.Docs(ctx, db.Address.String(), nil)

	if err != nil {
		log.Println("Unable to open the document database")
		return User{}, err
	}

	// according to
	// https://github.com/berty/go-orbit-db/blob/07c8fbab7657926d723aaee4ba50c58c88b7959b/tests/persistence_test.go#L24
	// -1 is infinity
	//
	// Loading the database
	err = docs.Load(ctx, -999)
	if err != nil {
		log.Println("Error loading users from database")
		return User{}, err
	}

	// Query an item from the database, having the key of the user ID.
	get, err := docs.Get(ctx, key, nil)
	if err != nil {
		log.Println("Cannot GET user from Database")
		return User{}, err
	}

	// Print the user to logs
	log.Println(get)

	// get should be a list with 1 entry, containing the encoded user
	if len(get) != 1 {
		log.Println("Invalid response length")
		return User{}, err
	}

	// type-casting the user
	var rawUser = get[0].(map[string]interface{})

	// extract the user id
	id, err := uuid.Parse(rawUser["_id"].(string))

	// extract the data
	data := rawUser["data"].(string)

	rawUserData, err := orbitdb.UnmarshalItem(data)
	if err != nil {
		log.Println("Error parsing user data to appropiate format")
		return User{}, err
	}

	// finalize parsing the complete User
	user := parseRawUserData(id, rawUserData.(map[string]interface{}))
	return *user, nil
}

// UpdateNotes updates the user notes with a corresponding note id
func UpdateNotes(uid, noteId string) (*User, error) {
	// find the existing user
	u, err := Find(uid)
	if err != nil {
		return nil, err
	}

	// add note to user object
	u.Notes = u.Notes + ";" + noteId

	// use a basic context. For production, consider using context.WithTimeout
	ctx := context.Background()

	// chose the database to operate from
	db, err := orbitdb.OpenDatabase(ctx, "default")

	if err != nil {
		log.Println("Unable to open the default database")
		return nil, err
	}

	// open the document database
	docs, err := orbitdb.Client.Docs(ctx, db.Address.String(), nil)

	if err != nil {
		log.Println("Unable to open the document database")
		return nil, err
	}

	// according to
	// https://github.com/berty/go-orbit-db/blob/07c8fbab7657926d723aaee4ba50c58c88b7959b/tests/persistence_test.go#L24
	// -1 is infinity
	//
	// Loading the database
	err = docs.Load(ctx, -999)
	if err != nil {
		log.Println("Error loading users from database")
		return nil, err
	}

	// Update the user
	_, err = db.Update(u.ID.String(), gin.H{
		"id":        u.ID.String(),
		"publicKey": u.PublicKey,
		"nonce":     u.Nonce,
		"isAdmin":   u.IsAdmin,
		"createdAt": u.CreatedAt,
		"updatedAt": u.UpdatedAt,
		"notes":     u.Notes,
	})

	if err != nil {
		log.Println("Error updating user")
		return nil, err
	}

	return &u, nil
}

// parseRawUserData completes the parsing of a User and returns a reference
func parseRawUserData(id uuid.UUID, raw map[string]interface{}) *User {
	return &User{
		ID:        id,
		PublicKey: raw["publicKey"].(string),
		Nonce:     raw["nonce"].(string),
		IsAdmin:   raw["isAdmin"].(bool),
		CreatedAt: int64(raw["createdAt"].(float64)),
		UpdatedAt: int64(raw["updatedAt"].(float64)),
		Notes:     raw["notes"].(string),
	}
}

package note

import (
	"asteroid-api/internal/middleware/user"
	"asteroid-api/internal/orbitdb"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

func TestNewNotes(t *testing.T) {
	//var currentNote *Note

	privateK, _ := rsa.GenerateKey(rand.Reader, 4096)
	pubkBytes := x509.MarshalPKCS1PublicKey(&privateK.PublicKey)
	pubkPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubkBytes,
	})

	PublicKey := string(pubkPEM)

	cancelFunc, err := orbitdb.InitializeOrbitDB("http://localhost:5001", t.TempDir())

	if err != nil {
		t.Fatalf("Error initializing OrbitDB: %v", err)
	}
	defer cancelFunc()
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	item := "Lorem Ipsum"
	tUser, err := user.NewUser(PublicKey, false)

	if err != nil {
		t.Fatalf("Error creating tUser: %v", err)
	}

	t.Run("Create a note", func(t *testing.T) {
		note, err := NewNote(item, tUser.ID)

		if err != nil {
			t.Fatalf("Error creating note: %v", err)
		}

		if note.Data != item {
			t.Fatalf("Error creating note: %v", err)
		}

		//currentNote = note
	})

	t.Run("Get a note", func(t *testing.T) {
		tNote, err := NewNote(item, tUser.ID)

		if err != nil {
			t.Fatalf("Error creating note: %v", err)
		}

		note, err := GetNote(tNote.ID)

		if err != nil {
			t.Fatalf("Error getting note: %v", err)
		}

		if note.Data != item {
			t.Fatalf("Error getting note: %v", err)
		}
	})
}

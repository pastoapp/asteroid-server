package user

import (
	"asteroid-api/internal/orbitdb"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"github.com/google/uuid"
	"testing"
)

// WARNING: DO NOT USE THOSE KEYS

func TestNewUser(t *testing.T) {
	privateK, _ := rsa.GenerateKey(rand.Reader, 2048)

	//privkBytes := x509.MarshalPKCS1PrivateKey(privateK)
	//privPEM := pem.EncodeToMemory(
	//	&pem.Block{
	//		Type:  "RSA PRIVATE KEY",
	//		Bytes: privkBytes,
	//	},
	//)

	//t.Logf("%v\n", string(privPEM))

	pubkBytes := x509.MarshalPKCS1PublicKey(&privateK.PublicKey)
	pubkPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubkBytes,
	})

	t.Logf("%v\n", string(pubkPEM))

	privKey := x509.MarshalPKCS1PrivateKey(privateK)
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privKey,
	})
	t.Logf("%v\n", string(privPEM))

	PublicKey := string(pubkPEM)
	//PrivateKey := string(&privateK)

	cancelFunc, err := orbitdb.InitializeOrbitDB("http://localhost:5001", t.TempDir())
	if err != nil {
		t.Fatalf("Error initializing OrbitDB: %v", err)
	}
	defer cancelFunc()
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("should create a new user", func(t *testing.T) {
		user, err := NewUser(PublicKey, false)

		if err != nil {
			t.Errorf("an error occurred %v\n", err)
		}

		if user.PublicKey != PublicKey {
			t.Errorf("public key does not match")
		}

		if user.Nonce == "" {
			t.Errorf("no valid Nonce found")
		}

		if user.IsAdmin {
			t.Errorf("user should not be admin")
		}
	})

	t.Run("should create a user and find it", func(t *testing.T) {
		user, err := NewUser(PublicKey, false)
		if err != nil {
			t.Errorf("error creating the user, %v\n", err)
		}

		resp, err := Find(user.ID.String())
		if err != nil {
			t.Errorf("error finding the user %v - %v\n", user, resp)
		}
	})

	t.Run("should verify a user", func(t *testing.T) {

		user, err := NewUser(PublicKey, false)

		if err != nil {
			t.Errorf("error creating the user %v\n", err)
		}

		if user.Nonce == "" {
			t.Errorf("no valid Nonce found")
		}

		nonce, err := base64.StdEncoding.DecodeString(user.Nonce)

		if err != nil {
			t.Errorf("error decoding the nonce %v\n", err)
		}

		sign, err := privateK.Sign(rand.Reader, nonce, &rsa.PSSOptions{
			SaltLength: rsa.PSSSaltLengthAuto,
			Hash:       crypto.SHA256,
		})

		if err != nil {
			t.Errorf("error signing the user %v\n", err)
		}

		err = user.VerifyUser(sign)
		if err != nil {
			t.Errorf("error verifying the user %v\n", err)
		}
	})

	t.Run("should create multiple users and find them", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			user, err := NewUser(PublicKey, false)
			if err != nil {
				t.Errorf("error creating user %v\n", err)
			}
			resp, err := Find(user.ID.String())

			_, err = uuid.Parse(resp.ID.String())
			if err != nil {
				t.Errorf("user could not be queried")
			}
		}

	})
}

package orbitdb

import (
	"berty.tech/go-orbit-db/iface"
	"context"
	"testing"
)

const IpfsApiURL string = "http://localhost:5001"

func TestODBCreation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("Should create an OrbitDB instance", func(t *testing.T) {
		// NOTE: t.TempDir could cause some permission errors on Windows as of Go 1.18.5
		odb, err := NewOrbitDB(ctx, t.TempDir(), IpfsApiURL)

		Client = odb

		if err != nil {
			t.Errorf("Error creating OrbitDB instance: %s", err)
		}

		if odb == nil {
			t.Errorf("OrbitDB instance is nil")
		}

		if Client == nil {
			t.Errorf("OrbitDB client is nil")
		}

		if Client.Identity().ID == "" {
			t.Errorf("OrbitDB instance has no identity")
		}

		if Client.Identity().ID != odb.Identity().ID {
			t.Errorf("OrbitDB instance has no address")
		}
	})
}

func TestOrbitDBInit(t *testing.T) {
	t.Run("should initialize the Client global var", func(t *testing.T) {
		cancel, err := InitializeOrbitDB(IpfsApiURL, t.TempDir())

		defer cancel()

		if err != nil {
			t.Errorf("Error initializing OrbitDB: %s", err)
		}

		if Client == nil {
			t.Errorf("OrbitDB client is nil")
		}

		if Client.Identity().ID == "" {
			t.Errorf("OrbitDB instance has no identity")
		}

	})
}

func TestDocStore(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	odb, _ := NewOrbitDB(ctx, t.TempDir(), IpfsApiURL)

	Client = odb

	t.Run("should create a new document store", func(t *testing.T) {
		docs, err := Client.Docs(ctx, "test-doc", nil)
		defer closeDocs(docs, t)

		if err != nil {
			t.Errorf("Error creating document store: %s", err)
		}

		if docs == nil {
			t.Errorf("Document store is nil")
		}

		if docs.Identity().ID == "" {
			t.Errorf("Document store has no identity")
		}

		if docs.Address().String() == "" {
			t.Errorf("Document store has no address")
		}

		if docs.DBName() != "test-doc" {
			t.Errorf("Document store has wrong name")
		}
	})

	t.Run("should save a new document to the store (global context)", func(t *testing.T) {
		docs, err := Client.Docs(ctx, "test-doc-add", nil)
		defer closeDocs(docs, t)

		if err != nil {
			t.Errorf("Error creating document store: %s", err)
		}
		item := map[string]interface{}{
			"_id": "1",
			"Hi":  "Mom",
		}

		put, err := docs.Put(ctx, item)
		if err != nil {
			t.Errorf("Error saving document: %s", err)
		}

		if *put.GetKey() != "1" {
			t.Errorf("Document key is wrong")
		}

		value := string(put.GetValue())
		if value == "" {
			t.Errorf("Document value is nil")
		}

		result, err := docs.Get(ctx, "1", nil)

		if err != nil {
			t.Errorf("Error getting document: %s", err)
		}

		if result == nil {
			t.Errorf("Document is nil")
		}

		if len(result) == 0 {
			t.Errorf("Response is empty")
		}
	})

	customContext, customCancel := context.WithCancel(ctx)
	defer customCancel()

	t.Run("should put an item into the database with the same ID with a custom context", func(t *testing.T) {
		docs, err := Client.Docs(customContext, "test-doc-add", nil)
		defer closeDocs(docs, t)
		if err != nil {
			t.Errorf("Error creating document store: %s", err)
		}
		item := map[string]interface{}{
			"_id": "1",
			"Hi":  "Mom",
		}

		put, err := docs.Put(customContext, item)
		if err != nil {
			t.Errorf("Error saving document: %s", err)
		}

		if *put.GetKey() != "1" {
			t.Errorf("Document key is wrong")
		}

		value := string(put.GetValue())
		if value == "" {
			t.Errorf("Document value is nil")
		}
	})

	t.Run("should return the previous put items under the same context", func(t *testing.T) {
		docs, err := Client.Docs(customContext, "test-doc-add", nil)
		defer closeDocs(docs, t)

		// -1 loads all entries
		err = docs.Load(customContext, -1)
		if err != nil {
			t.Errorf("Error loading documents: %s", err)
		}

		if err != nil {
			t.Errorf("Error creating document store: %s", err)
		}

		result, err := docs.Get(customContext, "1", nil)

		if err != nil {
			t.Errorf("Error getting document: %s", err)
		}

		if result == nil {
			t.Errorf("Document is nil")
		}

		if len(result) == 0 {
			t.Errorf("Response is empty")
		}
	})
}

func closeDocs(docs iface.DocumentStore, t *testing.T) {
	err := docs.Close()
	if err != nil {
		t.Errorf("Error closing document store: %s", err)
	}
}

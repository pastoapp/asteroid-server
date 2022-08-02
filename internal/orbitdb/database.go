package orbitdb

import (
	"berty.tech/go-orbit-db/address"
	"berty.tech/go-orbit-db/iface"
	"berty.tech/go-orbit-db/stores/operation"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/docker/distribution/uuid"
	"log"
	"time"
)
import "context"

// Database is the main interface for interacting with OrbitDB
type Database struct {
	Store   *iface.DocumentStore
	Name    string
	Address address.Address
}

// DatabaseCreateOptions which handle the creation of a new entry
type DatabaseCreateOptions struct {
	ID string
}

// init runs at initialization of the module
func init() {
	log.SetPrefix("[orbitdb/database] ")
}

// timeout is used to set the timeout for the database operations
var timeout = 10 * time.Duration(time.Second)

// infinite items to return
var infinite = -1

// OpenDatabase creates or opens a database
func OpenDatabase(ctx context.Context, name string) (*Database, error) {
	// Check if the ODB client is initialized
	if Client == nil {
		log.Printf("Client is not initialized")
		return nil, fmt.Errorf("client is not initialized." +
			" Please run orbitdb.InitializeOrbitDB")
	}

	// create a new document-DB
	docs, err := Client.Docs(ctx, name, nil)

	if err != nil {
		log.Printf("Could not open/create database: %v", err)
		return nil, err
	}

	err = docs.Load(ctx, infinite)
	if err != nil {
		log.Printf("Could not load database: %v", err)
		return nil, err
	}

	// return a reference to the document DB
	return &Database{
		Name:    name,
		Store:   &docs,
		Address: docs.Address(),
	}, nil
}

// MarshalItem parses any matching go-lang object into a base64-encoded json string
func MarshalItem(item interface{}) (string, error) {
	b, err := json.Marshal(item)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// UnmarshalItem reverses the process of MarshalItem
func UnmarshalItem(item string) (interface{}, error) {
	b, err := base64.StdEncoding.DecodeString(item)
	if err != nil {
		return nil, err
	}
	var i interface{}
	err = json.Unmarshal(b, &i)
	if err != nil {
		return nil, err
	}
	return i, nil
}

// Create creates a new document in the database
func (d Database) Create(item interface{}, options *DatabaseCreateOptions) (map[string]interface{}, error) {
	ctx := context.Background() //context.WithTimeout(context.Background(), timeout)
	//defer cancel()

	store := *d.Store
	var put operation.Operation
	var err error

	mItem, err := MarshalItem(item)

	if options != nil {
		put, err = store.Put(ctx, map[string]interface{}{
			"_id":  options.ID,
			"data": mItem,
		})
	} else {
		put, err = store.Put(ctx, map[string]interface{}{
			"_id":  uuid.Generate().String(),
			"data": mItem,
		})
	}

	if err != nil {
		log.Printf("Could not create item: %v", err)
		return nil, err
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(put.GetValue(), &m)
	if err != nil {
		log.Printf("Could not unmarshal item: %v", err)
		return nil, err
	}

	return m, nil
}

// Read reads a document from the database
func (d Database) Read(key string) (map[string]interface{}, error) {
	//ctx, cancel := context.WithTimeout(context.Background(), timeout)
	//defer cancel()
	ctx := context.Background()

	store := *d.Store
	err := store.Load(ctx, infinite)

	if err != nil {
		log.Printf("Could not load database: %v", err)
		return nil, err
	}

	get, err := store.Get(ctx, key, nil)

	if err != nil {
		log.Printf("Could not read item: %v", err)
		return nil, err
	}

	// in case more or less than one item is found
	if len(get) != 1 {
		log.Printf("Could not read item: %v", get)
		return nil, fmt.Errorf("more than one item or none are found")
	}

	//log.Printf("Read item: %v", get)
	item := get[0]

	if err != nil {
		log.Printf("Could not unmarshal item: %v", err)
		return nil, err
	}

	return item.(map[string]interface{}), nil
}

func (d Database) ReadAll() []interface{} {
	ctx := context.Background()

	store := *d.Store
	//err := store.Load(ctx, 100)
	//
	//if err != nil {
	//	log.Printf("Could not load database: %v", err)
	//	return nil
	//}

	get, err := store.Get(ctx, "-", &iface.DocumentStoreGetOptions{PartialMatches: true})

	if err != nil {
		log.Printf("Could not read item: %v", err)
		return nil
	}

	return get
}

// Update updates a document in the database, using the corresponding key and the new information, item.
func (d Database) Update(key string, item interface{}) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	store := *d.Store
	err := store.Load(ctx, infinite)
	if err != nil {
		log.Printf("Could not load database: %v", err)
		return nil, err
	}

	// find the item to update
	get, err := store.Get(ctx, key, nil)

	if err != nil {
		log.Printf("Error reading item: %v", err)
		return nil, err
	}

	if len(get) != 1 {
		log.Printf("Cannot find exactly one item with key %s", key)
		return nil, err
	}

	marshalItem, err := MarshalItem(item)
	if err != nil {
		log.Printf("Could not marshal item: %v", err)
		return nil, err
	}

	// update the item
	put, err := store.Put(ctx, map[string]interface{}{
		"_id":  key,
		"data": marshalItem,
	})

	if err != nil {
		log.Printf("Could not create item: %v", err)
		return nil, err
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(put.GetValue(), &m)

	if err != nil {
		log.Printf("Could not unmarshal item: %v", err)
		return nil, err
	}

	return m, nil
}

// Delete deletes a document from the database
func (d Database) Delete(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	store := *d.Store
	_, err := store.Delete(ctx, key)

	if err != nil {
		log.Printf("Could not delete item: %v", err)
		return err
	}

	return nil
}

// Close closes the database
func (d Database) Close() error {
	store := *d.Store
	return store.Close()
}

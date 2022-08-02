package orbitdb

import (
	"reflect"
	"testing"
)
import "context"

func closeDb(db *Database, t *testing.T) {
	err := db.Close()
	if err != nil {
		t.Fatalf("Error closing database: %v", err)
	}
}

func TestNewDatabase(t *testing.T) {
	// NOTE: t.TempDir could cause some permission errors on Windows as of Go 1.18.5
	cancelFunc, err := InitializeOrbitDB("http://localhost:5001", t.TempDir())

	if err != nil {
		t.Fatalf("Error initializing OrbitDB: %v", err)
	}

	defer cancelFunc()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("should create a new database", func(t *testing.T) {
		db, err := OpenDatabase(ctx, "create-test")
		defer closeDb(db, t)

		if db == nil {
			t.Errorf("expected database to be created")
		}

		if err != nil {
			t.Errorf("error creating database: %s", err)
		}

		if db.Name != "create-test" {
			t.Errorf("expected database name to be 'testdb', got %s", db.Name)
		}

		if db.Store == nil {
			t.Errorf("expected database store to be created")
		}

		if len(db.Address.String()) < 10 {
			t.Errorf("expected database address to be set")
		}
	})

	item := map[string]interface{}{"Hi": "mom"}

	t.Run("should create a new item in the database", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		db, err := OpenDatabase(ctx, "rw-global-test")
		defer closeDb(db, t)

		if err != nil {
			t.Errorf("error creating database: %s", err)
		}
		resp, err := db.Create(item, nil)

		if err != nil {
			t.Errorf("error adding item: %s", err)
		}

		if len(resp) == 0 {
			t.Errorf("expected response to have length > 0")
		}

		if resp == nil {
			t.Errorf("expected response to be returned")
		}
	})

	t.Run("should create an item and then read it between closes", func(t *testing.T) {

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		db, err := OpenDatabase(ctx, "rw-global-test")

		if err != nil {
			t.Errorf("error creating database: %s", err)
		}
		resp, err := db.Create(item, nil)

		if err != nil {
			t.Errorf("error adding item: %s", err)
		}

		if len(resp) == 0 {
			t.Errorf("expected response to have length > 0")
		}

		if resp == nil {
			t.Errorf("expected response to be returned")
		}
		closeDb(db, t)
		cancel()

		// read the item
		ctx, cancel = context.WithCancel(context.Background())
		defer cancel()
		db, err = OpenDatabase(ctx, "rw-global-test")

		if err != nil {
			t.Errorf("error creating database: %s", err)
		}

		defer closeDb(db, t)
		key := resp["_id"].(string)

		read, err := db.Read(key)

		if err != nil {
			t.Errorf("error reading item: %s", err)
		}

		if !reflect.DeepEqual(read["data"], resp["data"]) {
			t.Errorf("expected read to be equal to write")
		}

		if !reflect.DeepEqual(read["_id"], resp["_id"]) {
			t.Errorf("expected read to be equal to write")
		}
	})

	t.Run("should delete an item with the specified key", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		db, err := OpenDatabase(ctx, "rw-global-test")
		defer closeDb(db, t)

		if err != nil {
			t.Errorf("error creating database: %s", err)
		}

		m, err := db.Create(item, nil)

		_id := m["_id"].(string)

		err = db.Delete(_id)

		if err != nil {
			t.Errorf("error deleting item: %s", err)
		}

		get, err := db.Read(_id)

		if len(get) != 0 {
			t.Errorf("expected item to be deleted")
		}
	})
}

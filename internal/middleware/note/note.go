package note

import (
	"asteroid-api/internal/middleware/user"
	"asteroid-api/internal/orbitdb"
	"context"
	"github.com/docker/distribution/uuid"
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

// Note is a note entity
type Note struct {
	ID   uuid.UUID
	UID  uuid.UUID
	Data string // Change it to interface{} for production
}

// init is called before main
func init() {
	log.SetPrefix("[middleware/note/note] ")
}

// NewNote creates a new note entry in the ODB
func NewNote(text string, uid uuid.UUID) (*Note, error) {
	note := &Note{
		ID:   uuid.Generate(),
		UID:  uid,
		Data: text,
	}

	// give the note a few seconds to be created
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := orbitdb.OpenDatabase(ctx, "default")

	if err != nil {
		log.Println("Failed to create note")
		return nil, err
	}

	// create the note
	resp, err := db.Create(gin.H{
		"id":   note.ID.String(),
		"data": note.Data,
		"uid":  note.UID.String(),
	}, nil)

	defer func(db *orbitdb.Database) {
		err := db.Close()
		if err != nil {
			log.Printf("Could not close note database %v\n", note)
		}
	}(db)

	if err != nil {
		log.Println("Failed to create note")
		return nil, err
	}

	// extract the note ID from the response
	_id := resp["_id"].(string)

	newID, err := uuid.Parse(_id)

	if err != nil {
		log.Println("Failed to parse note id")
		return nil, err
	}

	// update the user notes
	_, err = user.UpdateNotes(uid.String(), newID.String())

	if err != nil {
		log.Println("Failed to update user notes")
		return nil, err
	}

	// return a new note
	return &Note{
		ID:   newID,
		UID:  uid,
		Data: note.Data,
	}, nil
}

// GetNote returns a note from the ODB
func GetNote(id uuid.UUID) (*Note, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := orbitdb.OpenDatabase(ctx, "default")

	if err != nil {
		log.Println("Failed to open note database")
		return nil, err
	}

	defer func(db *orbitdb.Database) {
		err := db.Close()
		if err != nil {
			log.Printf("Could not close note database %v\n", id)
		}
	}(db)

	// get the note
	resp, err := db.Read(id.String())

	if err != nil {
		log.Println("Failed to get note")
		return nil, err
	}

	// parse the ODB format to a note
	item, err := orbitdb.UnmarshalItem(resp["data"].(string))

	if err != nil {
		return nil, err
	}

	inferred := item.(map[string]interface{})
	uid, err := uuid.Parse(inferred["uid"].(string))

	if err != nil {
		log.Println("Failed to parse note uid")
		return nil, err
	}

	note := &Note{
		ID:   id,
		UID:  uid,
		Data: inferred["data"].(string),
	}

	return note, nil
}

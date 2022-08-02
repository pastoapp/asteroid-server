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

type Note struct {
	ID   uuid.UUID
	UID  uuid.UUID
	Data string // Change it to interface{} for production
}

func init() {
	log.SetPrefix("[middleware/note/note] ")
}

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

	_id := resp["_id"].(string)

	newID, err := uuid.Parse(_id)

	if err != nil {
		log.Println("Failed to parse note id")
		return nil, err
	}
	//userObject, err := user.Find(uid.String())
	//if err != nil {
	//	log.Println("Failed to find user")
	//	return nil, err
	//}
	_, err = user.UpdateNotes(uid.String(), newID.String())
	if err != nil {
		log.Println("Failed to update user notes")
		return nil, err
	}

	return &Note{
		ID:   newID,
		UID:  uid,
		Data: note.Data,
	}, nil
}

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

	resp, err := db.Read(id.String())

	if err != nil {
		log.Println("Failed to get note")
		return nil, err
	}

	item, err := orbitdb.UnmarshalItem(resp["data"].(string))
	if err != nil {
		return nil, err
	}

	log.Println(item, id.String())

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

func GetAllNotes() ([]Note, error) {
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
			log.Printf("Could not close note database %v\n", "")
		}
	}(db)

	all := db.ReadAll()

	log.Println("RETURNS", all)

	return nil, nil
}

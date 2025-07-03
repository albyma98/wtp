/*
Package database is the middleware between the app database and the code. All data (de)serialization (save/load) from a
persistent database are handled here. Database specific logic should never escape this package.

To use this package you need to apply migrations to the database if needed/wanted, connect to it (using the database
data source name from config), and then initialize an instance of AppDatabase from the DB connection.

For example, this code adds a parameter in `webapi` executable for the database data source name (add it to the
main.WebAPIConfiguration structure):

	DB struct {
		Filename string `conf:""`
	}

This is an example on how to migrate the DB and connect to it:

	// Start Database
	logger.Println("initializing database support")
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		logger.WithError(err).Error("error opening SQLite DB")
		return fmt.Errorf("opening SQLite: %w", err)
	}
	defer func() {
		logger.Debug("database stopping")
		_ = db.Close()
	}()

Then you can initialize the AppDatabase and pass it to the api package.
*/
package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
)

// AppDatabase is the high level interface for the DB
type AppDatabase interface {
	GetName() (string, error)
	SetName(name string) error

	// user.go
	CreateUser(uuid string, username string, photoUrl string) error
	GetUserByUUID(uuid string) (User, error)
	SetUserName(uuid string, newUsername string) error
	SetPhotoUrl(uuid string, newPhotoUrl string) error
	SearchUsersByPrefix(prefix string) ([]User, error)
	GetAllUsers() ([]User, error)
	UserExists(uuid string) (bool, error)
	GetPeerData(convID int64, uuidMe string) (User, error)

	// message.go
	CreateMessage(msg Message) (int64, error)
	GetMessageByID(id int64) (Message, error)
	GetMessagesByConversationID(convoID int64) ([]Message, error)
	DeleteMessageByID(id int64, uuidSender string) error
	ForwardMessage(originalMsgID int64, destConversationID int64, senderUUID string) (int64, error)
	GetLastMessage(convID int64) (Message, error)

	// reaction.go
	AddReaction(messageID int64, uuid string, emoji string) error
	RemoveReaction(messageID int64, uuid string) error
	GetReactionsByMessageID(messageID int64) ([]Reaction, error)
	GetReactionsWithUserByMessageID(messageID int64) ([]ReactionWithUser, error)

	// message_status.go
	SetDelivered(uuidUser string, idMessage int64) error
	SetSeen(uuidUser string, idMessage int64) error
	GetMessageStatus(uuidUser string, idMessage int64) (MessageStatus, error)
	GetAllStatusesByMessage(idMessage int64) ([]MessageStatus, error)

	// conversation.go
	CreateDirectConversation(uuid1, uuid2 string) (Conversation, error)
	CreateGroupConversation(creatorUUID string, groupName, groupPhoto *string) (Conversation, error)
	GetConversationsByUser(uuid string) ([]Conversation, error)
	GetLastMessageByConversation(id int64) (Message, error)
	GetDirectConversationBetween(uuid1, uuid2 string) (Conversation, error)
	DeleteConversationIfEmpty(id int64) error
	GetConversationByID(id int64) (Conversation, error)
	SetGroupName(id int64, newName string) error
	SetGroupPhoto(id int64, newPhoto string) error

	// member.go
	AddMember(uuidUser string, idConversation int64) error
	RemoveMember(uuidUser string, idConversation int64) error
	IsMember(uuidUser string, idConversation int64) (bool, error)
	GetMembersByConversation(idConversation int64) ([]string, error)
	GetJoinedAt(uuidUser string, idConversation int64) (string, error)

	Ping() error
}

type appdbimpl struct {
	c *sql.DB
}

// New returns a new instance of AppDatabase based on the SQLite connection `db`.
// `db` is required - an error will be returned if `db` is `nil`.
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building a AppDatabase")
	}

	// Verifica se ci sono già tabelle nel DB
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table';`).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("errore durante controllo struttura DB: %w", err)
	}

	// Se non ci sono tabelle → esegui schema.sql
	if count == 0 {
		log.Printf("Database vuoto, eseguo schema.sql...")

		schemaBytes, err := os.ReadFile("service/database/schema.sql")
		if err != nil {
			return nil, fmt.Errorf("errore lettura schema.sql: %w", err)
		}

		_, err = db.Exec(string(schemaBytes))
		if err != nil {
			return nil, fmt.Errorf("errore esecuzione schema.sql: %w", err)
		}

		log.Printf("schema.sql eseguito con successo")
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		log.Printf("errore nell'abilitare le foreign keys: %v", err)
	}

	return &appdbimpl{
		c: db,
	}, nil
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}

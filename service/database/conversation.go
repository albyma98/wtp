package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
)

type Conversation struct {
	ID                   int64
	IsDirect             bool
	GroupName            *string
	GroupPhoto           *string
	TimestampCreated     string
	TimestampLastMessage string
}

func (db *appdbimpl) CreateDirectConversation(uuid1, uuid2 string) (Conversation, error) {
	// Prendi il timestamp corrente in RFC3339
	now := time.Now().Format(time.RFC3339)

	// Inizia la transazione
	tx, err := db.c.Begin()
	if err != nil {
		return Conversation{}, err
	}

	// Defer con gestione sicura del rollback
	defer func() {
		if rErr := tx.Rollback(); rErr != nil && !errors.Is(rErr, sql.ErrTxDone) {
			log.Printf("rollback fallito: %v", rErr)
		}
	}()

	// Inserisci la conversazione
	res, err := tx.Exec(`
		INSERT INTO conversation (isDirect, timestampCreated, timestampLastMessage)
		VALUES (true, ?, ?)
	`, now, now)
	if err != nil {
		return Conversation{}, err
	}

	conversationID, err := res.LastInsertId()
	if err != nil {
		return Conversation{}, err
	}

	// Aggiungi i due utenti come membri
	_, err = tx.Exec(`
		INSERT INTO member (uuidUser, idConversation, timestampJoined)
		VALUES (?, ?, ?), (?, ?, ?)
	`, uuid1, conversationID, now, uuid2, conversationID, now)
	if err != nil {
		return Conversation{}, err
	}

	// Conferma la transazione
	err = tx.Commit()
	if err != nil {
		return Conversation{}, err
	}

	return Conversation{
		ID:                   conversationID,
		IsDirect:             true,
		TimestampCreated:     now,
		TimestampLastMessage: now,
	}, nil
}

func (db *appdbimpl) CreateGroupConversation(creatorUUID string, groupName, groupPhoto *string) (Conversation, error) {
	tx, err := db.c.Begin()
	if err != nil {
		return Conversation{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Ottieni timestamp corrente
	timestamp := time.Now().Format(time.RFC3339)
	// Inserisci conversazione
	res, err := tx.Exec(`
		INSERT INTO conversation (isDirect, groupName, groupPhoto, timestampCreated, timestampLastMessage)
		VALUES (FALSE, ?, ?, ?, ?)`,
		groupName, groupPhoto, timestamp, timestamp)
	if err != nil {
		return Conversation{}, err
	}

	conversationID, err := res.LastInsertId()
	if err != nil {
		return Conversation{}, err
	}

	// Inserisci il creatore come primo membro
	_, err = tx.Exec(`
		INSERT INTO member (uuidUser, idConversation, timestampJoined)
		VALUES (?, ?, ?)`,
		creatorUUID, conversationID, timestamp)
	if err != nil {
		return Conversation{}, err
	}

	err = tx.Commit()
	if err != nil {
		return Conversation{}, err
	}

	return Conversation{
		ID:                   conversationID,
		IsDirect:             false,
		GroupName:            groupName,
		GroupPhoto:           groupPhoto,
		TimestampCreated:     timestamp,
		TimestampLastMessage: timestamp,
	}, nil
}

func (db *appdbimpl) GetConversationsByUser(uuid string) ([]Conversation, error) {
	rows, err := db.c.Query(`
		SELECT c.id, c.isDirect, c.groupName, c.groupPhoto, c.timestampCreated, c.timestampLastMessage
		FROM conversation c
		JOIN member m ON c.id = m.idConversation
		WHERE m.uuidUser = ?
		ORDER BY c.timestampLastMessage DESC
	`, uuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []Conversation

	for rows.Next() {
		var c Conversation
		err := rows.Scan(&c.ID, &c.IsDirect, &c.GroupName, &c.GroupPhoto, &c.TimestampCreated, &c.TimestampLastMessage)
		if err != nil {
			return nil, err
		}
		conversations = append(conversations, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return conversations, nil
}

func (db *appdbimpl) GetLastMessageByConversation(id int64) (Message, error) {
	var msg Message
	err := db.c.QueryRow(`
		SELECT id, type, content, mediaUrl, timestamp, idConversation, uuidSender, idRepliesTo
		FROM message
		WHERE idConversation = ?
		ORDER BY timestamp DESC
		LIMIT 1
	`, id).Scan(
		&msg.ID, &msg.Type, &msg.Content, &msg.MediaUrl, &msg.Timestamp,
		&msg.IDConversation, &msg.UUIDSender, &msg.IDRepliesTo,
	)
	return msg, err
}

func (db *appdbimpl) GetDirectConversationBetween(uuid1, uuid2 string) (Conversation, error) {
	var conv Conversation
	err := db.c.QueryRow(`
		SELECT c.id, c.isDirect, c.groupName, c.groupPhoto, c.timestampCreated, c.timestampLastMessage
		FROM conversation c
		JOIN member m1 ON c.id = m1.idConversation
		JOIN member m2 ON c.id = m2.idConversation
		WHERE c.isDirect = TRUE
		AND m1.uuidUser = ? AND m2.uuidUser = ?
	`, uuid1, uuid2).Scan(&conv.ID, &conv.IsDirect, &conv.GroupName, &conv.GroupPhoto, &conv.TimestampCreated, &conv.TimestampLastMessage)

	return conv, err
}

func (db *appdbimpl) DeleteConversationIfEmpty(id int64) error {
	var count int
	err := db.c.QueryRow(`
		SELECT COUNT(*)
		FROM member
		WHERE idConversation = ?
	`, id).Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		_, err := db.c.Exec(`DELETE FROM conversation WHERE id = ?`, id)
		return err
	}

	return nil
}

func (db *appdbimpl) UpdateGroupConversation(id int64, newName, newPhoto *string) error {
	// Verifica che la conversazione sia un gruppo
	var isDirect bool
	err := db.c.QueryRow(`SELECT isDirect FROM conversation WHERE id = ?`, id).Scan(&isDirect)
	if err != nil {
		return err
	}
	if isDirect {
		return fmt.Errorf("impossibile modificare nome/foto: conversazione diretta")
	}

	_, err = db.c.Exec(`
		UPDATE conversation
		SET groupName = ?, groupPhoto = ?
		WHERE id = ?
	`, newName, newPhoto, id)

	return err
}

func (db *appdbimpl) SetGroupName(id int64, newName string) error {
	// Verifica che la conversazione sia un gruppo
	var isDirect bool
	err := db.c.QueryRow(`SELECT isDirect FROM conversation WHERE id = ?`, id).Scan(&isDirect)
	if err != nil {
		return err
	}
	if isDirect {
		return fmt.Errorf("impossibile modificare nome/foto: conversazione diretta")
	}
	_, err = db.c.Exec(`
		UPDATE conversation
		SET groupName = ?
		WHERE id = ?
	`, newName, id)
	return err
}

func (db *appdbimpl) SetGroupPhoto(id int64, newPhoto string) error {
	// Verifica che la conversazione sia un gruppo
	var isDirect bool
	err := db.c.QueryRow(`SELECT isDirect FROM conversation WHERE id = ?`, id).Scan(&isDirect)
	if err != nil {
		return err
	}
	if isDirect {
		return fmt.Errorf("impossibile modificare nome/foto: conversazione diretta")
	}

	_, err = db.c.Exec(`
		UPDATE conversation
		SET groupPhoto = ?
		WHERE id = ?
	`, newPhoto, id)

	return err
}

func (db *appdbimpl) GetConversationByID(id int64) (Conversation, error) {
	var c Conversation
	err := db.c.QueryRow(`
		SELECT id, isDirect, groupName, groupPhoto, timestampCreated, timestampLastMessage
		FROM conversation
		WHERE id = ?
	`, id).Scan(&c.ID, &c.IsDirect, &c.GroupName, &c.GroupPhoto, &c.TimestampCreated, &c.TimestampLastMessage)
	return c, err
}

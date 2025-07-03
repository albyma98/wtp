package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
)

type Message struct {
	ID              int64
	Type            string
	Content         string
	MediaUrl        *string
	Timestamp       string
	IDConversation  int64
	UUIDSender      string
	IDRepliesTo     *int64
	IDForwardedFrom *int64             `json:"idForwardedFrom"`
	Reactions       []ReactionWithUser `json:"reactions"`
}

// 1. CreateMessage
func (db *appdbimpl) CreateMessage(msg Message) (int64, error) {
	timestamp := time.Now().Format(time.RFC3339)
	result, err := db.c.Exec(
		`INSERT INTO message (type, content, mediaUrl, timestamp, idRepliesTo, idForwardedFrom, idConversation ,uuidSender)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		msg.Type, msg.Content, msg.MediaUrl, timestamp, msg.IDRepliesTo, msg.IDForwardedFrom, msg.IDConversation, msg.UUIDSender,
	)

	if err != nil {
		return 0, err
	}
	// 3. Recupera ID del messaggio appena creato
	messageID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Aggiorna il timestamp dell'ultima modifica nella conversazione
	_, err = db.c.Exec(`
            UPDATE conversation
            SET timestampLastMessage = ?
            WHERE id = ?
    `, timestamp, msg.IDConversation)
	if err != nil {
		return 0, err
	}

	// 4. Inserisci record in message_status
	err = db.InsertMessageStatusForRecipients(messageID, msg.IDConversation, msg.UUIDSender)
	if err != nil {
		return 0, err
	}

	return messageID, nil
}

func (db *appdbimpl) InsertMessageStatusForRecipients(messageID int64, conversationID int64, senderUUID string) error {
	// Verifica se è una conversazione diretta
	var isDirect bool
	err := db.c.QueryRow(`SELECT isDirect FROM conversation WHERE id = ?`, conversationID).Scan(&isDirect)
	if err != nil {
		return err
	}

	// Recupera tutti gli utenti tranne il mittente
	rows, err := db.c.Query(`
		SELECT uuidUser FROM member WHERE idConversation = ? AND uuidUser != ?
	`, conversationID, senderUUID)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Inizia transazione
	tx, err := db.c.Begin()
	if err != nil {
		return err
	}

	// Defer per Rollback sicuro
	defer func() {
		if rErr := tx.Rollback(); rErr != nil && !errors.Is(rErr, sql.ErrTxDone) {
			log.Printf("rollback fallito: %v", rErr)
		}
	}()

	// Prepara l'inserimento dei messageStatus
	stmt, err := tx.Prepare(`
		INSERT INTO messageStatus (uuidUser, idMessage, delivered, seen)
		VALUES (?, ?, true, false)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for rows.Next() {
		var uuid string
		if err := rows.Scan(&uuid); err != nil {
			return err
		}
		if _, err := stmt.Exec(uuid, messageID); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// Commit finale
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// 2. GetMessageByID
func (db *appdbimpl) GetMessageByID(id int64) (Message, error) {

	var msg Message
	err := db.c.QueryRow(
		`SELECT id, type, content, mediaUrl, timestamp, idConversation, uuidSender, idRepliesTo, idForwardedFrom FROM message WHERE id = ?`, id,
	).Scan(
		&msg.ID, &msg.Type, &msg.Content, &msg.MediaUrl, &msg.Timestamp, &msg.IDConversation, &msg.UUIDSender, &msg.IDRepliesTo, &msg.IDForwardedFrom,
	)
	return msg, err
}

// 3. GetMessagesByConversationID
func (db *appdbimpl) GetMessagesByConversationID(convoID int64) ([]Message, error) {
	rows, err := db.c.Query(`SELECT id, type, content, mediaUrl, timestamp, idConversation, uuidSender, idRepliesTo, idForwardedFrom FROM message WHERE idConversation = ? ORDER BY timestamp ASC`, convoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.ID, &msg.Type, &msg.Content, &msg.MediaUrl, &msg.Timestamp, &msg.IDConversation, &msg.UUIDSender, &msg.IDRepliesTo, &msg.IDForwardedFrom)
		if err != nil {
			return nil, err
		}
		// Recupera le reazioni associate a questo messaggio con i rispettivi utenti
		reactions, err := db.GetReactionsWithUserByMessageID(msg.ID)
		if err != nil {
			return nil, err
		}
		msg.Reactions = reactions

		messages = append(messages, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return messages, nil
}

// 4. DeleteMessageByID (solo se lo ha mandato l'utente)
func (db *appdbimpl) DeleteMessageByID(id int64, uuidSender string) error {
	result, err := db.c.Exec(`DELETE FROM message WHERE id = ? AND uuidSender = ?`, id, uuidSender)
	if err != nil {
		return err
	}
	af, err := result.RowsAffected()
	if err == nil && af == 0 {
		return fmt.Errorf("nessun messaggio eliminato (ID non esistente o UUID non corrispondente)")
	}
	return err
}

func (db *appdbimpl) ForwardMessage(originalMsgID int64, destConversationID int64, senderUUID string) (int64, error) {
	// 1. Recupera i dati del messaggio originale
	var msgType, content string
	var mediaUrl *string

	err := db.c.QueryRow(`
		SELECT type, content, mediaUrl
		FROM message
		WHERE id = ?
	`, originalMsgID).Scan(&msgType, &content, &mediaUrl)
	if err != nil {
		return 0, fmt.Errorf("messaggio originale non trovato")
	}

	// 2. Verifica che il sender sia membro della conversazione di destinazione
	var count int
	err = db.c.QueryRow(`
		SELECT COUNT(*) FROM member
		WHERE idConversation = ? AND uuidUser = ?
	`, destConversationID, senderUUID).Scan(&count)
	if err != nil {
		return 0, err
	}
	if count == 0 {
		return 0, fmt.Errorf("utente non autorizzato")
	}

	finalContent := content

	// 4. Costruisci oggetto Message
	newMsg := Message{
		Type:            msgType,
		Content:         finalContent,
		MediaUrl:        mediaUrl,
		Timestamp:       "", // gestito dentro CreateMessage
		IDConversation:  destConversationID,
		UUIDSender:      senderUUID,
		IDRepliesTo:     nil, // non è una reply
		IDForwardedFrom: &originalMsgID,
	}

	// 5. Crea il nuovo messaggio nel DB
	newID, err := db.CreateMessage(newMsg)
	if err != nil {
		return 0, err
	}

	return newID, nil
}

func (db *appdbimpl) GetLastMessage(convID int64) (Message, error) {
	var msg Message

	err := db.c.QueryRow(`
		SELECT id, type, content, mediaUrl, timestamp, idConversation, uuidSender, idRepliesTo, idForwardedFrom
		FROM message
		WHERE idConversation = ?
		ORDER BY timestamp DESC
		LIMIT 1
	`, convID).Scan(
		&msg.ID,
		&msg.Type,
		&msg.Content,
		&msg.MediaUrl,
		&msg.Timestamp,
		&msg.IDConversation,
		&msg.UUIDSender,
		&msg.IDRepliesTo,
		&msg.IDForwardedFrom,
	)

	if err != nil {
		return Message{}, err
	}

	return msg, nil
}

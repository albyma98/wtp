package database

import (
	"time"
)

type Member struct {
	UUIDUser        string
	IDConversation  int64
	TimestampJoined string
}

func (db *appdbimpl) AddMember(uuidUser string, idConversation int64) error {
	timestamp := time.Now().Format(time.RFC3339)
	_, err := db.c.Exec(`
		INSERT INTO member(uuidUser, idConversation, timestampJoined)
		VALUES (?, ?, ?);
	`, uuidUser, idConversation, timestamp)
	return err
}

func (db *appdbimpl) RemoveMember(uuidUser string, idConversation int64) error {
	_, err := db.c.Exec(`
		DELETE FROM member
		WHERE uuidUser = ? AND idConversation = ?;
	`, uuidUser, idConversation)
	return err
}

func (db *appdbimpl) IsMember(uuidUser string, idConversation int64) (bool, error) {
	var count int
	err := db.c.QueryRow(`
		SELECT COUNT(*)
		FROM member
		WHERE uuidUser = ? AND idConversation = ?;
	`, uuidUser, idConversation).Scan(&count)
	return count > 0, err
}

func (db *appdbimpl) GetMembersByConversation(idConversation int64) ([]string, error) {
	rows, err := db.c.Query(`
		SELECT uuidUser
		FROM member
		WHERE idConversation = ?;
	`, idConversation)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []string
	for rows.Next() {
		var uuid string
		if err := rows.Scan(&uuid); err != nil {
			return nil, err
		}
		members = append(members, uuid)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return members, nil
}

func (db *appdbimpl) GetJoinedAt(uuidUser string, idConversation int64) (string, error) {
	var timestamp string
	err := db.c.QueryRow(`
		SELECT timestampJoined
		FROM member
		WHERE uuidUser = ? AND idConversation = ?;
	`, uuidUser, idConversation).Scan(&timestamp)
	return timestamp, err
}

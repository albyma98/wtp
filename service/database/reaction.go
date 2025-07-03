package database

type Reaction struct {
	UUIDUser  string
	IDMessage int64
	Emoji     string
}

type ReactionWithUser struct {
	UUIDUser string `json:"uuidUser"`
	Username string `json:"username"`
	Emoji    string `json:"emoji"`
}

// 5. AddReaction
func (db *appdbimpl) AddReaction(messageID int64, uuid string, emoji string) error {
	_, err := db.c.Exec(`INSERT INTO reaction (uuidUser, idMessage, emoji) VALUES (?, ?, ?)`, uuid, messageID, emoji)
	return err
}

// 6. RemoveReaction
func (db *appdbimpl) RemoveReaction(messageID int64, uuid string) error {
	_, err := db.c.Exec(`DELETE FROM reaction WHERE idMessage = ? AND uuidUser = ?`, messageID, uuid)
	return err
}

// 7. GetReactionsByMessageID
func (db *appdbimpl) GetReactionsByMessageID(messageID int64) ([]Reaction, error) {
	rows, err := db.c.Query(`SELECT uuidUser, idMessage, emoji FROM reaction WHERE idMessage = ?`, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reactions []Reaction
	for rows.Next() {
		var r Reaction
		err := rows.Scan(&r.UUIDUser, &r.IDMessage, &r.Emoji)
		if err != nil {
			return nil, err
		}
		reactions = append(reactions, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reactions, nil
}

func (db *appdbimpl) GetReactionsWithUserByMessageID(messageID int64) ([]ReactionWithUser, error) {
	rows, err := db.c.Query(`
                SELECT r.uuidUser, u.username, r.emoji
                FROM reaction r
                JOIN user u ON r.uuidUser = u.uuid
                WHERE r.idMessage = ?`, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reactions []ReactionWithUser
	for rows.Next() {
		var r ReactionWithUser
		if err := rows.Scan(&r.UUIDUser, &r.Username, &r.Emoji); err != nil {
			return nil, err
		}
		reactions = append(reactions, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reactions, nil
}

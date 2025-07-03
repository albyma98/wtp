package database

type MessageStatus struct {
	UUIDUser  string
	IDMessage int64
	Delivered bool
	Seen      bool
}

func (db *appdbimpl) SetDelivered(uuidUser string, idMessage int64) error {
	_, err := db.c.Exec(`
		INSERT INTO messageStatus (uuidUser, idMessage, delivered, seen)
		VALUES (?, ?, TRUE, FALSE)
		ON CONFLICT(uuidUser, idMessage)
		DO UPDATE SET delivered = TRUE;
	`, uuidUser, idMessage)
	return err
}

func (db *appdbimpl) SetSeen(uuidUser string, idMessage int64) error {
	_, err := db.c.Exec(`
		INSERT INTO messageStatus (uuidUser, idMessage, delivered, seen)
		VALUES (?, ?, TRUE, TRUE)
		ON CONFLICT(uuidUser, idMessage)
		DO UPDATE SET seen = TRUE;
	`, uuidUser, idMessage)
	return err
}

func (db *appdbimpl) GetMessageStatus(uuidUser string, idMessage int64) (MessageStatus, error) {
	var status MessageStatus
	err := db.c.QueryRow(`
		SELECT uuidUser, idMessage, delivered, seen
		FROM messageStatus
		WHERE uuidUser = ? AND idMessage = ?;
	`, uuidUser, idMessage).Scan(
		&status.UUIDUser,
		&status.IDMessage,
		&status.Delivered,
		&status.Seen,
	)
	if err != nil {
		return MessageStatus{}, err
	}
	return status, nil
}

func (db *appdbimpl) GetAllStatusesByMessage(idMessage int64) ([]MessageStatus, error) {
	rows, err := db.c.Query(`
		SELECT uuidUser, delivered, seen
		FROM messageStatus
		WHERE idMessage = ?;
	`, idMessage)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []MessageStatus
	for rows.Next() {
		var status MessageStatus
		status.IDMessage = idMessage
		if err := rows.Scan(&status.UUIDUser, &status.Delivered, &status.Seen); err != nil {
			return nil, err
		}
		statuses = append(statuses, status)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return statuses, nil
}

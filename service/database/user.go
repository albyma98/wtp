package database

type User struct {
	UUID     string  `json:"uuid"`
	Username string  `json:"username"`
	PhotoUrl *string `json:"photoUrl"`
}

func (db *appdbimpl) CreateUser(uuid string, username string, photoUrl string) error {
	_, err := db.c.Exec(`
		INSERT INTO user (uuid, username, photoUrl)
		VALUES (?, ?, ?)`,
		uuid, username, photoUrl,
	)
	return err
}

func (db *appdbimpl) GetUserByUUID(uuid string) (User, error) {
	var user User
	err := db.c.QueryRow(`
		SELECT uuid, username, photoUrl
		FROM user
		WHERE uuid = ?`,
		uuid,
	).Scan(&user.UUID, &user.Username, &user.PhotoUrl)

	return user, err
}

func (db *appdbimpl) SetUserName(uuid string, newUsername string) error {
	_, err := db.c.Exec(`
		UPDATE user
		SET username = ?
		WHERE uuid = ?`,
		newUsername, uuid,
	)
	return err
}

func (db *appdbimpl) SetPhotoUrl(uuid string, newPhotoUrl string) error {
	_, err := db.c.Exec("UPDATE user SET photoUrl = ? WHERE uuid = ?", newPhotoUrl, uuid)
	return err
}

func (db *appdbimpl) SearchUsersByPrefix(prefix string) ([]User, error) {
	rows, err := db.c.Query(`
		SELECT uuid, username, photoUrl
		FROM user
		WHERE username LIKE ?`,
		prefix+"%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.UUID, &user.Username, &user.PhotoUrl)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (db *appdbimpl) GetAllUsers() ([]User, error) {
	rows, err := db.c.Query("SELECT * FROM user")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.UUID, &user.Username, &user.PhotoUrl)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (db *appdbimpl) UserExists(uuid string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM user WHERE uuid = ?)`
	err := db.c.QueryRow(query, uuid).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (db *appdbimpl) GetPeerData(convID int64, uuidMe string) (User, error) {
	var peerUUID string
	err := db.c.QueryRow(`
		SELECT uuidUser
		FROM member
		WHERE idConversation = ? AND uuidUser != ?
		LIMIT 1;
	`, convID, uuidMe).Scan(&peerUUID)
	if err != nil {
		return User{}, err
	}

	peer, err := db.GetUserByUUID(peerUUID)
	if err != nil {
		return User{}, err
	}

	return peer, nil

}

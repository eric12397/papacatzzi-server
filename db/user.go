package db

import "github.com/papacatzzi-server/models"

func (s Store) GetUserByName(username string) (user models.User, err error) {

	err = s.db.QueryRow(`
		SELECT username
		FROM users
		WHERE username = $1
	`, username).Scan(&user.Username)

	return
}

func (s Store) GetUserByEmail(email string) (user models.User, err error) {

	err = s.db.QueryRow(`
		SELECT email, is_active
		FROM users
		WHERE email = $1
	`, email).Scan(&user.Email, &user.IsActive)

	return
}

func (s Store) InsertUser(user models.User) (err error) {

	_, err = s.db.Exec(`
		INSERT INTO users 
		(username, email, password, created_at, is_active)
		VALUES ($1, $2, $3, $4, $5)
	`, user.Username, user.Email, user.Password, user.CreatedAt, user.IsActive)

	return
}

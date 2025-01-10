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
		SELECT email
		FROM users
		WHERE email = $1
	`, email).Scan(&user.Email)

	return
}

func (s Store) InsertUser(user models.User) (err error) {

	_, err = s.db.Exec(`
		INSERT INTO users 
		(username, email, password, created_at)
		VALUES ($1, $2, $3, $4)
	`, user.Username, user.Email, user.Password, user.CreatedAt)

	return
}

func (s Store) ActivateUser(user models.User) (err error) {

	_, err = s.db.Exec(`
		UPDATE users
		SET username=$1, password=$2, is_active=$3
		WHERE email=$4
	`, user.Username, user.Password, user.IsActive, user.Email)

	return
}

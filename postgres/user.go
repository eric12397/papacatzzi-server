package postgres

import (
	"database/sql"

	"github.com/papacatzzi-server/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return UserRepository{db: db}
}

func (r UserRepository) GetUserByName(username string) (user domain.User, err error) {

	err = r.db.QueryRow(`
		SELECT username
		FROM users
		WHERE username = $1
	`, username).Scan(&user.Username)

	return
}

func (r UserRepository) GetUserByEmail(email string) (user domain.User, err error) {

	err = r.db.QueryRow(`
		SELECT id, email, password, is_active
		FROM users
		WHERE email = $1
	`, email).Scan(&user.ID, &user.Email, &user.Password, &user.IsActive)

	return
}

func (r UserRepository) GetUserByOAuthID(id string) (user domain.User, err error) {

	err = r.db.QueryRow(`
		SELECT id, email
		FROM users
		WHERE oauth_id = $1
	`, id).Scan(&user.ID, &user.Email)

	return
}

func (r UserRepository) InsertUser(user domain.User) (err error) {

	_, err = r.db.Exec(`
		INSERT INTO users 
		(username, email, password, created_at, is_active, oauth_id)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, user.Username, user.Email, user.Password, user.CreatedAt, user.IsActive, user.OAuthID)

	return
}

package data

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// User is a struct that holds the necessary information for registering a new authentication record.
// It contains the name, email, and password of the user being registered.
type User struct {
	Id       int
	Name     string
	Email    string
	Password string
}

type UserDB struct {
	DB *sql.DB
}

func (r *UserDB) Get(id int) (*User, error) {
	stmt := `SELECT username, email FROM users WHERE user_id=$1`
	user := &User{}

	err := r.DB.QueryRow(stmt, id).Scan(
		&user.Name,
		&user.Email,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return user, nil
}

func (r *UserDB) GetAllUserNames() ([]*User, error) {
	stmt := `SELECT username FROM users`
	users := []*User{}

	row, err := r.DB.Query(stmt)
	for row.Next() {
		var u User
		err := row.Scan(&u.Name)
		if err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return users, nil
}

// Register adds a new authentication record to the database. If a record with the
// same name or email already exists, it returns ErrDuplicateName or ErrDuplicateEmail.
func (r *UserDB) Register(u User) (int, error) {
	// Create a bcrypt hash of the plain-text password.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), 12)
	if err != nil {
		return 0, err
	}
	query := `
	INSERT INTO users (username, email, password_hash)
	VALUES ($1, $2, $3)
	RETURNING user_id`
	args := []interface{}{
		u.Name,
		u.Email,
		string(hashedPassword),
	}
	var id int
	err = r.DB.QueryRow(query, args...).Scan(&id)
	var pqSQLError *pq.Error
	if err != nil {
		if errors.As(err, &pqSQLError) {
			if pqSQLError.Code == "23505" && pqSQLError.Constraint == "authentification_name_key" {
				return 0, ErrDuplicateName
			} else if pqSQLError.Code == "23505" && pqSQLError.Constraint == "authentification_email_key" {
				return 0, ErrDuplicateEmail
			}
		}
		return 0, err
	}

	return id, nil
}

func (r *UserDB) Athentificate(username, password string) (int, error) {
	var id int
	var hashedPassword []byte
	query := `SELECT user_id, password_hash from users where username=$1`
	err := r.DB.QueryRow(query, username).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		}
		return 0, err
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		}
		return 0, err
	}
	return id, nil
}

func (u *UserDB) Exists(id int) (bool, error) {
	var exists bool

	stmt := `SELECT EXISTS(SELECT 1 FROM users WHERE user_id=$1)`

	err := u.DB.QueryRow(stmt, id).Scan(&exists)

	return exists, err
}

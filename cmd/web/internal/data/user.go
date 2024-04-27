package data

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// UserAuth is a struct that holds the necessary information for registering a new authentication record.
// It contains the name, email, and password of the user being registered.
type UserAuth struct {
	Name     string
	Email    string
	Password string
}

type UserData struct {
	DB *sql.DB
}

// Register adds a new authentication record to the database. If a record with the
// same name or email already exists, it returns ErrDuplicateName or ErrDuplicateEmail.
func (r *UserData) Register(u UserAuth) (int, error) {
	// Create a bcrypt hash of the plain-text password.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), 12)
	if err != nil {
		return 0, err
	}
	query := `
	INSERT INTO authentification (name, email, password)
	VALUES ($1, $2, $3)
	RETURNING id`
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

func (r *UserData) Athentificate(email,password string) (int, error) {
	var id int
	var hashedPassword []byte
	query := `SELECT id, password from authentification where email=$1`
	err := r.DB.QueryRow(query, email).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		}
		return 0, err
	}

	err= bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials	
		}	
		return 0, err
	}
	return id, nil
}

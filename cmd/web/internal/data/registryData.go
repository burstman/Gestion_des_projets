package data

import (
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

// RegistryWorker is a struct that represents a worker record in the registry.
// It contains information about the worker, such as their name, nationality,
// ID number, place of residence, workplace, blood type, sponsor information,
// and an image reference.
type RegistryWorker struct {
	ID               uint
	CreatedAt        time.Time
	Name             string
	Nationality      string
	IDnumber         string
	PlaceOfResidence string
	Workplace        string
	BloodType        string
	NameOfSponsor    string
	Sponsor          bool
	ImageReference   string
	Version          uint
}



// RegistryData is a struct that holds a reference to an SQL database connection.
// It is used to manage data related to the registry, such as authentication records
// and worker records.
type RegistryData struct {
	DB *sql.DB
}



// Insert adds a new worker record to the registry. If a record with the
// same ID number already exists, it returns ErrDuplicateRecord.
func (r *RegistryData) Insert(rg RegistryWorker) (int, error) {
	query := `
	INSERT INTO workers_registry (created_at, name, nationality, id_number,
		 place_of_residence, workplace, blood_type, name_of_sponsor, sponsor, image_reference, version)
	 VALUES (CURRENT_TIMESTAMP, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	 RETURNING id`
	args := []interface{}{
		rg.Name,
		rg.Nationality,
		rg.IDnumber,
		rg.PlaceOfResidence,
		rg.Workplace,
		rg.BloodType,
		rg.NameOfSponsor,
		rg.Sponsor,
		rg.ImageReference,
		rg.Version,
	}

	var id int
	err := r.DB.QueryRow(query, args...).Scan(&id)
	if err != nil {

		var pqSQLError *pq.Error
		if errors.As(err, &pqSQLError) {
			if pqSQLError.Code == "23505" && pqSQLError.Constraint == "workers_registry_id_number_key" {
				return 0, ErrDuplicateRecord
			}
		}
		return 0, err
	}
	return id, nil
}

// Get retrieves a single RegistryWorker record from the database by its ID.
// If no record is found, it returns ErrNoRecord.
func (r *RegistryData) Get(id int) (*RegistryWorker, error) {
	stmt := `SELECT * FROM workers_registry WHERE id = $1`
	rg := &RegistryWorker{}

	err := r.DB.QueryRow(stmt, id).Scan(
		&rg.ID,
		&rg.CreatedAt,
		&rg.Name,
		&rg.Nationality,
		&rg.IDnumber,
		&rg.PlaceOfResidence,
		&rg.Workplace,
		&rg.BloodType,
		&rg.NameOfSponsor,
		&rg.Sponsor,
		&rg.ImageReference,
		&rg.Version,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return rg, nil
}

// Latest retrieves all RegistryWorker records from the database, ordered by their ID.
// If there are no records, it returns an empty slice and no error.
func (r *RegistryData) Latest() ([]*RegistryWorker, error) {
	stmt := `SELECT * FROM workers_registry ORDER BY id  `

	rows, err := r.DB.Query(stmt)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	dataPersons := []*RegistryWorker{}

	for rows.Next() {
		rg := &RegistryWorker{}

		err := rows.Scan(
			&rg.ID,
			&rg.CreatedAt,
			&rg.Name,
			&rg.Nationality,
			&rg.IDnumber,
			&rg.PlaceOfResidence,
			&rg.Workplace,
			&rg.BloodType,
			&rg.NameOfSponsor,
			&rg.Sponsor,
			&rg.ImageReference,
			&rg.Version,
		)
		if err != nil {
			return nil, err
		}
		dataPersons = append(dataPersons, rg)
		if err = rows.Err(); err != nil {
			return nil, err
		}
	}

	return dataPersons, nil
}

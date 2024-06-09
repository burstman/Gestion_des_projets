package data

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

type ChatOrder struct {
	Intent      string   `json:"intent"`
	Tasks       []string `json:"tasks"`
	Users       []string `json:"users"`
	Comments    []string `json:"comments"`
	Projects    []string `json:"projects"`
	Deadline    []string `json:"deadline"`
	Description []string `json:"description"`
}
type Record struct {
	ID   int
	Data ChatOrder
}

type ChatData struct {
	DB *sql.DB
}

// retrieveUserOrder retrieves the order details for the given user ID.
// It queries the orderin_jason table and returns a ChatOrder struct
// containing the product and quantity information.
func (c *ChatData) RetrieveUserOrder(responseId int) (*ChatOrder, error) {

	var record Record
	var jsonData string

	query := "SELECT id, data FROM json_data WHERE id = $1"
	err := c.DB.QueryRow(query, responseId).Scan(&record.ID, &jsonData)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no record found with ID %d", responseId)
		}
		return nil, err
	}

	// Unmarshal the JSON data into the struct
	err = json.Unmarshal([]byte(jsonData), &record.Data)
	if err != nil {
		return nil, err
	}

	return &record.Data, nil
}

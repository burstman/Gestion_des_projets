package data

import (
	"database/sql"
	"fmt"
)

type ChatOrder struct {
	Intent    string
	Task      string
	User_name string
	Types     string
}

type ChatData struct {
	DB *sql.DB
}

// retrieveUserOrder retrieves the order details for the given user ID.
// It queries the orderin_jason table and returns a ChatOrder struct
// containing the product and quantity information.
func (c *ChatData) RetrieveUserOrder(responseId int) (*ChatOrder, error) {
    var (
        intent, task, user_name, types sql.NullString
    )

    stmt := "SELECT data->>'intent' AS intent, data->>'task' AS task, data->>'user_name' AS user_name, data->>'types' AS types FROM json_data WHERE id = $1"
    order := ChatOrder{}
    err := c.DB.QueryRow(stmt, responseId).Scan(&intent, &task, &user_name, &types)
    if err != nil {
        return nil, err
    }

    if intent.Valid {
        order.Intent = intent.String
    } else {
        order.Intent = "" // or set a default value if desired
    }

    if task.Valid {
        order.Task = task.String
    } else {
        order.Task = "" // or set a default value if desired
    }

    if user_name.Valid {
        order.User_name = user_name.String
    } else {
        order.User_name = "" // or set a default value if desired
    }

    if types.Valid {
        order.Types = types.String
    } else {
        order.Types = "" // or set a default value if desired
    }

    fmt.Println(order)

    return &order, nil
}

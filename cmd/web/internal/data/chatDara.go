package data

import "database/sql"

type ChatOrder struct {
	product string
	newtype string
	qty     int
}

type ChatData struct {
	DB *sql.DB
}

// retrieveUserOrder retrieves the order details for the given user ID.
// It queries the orderin_jason table and returns a ChatOrder struct
// containing the product and quantity information.
func (c *ChatData) RetrieveUserOrder(reponseId int) (*ChatOrder, error) {
	stmt := "SELECT data->>'product' AS product, data->>'type' AS type, data->>'qty' AS qty FROM orderin_json WHERE id = $1"
	order := &ChatOrder{}
	err := c.DB.QueryRow(stmt, reponseId).Scan(&order.product, &order.newtype, &order.qty)

	if err != nil {
		return nil, err
	}

	return order, nil
}

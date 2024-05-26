package data

import "database/sql"

type taskManager struct {
	DB *sql.DB
}

// func (t *taskManager) insert()  {
// 	stmt="INSERT INTO tasks (project_id, title, description, status, priority, due_date, created_by, assigned_to)
// VALUES (1, 'Design database schema', 'Create tables for the task management system', 'pending', 2, '2);"

// }

package data

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type ProjectManager struct {
	DB *sql.DB
}

type Project struct {
	ProjectID   uint
	Name        *string
	Description *string
	CreatedBy   *uint
	CreatedAt   *time.Time
	Tasks       []Task
}

func (pm *ProjectManager) InsertProject(p Project) (uint, error) {
	stmt := `INSERT INTO projects (name,description, created_by)
	 VALUES ($1, $2, $3)  RETURNING project_id`
	if p.CreatedBy != nil {
		args := []any{
			p.Name,
			p.Description,
			p.CreatedBy,
		}

		err := pm.DB.QueryRow(stmt, args...).Scan(&p.ProjectID)
		if err != nil {
			return 0, err
		}
		return p.ProjectID, nil
	}
	return 0, nil

}

type Task struct {
	TaskID      uint
	Title       *string
	Description *string
	Status      *bool
	DueDate     *time.Time
	CreatedBy   *uint
	ProjectID   *uint
	AssignedTo  *uint
	CreatedAt   *time.Time
	Comments    []Comment
	Attachments []Attachment
}

func (pm *ProjectManager) InsertTask(t Task) (uint, error) {
	stmt := `INSERT INTO projects (title, description, status, due_date, created_by, assigned_to)
	 VALUES ($1, $2, $3, $4, $5, $6)  RETURNING task_id`
	if t.CreatedBy != nil {
		args := []any{
			t.Title,
			t.Description,
			t.Status,
			t.DueDate,
			t.CreatedBy,
		}
		err := pm.DB.QueryRow(stmt, args...).Scan(&t.TaskID)
		if err != nil {
			return 0, err
		}
		return t.TaskID, nil
	}
	return 0, nil
}

type Attachment struct {
	AttachmentID uint
	TaskID       *uint
	UploadedAt   *time.Time
	UploadedBy   *uint
}

func (pm *ProjectManager) AddAttach(a Attachment) (uint, error) {
	stmt := `INSERT INTO attachments (task_id, uploaded_by)
	 VALUES ($1, $2)  RETURNING attachment_id`
	fmt.Println("Attach")
	args := []any{
		a.TaskID,
		a.UploadedBy,
	}
	err := pm.DB.QueryRow(stmt, args...).Scan(&a.AttachmentID)
	if err != nil {
		return 0, err
	}
	return a.AttachmentID, nil
}

type Comment struct {
	CommentID   uint
	TaskID      *uint
	uploadedBy  *uint
	CommentText *string
	CreatedAt   *time.Time
}

func (pm *ProjectManager) AddComment(c Comment) (uint, error) {
	stmt := `INSERT INTO attachments (task_id, uploaded_by, comment_text)
	 VALUES ($1, $2)  RETURNING attachment_id`
	if c.uploadedBy != nil {
		args := []any{
			c.TaskID,
			c.uploadedBy,
			c.CommentText,
		}
		err := pm.DB.QueryRow(stmt, args...).Scan(&c.CommentID)
		if err != nil {
			return 0, err
		}
		return c.CommentID, nil
	}
	return 0, nil
}

func (pm *ProjectManager) GetAllProjects() ([]*Project, error) {
	query := `
		SELECT
			p.project_id, p.name, p.description, p.created_at, p.created_by,
			t.task_id, t.title, t.description, t.status, t.due_date, t.created_at, t.created_by, t.assigned_to,
			c.comment_id, c.user_id, c.comment_text, c.created_at,
			a.attachment_id, a.uploaded_at, a.uploaded_by
		FROM
			projects p
		LEFT JOIN
			tasks t ON p.project_id = t.project_id
		LEFT JOIN
			comments c ON t.task_id = c.task_id
		LEFT JOIN
			attachments a ON t.task_id = a.task_id
	`

	rows, err := pm.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projectMap := make(map[int]*Project)
	taskMap := make(map[int]*Task)

	for rows.Next() {
		var (
			pID, tID, cID, aID                                        sql.NullInt64
			pName, pDescription, tTitle, tDescription                 sql.NullString
			tStatus                                                   sql.NullBool
			cText                                                     sql.NullString
			pCreatedAt, tCreatedAt, cCreatedAt, aUploadedAt           sql.NullTime
			pCreatedBy, tCreatedBy, tAssignedTo, cUserID, aUploadedBy sql.NullInt64
			tDueDate                                                  sql.NullTime
		)

		err := rows.Scan(
			&pID, &pName, &pDescription, &pCreatedAt, &pCreatedBy,
			&tID, &tTitle, &tDescription, &tStatus, &tDueDate, &tCreatedAt, &tCreatedBy, &tAssignedTo,
			&cID, &cUserID, &cText, &cCreatedAt,
			&aID, &aUploadedAt, &aUploadedBy,
		)
		if err != nil {
			return nil, err
		}

		// Populate the project if it doesn't already exist
		if pID.Valid && projectMap[int(pID.Int64)] == nil {
			projectMap[int(pID.Int64)] = &Project{
				ProjectID:   uint(pID.Int64),
				Name:        StringPointer(pName),
				Description: StringPointer(pDescription),
				CreatedAt:   TimePointer(pCreatedAt),
				CreatedBy:   UintPointer(pCreatedBy),
				Tasks:       []Task{},
			}
		}

		// Populate the task if it doesn't already exist
		if tID.Valid && taskMap[int(tID.Int64)] == nil {
			task := Task{
				TaskID:      uint(tID.Int64),
				ProjectID:   UintPointer(pID),
				Title:       StringPointer(tTitle),
				Description: StringPointer(tDescription),
				Status:      BoolPointer(tStatus),
				DueDate:     TimePointer(tDueDate),
				CreatedAt:   TimePointer(tCreatedAt),
				CreatedBy:   UintPointer(tCreatedBy),
				AssignedTo:  UintPointer(tAssignedTo),
				Comments:    []Comment{},
				Attachments: []Attachment{},
			}
			taskMap[int(tID.Int64)] = &task
			projectMap[int(pID.Int64)].Tasks = append(projectMap[int(pID.Int64)].Tasks, task)
		}

		// Add comments and attachments to the task
		if tID.Valid {
			task := taskMap[int(tID.Int64)]
			if cID.Valid {
				comment := Comment{
					CommentID:   uint(cID.Int64),
					uploadedBy:  UintPointer(cUserID),
					CommentText: StringPointer(cText),
					CreatedAt:   TimePointer(cCreatedAt),
				}
				task.Comments = append(task.Comments, comment)
			}
			if aID.Valid {
				attachment := Attachment{
					AttachmentID: uint(aID.Int64),
					UploadedAt:   TimePointer(aUploadedAt),
					UploadedBy:   UintPointer(aUploadedBy),
				}
				task.Attachments = append(task.Attachments, attachment)
			}
		}
	}

	// Check for errors after iterating over rows
	if err = rows.Err(); err != nil {
		return nil, err
	}

	var projects []*Project
	for _, project := range projectMap {
		projects = append(projects, project)
	}

	return projects, nil
}

// StringPointer converts sql.NullString to *string
func StringPointer(s sql.NullString) *string {
	if s.Valid {
		return &s.String
	}
	return nil
}

// IntPointer converts sql.NullInt64 to *int
func UintPointer(n sql.NullInt64) *uint {
	if n.Valid {
		val := uint(n.Int64)
		return &val
	}
	return nil
}

// TimePointer converts sql.NullTime to *time.Time
func TimePointer(t sql.NullTime) *time.Time {
	if t.Valid {
		return &t.Time
	}
	return nil
}

// BoolPointer converts sql.NullBool to *bool
func BoolPointer(b sql.NullBool) *bool {
	if b.Valid {
		return &b.Bool
	}
	return nil
}

func (pm *ProjectManager) GetIDFromUserName(name string) (uint, error) {
	query := `SELECT user_id, LOWER(username) FROM users`
	rows, err := pm.DB.Query(query)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var id uint
	var username string
	for rows.Next() {
		err = rows.Scan(&id, &username)
		if err != nil {
			return 0, err
		}

		if strings.EqualFold(username, strings.ToLower(name)) {
			return id, nil
		}
	}

	if err = rows.Err(); err != nil {
		return 0, err
	}

	return 0, errors.New("user not found")
}

package data

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

type ProjectManager struct {
	DB *sql.DB
}

type Project struct {
	ProjectID   int64
	Name        *string
	Description *string
	CreatedBy   *int64
	CreatedAt   *time.Time
	Tasks       []Task
}

func (pm *ProjectManager) InsertProject(p Project) (int64, error) {
	stmt := `INSERT INTO projects (name,description, created_by)
	 VALUES ($1, $2, $3)  RETURNING project_id`
	if p.CreatedBy != nil {
		args := []any{
			p.Name,
			p.Description,
			*p.CreatedBy,
		}

		err := pm.DB.QueryRow(stmt, args...).Scan(&p.ProjectID)
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				// Unique violation error code
				if pqErr.Code == "23505" {
					return 0, ErrDuplicateRecord
				}
				return 0, err
			}
			return 0, err
		}
		return p.ProjectID, nil
	}
	return 0, nil
}

type Task struct {
	TaskID      int64
	Title       *string
	Description *string
	Status      *bool
	DueDate     *time.Time
	CreatedBy   *User
	ProjectID   *int64
	AssignedTo  []*User
	CreatedAt   *time.Time
	Comments    []Comment
}

func (pm *ProjectManager) InsertTask(t Task) (int64, error) {
	fmt.Println("title", *t.Title)
	stmt := `INSERT INTO tasks (title, project_id,created_by)
	 VALUES ($1, $2, $3)  RETURNING task_id`
	if t.CreatedBy != nil {
		fmt.Println("created by Id", t.CreatedBy.Id)
		args := []any{
			*t.Title,
			*t.ProjectID,
			t.CreatedBy.Id,
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
	AttachmentID int64
	TaskID       *int64
	UploadedAt   *time.Time
	UploadedBy   *int64
}

func (pm *ProjectManager) AddAttach(a Attachment) (int64, error) {
	stmt := `INSERT INTO attachments (task_id, uploaded_by)
	 VALUES ($1, $2)  RETURNING attachment_id`
	args := []any{
		*a.TaskID,
		*a.UploadedBy,
	}
	err := pm.DB.QueryRow(stmt, args...).Scan(&a.AttachmentID)
	if err != nil {
		return 0, err
	}
	return a.AttachmentID, nil
}

type Comment struct {
	CommentID   int64
	TaskID      *int64
	User        User
	UploadedBy  *int64
	CommentText *string
	CreatedAt   *time.Time
}

func (pm *ProjectManager) AddComment(c Comment) (int64, error) {
	stmt := `INSERT INTO comments (task_id, user_id, comment_text, created_at) 
	VALUES ($1, $2, $3, CURRENT_TIMESTAMP) 
	RETURNING comment_id`
	fmt.Println("TEXTComment", *c.CommentText)

	args := []any{
		*c.TaskID,
		c.User.Id,
		*c.CommentText,
	}
	err := pm.DB.QueryRow(stmt, args...).Scan(&c.CommentID)
	if err != nil {
		return 0, err
	}
	return c.CommentID, nil

}

// GetAllProjects retrieves all projects, tasks, and related data from the database.
// It returns a slice of pointers to Project structs, which contain the project details
// and a slice of Task structs for each project.
func (pm *ProjectManager) GetAllProjects() ([]Project, error) {
	query := `
		SELECT
			p.project_id, p.name, p.description, p.created_at, p.created_by, pc.username, pc.email,
			t.task_id, t.title, t.description, t.status, t.due_date, t.created_at, t.created_by, tc.username, tc.email,
			tua.user_id AS assigned_user_id, tua.username AS assigned_username, tua.email AS assigned_email,
			c.comment_id, c.user_id, cu.username, cu.email, c.comment_text, c.created_at
		FROM
			projects p
		LEFT JOIN
			users pc ON p.created_by = pc.user_id
		LEFT JOIN
			tasks t ON p.project_id = t.project_id
		LEFT JOIN
			users tc ON t.created_by = tc.user_id
		LEFT JOIN
			attachments a ON t.task_id = a.task_id
		LEFT JOIN
			users tua ON a.uploaded_by = tua.user_id
		LEFT JOIN
			comments c ON t.task_id = c.task_id
		LEFT JOIN
			users cu ON c.user_id = cu.user_id
	`

	rows, err := pm.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project

	for rows.Next() {
		var (
			pID, tID, cID, assignedUserID                        sql.NullInt64
			pName, pDescription, tTitle, tDescription            sql.NullString
			tStatus                                              sql.NullBool
			cText                                                sql.NullString
			pCreatedAt, tCreatedAt, cCreatedAt                   sql.NullTime
			pCreatedBy, tCreatedBy, cUserID                      sql.NullInt64
			pcUsername, tcUsername, assignedUsername, cuUsername sql.NullString
			pcEmail, tcEmail, assignedEmail, cuEmail             sql.NullString
			tDueDate                                             sql.NullTime
		)

		err := rows.Scan(
			&pID, &pName, &pDescription, &pCreatedAt, &pCreatedBy, &pcUsername, &pcEmail,
			&tID, &tTitle, &tDescription, &tStatus, &tDueDate, &tCreatedAt, &tCreatedBy, &tcUsername, &tcEmail,
			&assignedUserID, &assignedUsername, &assignedEmail,
			&cID, &cUserID, &cuUsername, &cuEmail, &cText, &cCreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Find or create the project
		var project *Project
		for i := range projects {
			if projects[i].ProjectID == pID.Int64 {
				project = &projects[i]
				break
			}
		}
		if project == nil {
			project = &Project{
				ProjectID:   pID.Int64,
				Name:        StringPointer(pName),
				Description: StringPointer(pDescription),
				CreatedAt:   TimePointer(pCreatedAt),
				CreatedBy:   IntPointer(pCreatedBy),
				Tasks:       []Task{},
			}
			projects = append(projects, *project)
			project = &projects[len(projects)-1]
		}

		// Find or create the task
		var task *Task
		for i := range project.Tasks {
			if project.Tasks[i].TaskID == tID.Int64 {
				task = &project.Tasks[i]
				break
			}
		}
		if task == nil {
			task = &Task{
				TaskID:      tID.Int64,
				ProjectID:   &pID.Int64,
				Title:       StringPointer(tTitle),
				Description: StringPointer(tDescription),
				Status:      BoolPointer(tStatus),
				DueDate:     TimePointer(tDueDate),
				CreatedAt:   TimePointer(tCreatedAt),
				CreatedBy: &User{
					Id:    int(tCreatedBy.Int64),
					Name:  tcUsername.String,
					Email: tcEmail.String,
				},
				AssignedTo: []*User{},
				Comments:   []Comment{},
			}
			project.Tasks = append(project.Tasks, *task)
			task = &project.Tasks[len(project.Tasks)-1]
		}

		// Collect assignees for the task
		if assignedUserID.Valid {
			assignedUser := &User{
				Id:    int(assignedUserID.Int64),
				Name:  assignedUsername.String,
				Email: assignedEmail.String,
			}
			task.AssignedTo = append(task.AssignedTo, assignedUser)
		}

		// Add comments to the task
		if cID.Valid {
			comment := Comment{
				CommentID: cID.Int64,
				User: User{
					Id:    int(cUserID.Int64),
					Name:  cuUsername.String,
					Email: cuEmail.String,
				},
				CommentText: StringPointer(cText),
				CreatedAt:   TimePointer(cCreatedAt),
			}

			task.Comments = append(task.Comments, comment)

			// Debug print to verify comment text
			//fmt.Printf("TaskID: %d, Comment added: %+v\n", task.TaskID, comment)
		}
	}
	// Final debug print
	//fmt.Printf("Projects after processing: %+v\n", projects)
	return projects, nil
}

// StringPointer converts sql.NullString to *string
func StringPointer(s sql.NullString) *string {
	if s.Valid {
		return &s.String
	}
	return nil
}

// IntPointer converts sql.NullInt64 to *int64
func IntPointer(n sql.NullInt64) *int64 {
	if n.Valid {
		return &n.Int64
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

func (pm *ProjectManager) GetIDFromUserName(name string) (int64, error) {
	query := `SELECT user_id, LOWER(username) FROM users`
	rows, err := pm.DB.Query(query)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var id int64
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

func (pm *ProjectManager) UpdateprojectDescription(idProject int64, text string) error {
	fmt.Printf("Updating project description: %d  %s\n", idProject, text)
	query := "UPDATE projects SET description = $1 WHERE project_id = $2"
	result, err := pm.DB.Exec(query, text, idProject)
	if err != nil {
		return fmt.Errorf("failed to update project description: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Println("row affacted:", rowsAffected)
	}
	if err != nil {
		return fmt.Errorf("could not retrieve affected rows count: %v", err)
	}

	return nil
}

func (pm *ProjectManager) UpdateTaskDescription(idTask, idProject int64, text string) error {
	query := "UPDATE tasks SET description = $1 WHERE task_id = $2 AND project_id = $3"
	_, err := pm.DB.Exec(query, text, idTask, idProject)
	if err != nil {
		return fmt.Errorf("failed to update project description: %w", err)
	}
	return nil
}

func (pm *ProjectManager) UpdateprojectDeadline(idProject int64, date string) error {
	query := "UPDATE projects SET deadline = $1 WHERE project_id = $2;"
	_, err := pm.DB.Exec(query, date, idProject)
	if err != nil {
		return fmt.Errorf("failed to update project description: %w", err)
	}
	return nil
}

func (pm *ProjectManager) UpdateTaskDeadline(idTask, idProject int64, date string) error {
	query := "UPDATE tasks SET due_date = $1 WHERE task_id = $2 AND project_id = $3"
	_, err := pm.DB.Exec(query, date, idTask, idProject)
	if err != nil {
		return fmt.Errorf("failed to update project description: %w", err)
	}
	return nil
}

func (pm *ProjectManager) ProjectExists(id uint) (bool, error) {
	var exists bool

	stmt := `SELECT EXISTS(SELECT 1 FROM projects WHERE project_id=$1)`

	err := pm.DB.QueryRow(stmt, id).Scan(&exists)

	return exists, err
}

func (pm *ProjectManager) GetIDFromProjectName(name string) (int64, error) {
	query := `SELECT project_id, LOWER(name) FROM projects`
	rows, err := pm.DB.Query(query)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var id int64
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

	return 0, nil
}

func (pm *ProjectManager) GetIDFromTaskName(name string) (int64, error) {
	query := `SELECT task_id, LOWER(title) FROM tasks`
	rows, err := pm.DB.Query(query)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var id int64
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

	return 0, nil
}

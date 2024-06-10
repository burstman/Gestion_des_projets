CREATE TABLE IF NOT EXISTS tasks (
    task_id SERIAL PRIMARY KEY,
    project_id INT REFERENCES projects(project_id),
    title VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    status VARCHAR(10) NOT NULL DEFAULT 'open', 
    priority INT, 
    due_date DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INT REFERENCES users(user_id),
    assigned_to INT REFERENCES users(user_id)
);

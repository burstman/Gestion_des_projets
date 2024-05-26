CREATE TABLE IF NOT EXISTS attachments (
    attachment_id SERIAL PRIMARY KEY,
    task_id INT REFERENCES tasks(task_id),
    file_path VARCHAR(255) NOT NULL,
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    uploaded_by INT REFERENCES users(user_id)
);

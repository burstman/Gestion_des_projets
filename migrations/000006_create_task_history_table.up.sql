CREATE TABLE IF NOT EXISTS task_history (
    history_id SERIAL PRIMARY KEY,
    task_id INT REFERENCES tasks(task_id),
    change_description TEXT NOT NULL,
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    changed_by INT REFERENCES users(user_id)
);


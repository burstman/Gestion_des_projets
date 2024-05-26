CREATE TABLE IF NOT EXISTS comments (
    comment_id SERIAL PRIMARY KEY,
    task_id INT REFERENCES tasks(task_id),
    user_id INT REFERENCES users(user_id),
    comment_text TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

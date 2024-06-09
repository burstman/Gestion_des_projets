ALTER TABLE attachments DROP CONSTRAINT IF EXISTS attachments_uploaded_by_fkey;
ALTER TABLE attachments DROP CONSTRAINT IF EXISTS attachments_task_id_fkey;
DROP TABLE IF EXISTS attachements;
-- Add log_file_path column to tasks table
ALTER TABLE tasks ADD COLUMN log_file_path VARCHAR(512) COMMENT 'Path to the full logs file for this task';

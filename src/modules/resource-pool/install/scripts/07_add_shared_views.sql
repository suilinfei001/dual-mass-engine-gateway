-- Add views for sharing tables with event_processor database
-- This allows resource_pool to access users and sessions from event_processor

-- Create view for users table
CREATE OR REPLACE VIEW users AS SELECT * FROM event_processor.users;

-- Create view for sessions table
CREATE OR REPLACE VIEW sessions AS SELECT * FROM event_processor.sessions;

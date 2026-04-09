package storage

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// OpenDB 打开数据库连接
func OpenDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 配置连接池
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("[Storage] Database connected successfully")

	return db, nil
}

// InitSchema 初始化数据库表
func InitSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS tasks (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		task_id VARCHAR(255) UNIQUE NOT NULL,
		task_name VARCHAR(100) NOT NULL,
		event_id BIGINT NOT NULL,
		check_type VARCHAR(100),
		stage VARCHAR(50) NOT NULL,
		stage_order INT NOT NULL,
		check_order INT,
		execute_order INT NOT NULL,
		resource_id BIGINT,
		request_url TEXT,
		build_id BIGINT,
		status VARCHAR(20) NOT NULL DEFAULT 'pending',
		start_time TIMESTAMP NULL,
		end_time TIMESTAMP NULL,
		error_message TEXT,
		log_file_path TEXT,
		analyzing BOOLEAN NOT NULL DEFAULT FALSE,
		testbed_uuid VARCHAR(255),
		testbed_ip VARCHAR(50),
		ssh_user VARCHAR(100),
		ssh_password VARCHAR(100),
		chart_url TEXT,
		allocation_uuid VARCHAR(255),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX idx_event_id (event_id),
		INDEX idx_status (status),
		INDEX idx_execute_order (execute_order),
		INDEX idx_analyzing (analyzing)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

	CREATE TABLE IF NOT EXISTS task_results (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		task_id BIGINT NOT NULL,
		check_type VARCHAR(100) NOT NULL,
		result VARCHAR(20) NOT NULL,
		output TEXT,
		extra JSON,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_task_id (task_id),
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	log.Printf("[Storage] Database schema initialized")

	return nil
}

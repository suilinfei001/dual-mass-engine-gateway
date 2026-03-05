package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github-hub/event-processor/internal/models"
)

type ResourceStorage interface {
	CreateResource(resource *models.ExecutableResource) error
	GetResource(id int) (*models.ExecutableResource, error)
	GetAllResources() ([]*models.ExecutableResource, error)
	GetResourcesByCreator(creatorID int) ([]*models.ExecutableResource, error)
	UpdateResource(resource *models.ExecutableResource) error
	DeleteResource(id int) error
	DeleteAllResources() error
}

type ConfigStorage interface {
	GetConfig(key string) (*models.SystemConfig, error)
	GetAllConfigs() ([]*models.SystemConfig, error)
	SetConfig(key, value string) error
}

type MySQLResourceStorage struct {
	db *sql.DB
}

func NewMySQLResourceStorage(db *sql.DB) *MySQLResourceStorage {
	return &MySQLResourceStorage{db: db}
}

func (s *MySQLResourceStorage) CreateResource(resource *models.ExecutableResource) error {
	query := `
		INSERT INTO executable_resources (
			resource_name, resource_type, allow_skip, organization, project, pipeline_id, pipeline_params,
			microservice_name, pod_name, repo_path, description,
			creator_id, creator_name, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	paramsJSON, err := resource.GetPipelineParamsJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal pipeline params: %w", err)
	}

	result, err := s.db.Exec(
		query,
		resource.ResourceName, resource.ResourceType, resource.AllowSkip,
		nullString(resource.Organization), nullString(resource.Project),
		resource.PipelineID, paramsJSON,
		nullString(resource.MicroserviceName), nullString(resource.PodName),
		resource.RepoPath, nullString(resource.Description),
		resource.CreatorID, nullString(resource.CreatorName),
		resource.CreatedAt, resource.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	resource.ID = int(id)
	return nil
}

func (s *MySQLResourceStorage) GetResource(id int) (*models.ExecutableResource, error) {
	query := `
		SELECT id, resource_name, resource_type, allow_skip, organization, project, pipeline_id, pipeline_params,
			microservice_name, pod_name, repo_path, description,
			creator_id, creator_name, created_at, updated_at
		FROM executable_resources
		WHERE id = ?
	`

	resource := &models.ExecutableResource{}
	var paramsJSON string
	var organization, project, microserviceName, podName, description sql.NullString

	err := s.db.QueryRow(query, id).Scan(
		&resource.ID, &resource.ResourceName, &resource.ResourceType, &resource.AllowSkip,
		&organization, &project, &resource.PipelineID, &paramsJSON,
		&microserviceName, &podName, &resource.RepoPath, &description,
		&resource.CreatorID, &resource.CreatorName, &resource.CreatedAt, &resource.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("resource not found")
		}
		return nil, fmt.Errorf("failed to get resource: %w", err)
	}

	if organization.Valid {
		resource.Organization = organization.String
	}
	if project.Valid {
		resource.Project = project.String
	}
	if microserviceName.Valid {
		resource.MicroserviceName = microserviceName.String
	}
	if podName.Valid {
		resource.PodName = podName.String
	}
	if description.Valid {
		resource.Description = description.String
	}

	if err := resource.SetPipelineParamsFromJSON(paramsJSON); err != nil {
		return nil, fmt.Errorf("failed to unmarshal pipeline params: %w", err)
	}

	return resource, nil
}

func (s *MySQLResourceStorage) GetAllResources() ([]*models.ExecutableResource, error) {
	query := `
		SELECT id, resource_name, resource_type, allow_skip, organization, project, pipeline_id, pipeline_params,
			microservice_name, pod_name, repo_path, description,
			creator_id, creator_name, created_at, updated_at
		FROM executable_resources
		ORDER BY created_at DESC
	`

	return s.listResourcesByQuery(query)
}

func (s *MySQLResourceStorage) GetResourcesByCreator(creatorID int) ([]*models.ExecutableResource, error) {
	query := `
		SELECT id, resource_name, resource_type, allow_skip, organization, project, pipeline_id, pipeline_params,
			microservice_name, pod_name, repo_path, description,
			creator_id, creator_name, created_at, updated_at
		FROM executable_resources
		WHERE creator_id = ?
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, creatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get resources by creator: %w", err)
	}
	defer rows.Close()

	return s.scanResources(rows)
}

func (s *MySQLResourceStorage) UpdateResource(resource *models.ExecutableResource) error {
	query := `
		UPDATE executable_resources SET
			allow_skip = ?, organization = ?, project = ?, pipeline_id = ?, pipeline_params = ?,
			microservice_name = ?, pod_name = ?, repo_path = ?, description = ?, updated_at = ?
		WHERE id = ?
	`

	paramsJSON, err := resource.GetPipelineParamsJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal pipeline params: %w", err)
	}

	_, err = s.db.Exec(query,
		resource.AllowSkip,
		nullString(resource.Organization), nullString(resource.Project), resource.PipelineID, paramsJSON,
		nullString(resource.MicroserviceName), nullString(resource.PodName),
		resource.RepoPath, nullString(resource.Description), time.Now(), resource.ID)
	if err != nil {
		return fmt.Errorf("failed to update resource: %w", err)
	}

	return nil
}

func (s *MySQLResourceStorage) DeleteResource(id int) error {
	query := `DELETE FROM executable_resources WHERE id = ?`

	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete resource: %w", err)
	}

	return nil
}

func (s *MySQLResourceStorage) DeleteAllResources() error {
	query := `DELETE FROM executable_resources`

	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete all resources: %w", err)
	}

	return nil
}

func (s *MySQLResourceStorage) listResourcesByQuery(query string) ([]*models.ExecutableResource, error) {
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query resources: %w", err)
	}
	defer rows.Close()

	return s.scanResources(rows)
}

func (s *MySQLResourceStorage) scanResources(rows *sql.Rows) ([]*models.ExecutableResource, error) {
	var resources []*models.ExecutableResource

	for rows.Next() {
		resource := &models.ExecutableResource{}
		var paramsJSON string
		var organization, project, microserviceName, podName, description sql.NullString

		err := rows.Scan(
			&resource.ID, &resource.ResourceName, &resource.ResourceType, &resource.AllowSkip,
			&organization, &project, &resource.PipelineID, &paramsJSON,
			&microserviceName, &podName, &resource.RepoPath, &description,
			&resource.CreatorID, &resource.CreatorName, &resource.CreatedAt, &resource.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan resource: %w", err)
		}

		if organization.Valid {
			resource.Organization = organization.String
		}
		if project.Valid {
			resource.Project = project.String
		}
		if microserviceName.Valid {
			resource.MicroserviceName = microserviceName.String
		}
		if podName.Valid {
			resource.PodName = podName.String
		}
		if description.Valid {
			resource.Description = description.String
		}

		if err := resource.SetPipelineParamsFromJSON(paramsJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal pipeline params: %w", err)
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

type MySQLConfigStorage struct {
	db *sql.DB
}

func NewMySQLConfigStorage(db *sql.DB) *MySQLConfigStorage {
	return &MySQLConfigStorage{db: db}
}

func (s *MySQLConfigStorage) GetConfig(key string) (*models.SystemConfig, error) {
	query := `
		SELECT id, config_key, config_value, description, created_at, updated_at
		FROM system_configs
		WHERE config_key = ?
	`

	config := &models.SystemConfig{}
	var description sql.NullString

	err := s.db.QueryRow(query, key).Scan(
		&config.ID, &config.ConfigKey, &config.ConfigValue, &description,
		&config.CreatedAt, &config.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("config not found")
		}
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	if description.Valid {
		config.Description = description.String
	}

	return config, nil
}

func (s *MySQLConfigStorage) GetAllConfigs() ([]*models.SystemConfig, error) {
	query := `
		SELECT id, config_key, config_value, description, created_at, updated_at
		FROM system_configs
		ORDER BY config_key
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query configs: %w", err)
	}
	defer rows.Close()

	var configs []*models.SystemConfig
	for rows.Next() {
		config := &models.SystemConfig{}
		var description sql.NullString

		err := rows.Scan(
			&config.ID, &config.ConfigKey, &config.ConfigValue, &description,
			&config.CreatedAt, &config.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan config: %w", err)
		}

		if description.Valid {
			config.Description = description.String
		}

		configs = append(configs, config)
	}

	if configs == nil {
		configs = make([]*models.SystemConfig, 0)
	}

	return configs, nil
}

func (s *MySQLConfigStorage) SetConfig(key, value string) error {
	query := `
		INSERT INTO system_configs (config_key, config_value, created_at, updated_at)
		VALUES (?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE config_value = ?, updated_at = NOW()
	`

	_, err := s.db.Exec(query, key, value, value)
	if err != nil {
		return fmt.Errorf("failed to set config: %w", err)
	}

	return nil
}

func (s *MySQLConfigStorage) GetConfigValue(key string) (string, error) {
	config, err := s.GetConfig(key)
	if err != nil {
		return "", err
	}
	return config.ConfigValue, nil
}

func (s *MySQLConfigStorage) GetConfigsAsMap() (map[string]string, error) {
	configs, err := s.GetAllConfigs()
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, c := range configs {
		result[c.ConfigKey] = c.ConfigValue
	}
	return result, nil
}

type MySQLUserStorage struct {
	db *sql.DB
}

func NewMySQLUserStorage(db *sql.DB) *MySQLUserStorage {
	return &MySQLUserStorage{db: db}
}

func (s *MySQLUserStorage) CreateUser(user *models.User) error {
	query := `
		INSERT INTO users (username, password, role, email, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := s.db.Exec(
		query,
		user.Username, user.Password, user.Role, user.Email,
		user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	user.ID = int(id)
	return nil
}

func (s *MySQLUserStorage) GetUser(id int) (*models.User, error) {
	query := `
		SELECT id, username, password, role, email, created_at, updated_at
		FROM users
		WHERE id = ?
	`

	user := &models.User{}
	var email sql.NullString

	err := s.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Password, &user.Role, &email,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if email.Valid {
		user.Email = email.String
	}

	return user, nil
}

func (s *MySQLUserStorage) GetUserByUsername(username string) (*models.User, error) {
	query := `
		SELECT id, username, password, role, created_at, updated_at
		FROM users
		WHERE username = ?
	`

	user := &models.User{}

	err := s.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Password, &user.Role,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (s *MySQLUserStorage) GetUserByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, username, password, role, email, created_at, updated_at
		FROM users
		WHERE email = ?
	`

	user := &models.User{}
	var userEmail sql.NullString

	err := s.db.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Password, &user.Role, &userEmail,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if userEmail.Valid {
		user.Email = userEmail.String
	}

	return user, nil
}

func (s *MySQLUserStorage) GetAllUsers() ([]*models.User, error) {
	query := `
		SELECT id, username, password, role, email, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		var email sql.NullString

		err := rows.Scan(
			&user.ID, &user.Username, &user.Password, &user.Role, &email,
			&user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		if email.Valid {
			user.Email = email.String
		}

		users = append(users, user)
	}

	return users, nil
}

func (s *MySQLUserStorage) GetUsersWithPagination(page, pageSize int) ([]*models.User, int64, error) {
	offset := (page - 1) * pageSize

	var total int64
	countQuery := "SELECT COUNT(*) FROM users"
	if err := s.db.QueryRow(countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	query := `
		SELECT id, username, password, role, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.db.Query(query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}

		err := rows.Scan(
			&user.ID, &user.Username, &user.Password, &user.Role,
			&user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}

		users = append(users, user)
	}

	return users, total, nil
}

func (s *MySQLUserStorage) GetUsersWithPaginationAndKeyword(page, pageSize int, keyword string) ([]*models.User, int64, error) {
	offset := (page - 1) * pageSize

	var countQuery string
	var query string
	var args []interface{}
	var likeKeyword string

	if keyword == "" {
		countQuery = "SELECT COUNT(*) FROM users"
		query = `
			SELECT id, username, password, role, created_at, updated_at
			FROM users
			ORDER BY created_at DESC
			LIMIT ? OFFSET ?
		`
		args = []interface{}{pageSize, offset}
	} else {
		countQuery = "SELECT COUNT(*) FROM users WHERE username LIKE ?"
		query = `
			SELECT id, username, password, role, created_at, updated_at
			FROM users
			WHERE username LIKE ?
			ORDER BY created_at DESC
			LIMIT ? OFFSET ?
		`
		likeKeyword = "%" + keyword + "%"
		args = []interface{}{likeKeyword, pageSize, offset}
	}

	var total int64
	if keyword == "" {
		if err := s.db.QueryRow(countQuery).Scan(&total); err != nil {
			return nil, 0, fmt.Errorf("failed to count users: %w", err)
		}
	} else {
		if err := s.db.QueryRow(countQuery, likeKeyword).Scan(&total); err != nil {
			return nil, 0, fmt.Errorf("failed to count users: %w", err)
		}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}

		err := rows.Scan(
			&user.ID, &user.Username, &user.Password, &user.Role,
			&user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}

		users = append(users, user)
	}

	return users, total, nil
}

func (s *MySQLUserStorage) SearchUsers(keyword string) ([]*models.User, error) {
	query := `
		SELECT id, username, password, role, email, created_at, updated_at
		FROM users
		WHERE username LIKE ? OR email LIKE ?
		ORDER BY created_at DESC
	`

	searchPattern := "%" + keyword + "%"
	rows, err := s.db.Query(query, searchPattern, searchPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}
	defer rows.Close()

	users := make([]*models.User, 0)
	for rows.Next() {
		user := &models.User{}
		var email sql.NullString

		err := rows.Scan(
			&user.ID, &user.Username, &user.Password, &user.Role, &email,
			&user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		if email.Valid {
			user.Email = email.String
		}

		users = append(users, user)
	}

	return users, nil
}

func (s *MySQLUserStorage) UpdatePassword(userID int, hashedPassword string) error {
	query := `UPDATE users SET password = ?, updated_at = ? WHERE id = ?`

	_, err := s.db.Exec(query, hashedPassword, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (s *MySQLUserStorage) DeleteUser(id int) error {
	query := `DELETE FROM users WHERE id = ?`

	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (s *MySQLUserStorage) ValidatePassword(username, password string) (map[string]interface{}, error) {
	user, err := s.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, nil
	}

	return map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"role":     string(user.Role),
	}, nil
}

func (s *MySQLUserStorage) CreateSession(userID int) (map[string]interface{}, error) {
	sessionID := generateSessionID()
	expiresAt := time.Now().Add(24 * time.Hour)

	query := `
		INSERT INTO sessions (id, user_id, expires_at, created_at)
		VALUES (?, ?, ?, ?)
	`

	_, err := s.db.Exec(query, sessionID, userID, expiresAt, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return map[string]interface{}{
		"session_id": sessionID,
		"user_id":    userID,
		"expires_at": expiresAt,
	}, nil
}

func (s *MySQLUserStorage) GetSession(sessionID string) (map[string]interface{}, error) {
	query := `
		SELECT id, user_id, expires_at, created_at
		FROM sessions
		WHERE id = ? AND expires_at > NOW()
	`

	var id string
	var userID int
	var expiresAt, createdAt time.Time

	err := s.db.QueryRow(query, sessionID).Scan(&id, &userID, &expiresAt, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return map[string]interface{}{
		"id":         id,
		"user_id":    userID,
		"expires_at": expiresAt,
		"created_at": createdAt,
	}, nil
}

func (s *MySQLUserStorage) DeleteSession(sessionID string) error {
	query := `DELETE FROM sessions WHERE id = ?`

	_, err := s.db.Exec(query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

func generateSessionID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	for i := range b {
		b[i] = charset[time.Now().Nanosecond()%len(charset)]
	}
	return string(b)
}

func (s *MySQLUserStorage) GetSessionWithUser(sessionID string) (map[string]interface{}, error) {
	query := `
		SELECT s.id, s.user_id, s.expires_at, s.created_at,
			u.id, u.username, u.role
		FROM sessions s
		JOIN users u ON s.user_id = u.id
		WHERE s.id = ? AND s.expires_at > NOW()
	`

	var sessionID2 string
	var userID int
	var expiresAt, createdAt time.Time
	var uid int
	var username, role string

	err := s.db.QueryRow(query, sessionID).Scan(
		&sessionID2, &userID, &expiresAt, &createdAt,
		&uid, &username, &role,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get session with user: %w", err)
	}

	return map[string]interface{}{
		"id":         sessionID2,
		"user_id":    userID,
		"expires_at": expiresAt,
		"created_at": createdAt,
		"user": map[string]interface{}{
			"id":       uid,
			"username": username,
			"role":     role,
			"email":    "",
		},
	}, nil
}

func (s *MySQLUserStorage) DeleteAllUsers() error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, err = tx.Exec("DELETE FROM sessions")
	if err != nil {
		return fmt.Errorf("failed to delete sessions: %w", err)
	}

	_, err = tx.Exec("DELETE FROM users WHERE role != 'admin'")
	if err != nil {
		return fmt.Errorf("failed to delete users: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteExpiredSessions deletes all expired sessions from the database
func (s *MySQLUserStorage) DeleteExpiredSessions() (int64, error) {
	result, err := s.db.Exec("DELETE FROM sessions WHERE expires_at <= NOW()")
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired sessions: %w", err)
	}
	return result.RowsAffected()
}

// DeleteAllSessions deletes all sessions from the database
func (s *MySQLUserStorage) DeleteAllSessions() (int64, error) {
	result, err := s.db.Exec("DELETE FROM sessions")
	if err != nil {
		return 0, fmt.Errorf("failed to delete all sessions: %w", err)
	}
	return result.RowsAffected()
}

func (s *MySQLConfigStorage) GetAIConfig() (*AIConfig, error) {
	configs, err := s.GetConfigsAsMap()
	if err != nil {
		return nil, err
	}

	return &AIConfig{
		IP:    configs[models.ConfigKeyAIIP],
		Model: configs[models.ConfigKeyAIModel],
		Token: configs[models.ConfigKeyAIToken],
	}, nil
}

type AIConfig struct {
	IP    string `json:"ip"`
	Model string `json:"model"`
	Token string `json:"token"`
}

func (s *MySQLConfigStorage) GetAzureConfig() (*AzureConfig, error) {
	value, err := s.GetConfigValue(models.ConfigKeyAzurePAT)
	if err != nil {
		return nil, err
	}

	return &AzureConfig{
		PAT: value,
	}, nil
}

type AzureConfig struct {
	PAT string `json:"pat"`
}

func (s *MySQLConfigStorage) GetEventReceiverIP() (string, error) {
	return s.GetConfigValue(models.ConfigKeyEventReceiverIP)
}

func (s *MySQLConfigStorage) SetEventReceiverIP(ip string) error {
	return s.SetConfig(models.ConfigKeyEventReceiverIP, ip)
}

func (s *MySQLConfigStorage) SetAIConfig(config *AIConfig) error {
	if err := s.SetConfig(models.ConfigKeyAIIP, config.IP); err != nil {
		return err
	}
	if err := s.SetConfig(models.ConfigKeyAIModel, config.Model); err != nil {
		return err
	}
	if err := s.SetConfig(models.ConfigKeyAIToken, config.Token); err != nil {
		return err
	}
	return nil
}

func (s *MySQLConfigStorage) SetAzurePAT(pat string) error {
	return s.SetConfig(models.ConfigKeyAzurePAT, pat)
}

// GetLogRetentionDays gets the log file retention period in days
// Returns default value of 7 days if not configured
func (s *MySQLConfigStorage) GetLogRetentionDays() (int, error) {
	value, err := s.GetConfigValue(models.ConfigKeyLogRetentionDays)
	if err != nil {
		// Return default value if not configured
		return 7, nil
	}

	var days int
	if _, err := fmt.Sscanf(value, "%d", &days); err != nil {
		return 7, nil // Default to 7 days if parsing fails
	}

	// Validate range: 1-30 days
	if days < 1 {
		return 1, nil
	}
	if days > 30 {
		return 30, nil
	}

	return days, nil
}

// SetLogRetentionDays sets the log file retention period in days
// Valid range: 1-30 days
func (s *MySQLConfigStorage) SetLogRetentionDays(days int) error {
	// Validate range
	if days < 1 || days > 30 {
		return fmt.Errorf("log retention days must be between 1 and 30, got: %d", days)
	}

	value := fmt.Sprintf("%d", days)
	return s.SetConfig(models.ConfigKeyLogRetentionDays, value)
}

// GetAIConcurrency gets the AI analysis concurrency setting
// Returns default value of 20 if not configured
func (s *MySQLConfigStorage) GetAIConcurrency() (int, error) {
	value, err := s.GetConfigValue(models.ConfigKeyAIConcurrency)
	if err != nil {
		// Return default value if not configured
		return 20, nil
	}

	var concurrency int
	if _, err := fmt.Sscanf(value, "%d", &concurrency); err != nil {
		return 20, nil // Default to 20 if parsing fails
	}

	// Ensure the value is within reasonable bounds
	if concurrency < 1 {
		return 1, nil
	}
	if concurrency > 50 {
		return 50, nil
	}

	return concurrency, nil
}

// SetAIConcurrency sets the AI analysis concurrency setting
// Valid range: 1-50 (default 20)
func (s *MySQLConfigStorage) SetAIConcurrency(concurrency int) error {
	// Validate range
	if concurrency < 1 || concurrency > 50 {
		return fmt.Errorf("AI concurrency must be between 1 and 50, got: %d", concurrency)
	}

	value := fmt.Sprintf("%d", concurrency)
	return s.SetConfig(models.ConfigKeyAIConcurrency, value)
}

// GetAIRequestPoolSize gets the AI request pool size setting
// Returns default value of 50 if not configured
func (s *MySQLConfigStorage) GetAIRequestPoolSize() (int, error) {
	value, err := s.GetConfigValue(models.ConfigKeyAIRequestPoolSize)
	if err != nil {
		// Return default value if not configured
		return 50, nil
	}

	var poolSize int
	if _, err := fmt.Sscanf(value, "%d", &poolSize); err != nil {
		return 50, nil // Default to 50 if parsing fails
	}

	// Ensure the value is within reasonable bounds
	if poolSize < 1 {
		return 1, nil
	}
	if poolSize > 200 {
		return 200, nil
	}

	return poolSize, nil
}

// SetAIRequestPoolSize sets the AI request pool size setting
// Valid range: 1-200 (default 50)
// Must be greater than AI concurrency per event
func (s *MySQLConfigStorage) SetAIRequestPoolSize(poolSize int) error {
	// Get current AI concurrency to validate
	concurrency, _ := s.GetAIConcurrency()
	if poolSize <= concurrency {
		return fmt.Errorf("AI request pool size must be greater than AI concurrency (%d), got: %d", concurrency, poolSize)
	}

	// Validate range
	if poolSize < 1 || poolSize > 200 {
		return fmt.Errorf("AI request pool size must be between 1 and 200, got: %d", poolSize)
	}

	value := fmt.Sprintf("%d", poolSize)
	return s.SetConfig(models.ConfigKeyAIRequestPoolSize, value)
}

func (s *MySQLConfigStorage) GetConfigAsJSON(key string) (string, error) {
	value, err := s.GetConfigValue(key)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(map[string]string{key: value})
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// nullString converts a string to sql.NullString for NULL values in database
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

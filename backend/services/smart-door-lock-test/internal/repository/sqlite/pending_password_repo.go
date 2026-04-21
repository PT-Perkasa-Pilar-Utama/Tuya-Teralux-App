package sqlite

import (
	"database/sql"
	"fmt"
	"sensio/backend/services/smart-door-lock-test/internal/domain"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// PendingPassword represents a password waiting to be synced
type PendingPassword struct {
	ID            int64
	DeviceID      string
	PasswordType  domain.PasswordType
	PasswordValue string
	ValidMinutes  int
	ExpireAt      time.Time
	Status        domain.SyncStatus
	RetryCount    int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// PendingPasswordRepository handles SQLite storage for pending passwords
type PendingPasswordRepository struct {
	db *sql.DB
}

// NewPendingPasswordRepository creates a new pending password repository
func NewPendingPasswordRepository(dbPath string) (*PendingPasswordRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	repo := &PendingPasswordRepository{db: db}

	if err := repo.initializeSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return repo, nil
}

// initializeSchema creates the required tables
func (r *PendingPasswordRepository) initializeSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS pending_passwords (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		device_id TEXT NOT NULL,
		password_type TEXT NOT NULL,
		password_value TEXT NOT NULL,
		valid_minutes INTEGER NOT NULL,
		expire_at INTEGER NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending_sync',
		retry_count INTEGER NOT NULL DEFAULT 0,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		CONSTRAINT chk_status CHECK (status IN ('active', 'pending_sync', 'failed', 'expired'))
	);

	CREATE INDEX IF NOT EXISTS idx_pending_passwords_device_id ON pending_passwords(device_id);
	CREATE INDEX IF NOT EXISTS idx_pending_passwords_status ON pending_passwords(status);
	CREATE INDEX IF NOT EXISTS idx_pending_passwords_expire_at ON pending_passwords(expire_at);
	`

	_, err := r.db.Exec(schema)
	return err
}

// Close closes the database connection
func (r *PendingPasswordRepository) Close() error {
	return r.db.Close()
}

// Save stores a pending password in the database
func (r *PendingPasswordRepository) Save(password *PendingPassword) error {
	query := `
	INSERT INTO pending_passwords (
		device_id, password_type, password_value, valid_minutes,
		expire_at, status, retry_count, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now().Unix()
	expireAtUnix := password.ExpireAt.Unix()

	result, err := r.db.Exec(query,
		password.DeviceID,
		password.PasswordType,
		password.PasswordValue,
		password.ValidMinutes,
		expireAtUnix,
		password.Status,
		password.RetryCount,
		now,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to save pending password: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	password.ID = id
	return nil
}

// UpdateStatus updates the status of a pending password
func (r *PendingPasswordRepository) UpdateStatus(id int64, status domain.SyncStatus) error {
	query := `
	UPDATE pending_passwords
	SET status = ?, updated_at = ?
	WHERE id = ?
	`

	_, err := r.db.Exec(query, status, time.Now().Unix(), id)
	return err
}

// IncrementRetryCount increments the retry count for a pending password
func (r *PendingPasswordRepository) IncrementRetryCount(id int64) error {
	query := `
	UPDATE pending_passwords
	SET retry_count = retry_count + 1, updated_at = ?
	WHERE id = ?
	`

	_, err := r.db.Exec(query, time.Now().Unix(), id)
	return err
}

// GetPendingByDeviceID returns all pending passwords for a device
func (r *PendingPasswordRepository) GetPendingByDeviceID(deviceID string) ([]*PendingPassword, error) {
	query := `
	SELECT id, device_id, password_type, password_value, valid_minutes,
	       expire_at, status, retry_count, created_at, updated_at
	FROM pending_passwords
	WHERE device_id = ? AND status = 'pending_sync'
	ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanPendingPasswords(rows)
}

// GetAllPending returns all pending passwords across all devices
func (r *PendingPasswordRepository) GetAllPending() ([]*PendingPassword, error) {
	query := `
	SELECT id, device_id, password_type, password_value, valid_minutes,
	       expire_at, status, retry_count, created_at, updated_at
	FROM pending_passwords
	WHERE status = 'pending_sync'
	ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanPendingPasswords(rows)
}

// GetExpired returns all expired pending passwords
func (r *PendingPasswordRepository) GetExpired() ([]*PendingPassword, error) {
	query := `
	SELECT id, device_id, password_type, password_value, valid_minutes,
	       expire_at, status, retry_count, created_at, updated_at
	FROM pending_passwords
	WHERE status = 'pending_sync' AND expire_at < ?
	`

	rows, err := r.db.Query(query, time.Now().Unix())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanPendingPasswords(rows)
}

// Delete removes a pending password from the database
func (r *PendingPasswordRepository) Delete(id int64) error {
	query := `DELETE FROM pending_passwords WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

// GetByID retrieves a pending password by ID
func (r *PendingPasswordRepository) GetByID(id int64) (*PendingPassword, error) {
	query := `
	SELECT id, device_id, password_type, password_value, valid_minutes,
	       expire_at, status, retry_count, created_at, updated_at
	FROM pending_passwords
	WHERE id = ?
	`

	row := r.db.QueryRow(query, id)

	var pp PendingPassword
	var expireAt, createdAt, updatedAt int64

	err := row.Scan(
		&pp.ID, &pp.DeviceID, &pp.PasswordType, &pp.PasswordValue,
		&pp.ValidMinutes, &expireAt, &pp.Status, &pp.RetryCount,
		&createdAt, &updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	pp.ExpireAt = time.Unix(expireAt, 0)
	pp.CreatedAt = time.Unix(createdAt, 0)
	pp.UpdatedAt = time.Unix(updatedAt, 0)

	return &pp, nil
}

// ToDomain converts PendingPassword to domain.Password
func (pp *PendingPassword) ToDomain() *domain.Password {
	return &domain.Password{
		Value:        pp.PasswordValue,
		Type:         pp.PasswordType,
		ExpireAt:     pp.ExpireAt,
		ValidMinutes: pp.ValidMinutes,
		SyncStatus:   pp.Status,
	}
}

// FromDomain creates PendingPassword from domain.Password
func PendingPasswordFromDomain(deviceID string, password *domain.Password) *PendingPassword {
	return &PendingPassword{
		DeviceID:      deviceID,
		PasswordType:  password.Type,
		PasswordValue: password.Value,
		ValidMinutes:  password.ValidMinutes,
		ExpireAt:      password.ExpireAt,
		Status:        domain.SyncStatusPending,
		RetryCount:    0,
	}
}

// scanPendingPasswords scans rows into PendingPassword slice
func (r *PendingPasswordRepository) scanPendingPasswords(rows *sql.Rows) ([]*PendingPassword, error) {
	var passwords []*PendingPassword

	for rows.Next() {
		var pp PendingPassword
		var expireAt, createdAt, updatedAt int64

		err := rows.Scan(
			&pp.ID, &pp.DeviceID, &pp.PasswordType, &pp.PasswordValue,
			&pp.ValidMinutes, &expireAt, &pp.Status, &pp.RetryCount,
			&createdAt, &updatedAt,
		)

		if err != nil {
			return nil, err
		}

		pp.ExpireAt = time.Unix(expireAt, 0)
		pp.CreatedAt = time.Unix(createdAt, 0)
		pp.UpdatedAt = time.Unix(updatedAt, 0)

		passwords = append(passwords, &pp)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return passwords, nil
}

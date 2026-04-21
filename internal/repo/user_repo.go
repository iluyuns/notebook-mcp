package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) EnsureInviteCode(ctx context.Context, code string, maxUses int) error {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil
	}
	if maxUses < 1 {
		maxUses = 1
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO notebook_invitation_codes (code, max_uses)
		VALUES ($1, $2)
		ON CONFLICT (code) DO NOTHING
	`, code, maxUses)
	return err
}

func (r *UserRepo) GetAuthByUsername(ctx context.Context, username string) (id int64, passwordHash string, err error) {
	username = strings.TrimSpace(username)
	err = r.db.QueryRowContext(ctx, `
		SELECT id, password_hash FROM notebook_users
		WHERE username = $1 AND status = 1
	`, username).Scan(&id, &passwordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, "", sql.ErrNoRows
	}
	return id, passwordHash, err
}

func (r *UserRepo) RegisterWithInvite(ctx context.Context, username, password, inviteCode string) (int64, error) {
	username = strings.TrimSpace(username)
	inviteCode = strings.TrimSpace(inviteCode)
	if inviteCode == "" {
		return 0, fmt.Errorf("invite code is required")
	}
	if !usernamePattern.MatchString(username) {
		return 0, fmt.Errorf("username must be 3-32 chars: letters, digits, underscore")
	}
	if len(password) < 8 {
		return 0, fmt.Errorf("password must be at least 8 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	var junk int64
	err = tx.QueryRowContext(ctx, `
		UPDATE notebook_invitation_codes
		SET used_count = used_count + 1
		WHERE code = $1
		  AND used_count < max_uses
		  AND (expires_at IS NULL OR expires_at > NOW())
		RETURNING id
	`, inviteCode).Scan(&junk)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("invalid or expired invite code")
		}
		return 0, err
	}

	var userID int64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO notebook_users (username, password_hash, status)
		VALUES ($1, $2, 1)
		RETURNING id
	`, username, string(hash)).Scan(&userID)
	if err != nil {
		if isUniqueViolation(err) {
			return 0, fmt.Errorf("username already taken")
		}
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return userID, nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint")
}

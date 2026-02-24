package repository

import (
	"context"
	"time"

	"backend/internal/model"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	query := `INSERT INTO users (id, tenant_id, email, password, name, role, created_at, updated_at)
			  VALUES (:id, :tenant_id, :email, :password, :name, :role, :created_at, :updated_at)`

	_, err := r.db.NamedExecContext(ctx, query, user)
	return err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	query := `SELECT * FROM users WHERE email = ?`
	query = r.db.Rebind(query)
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	query := `SELECT * FROM users WHERE id = ?`
	query = r.db.Rebind(query)
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByTenantID(ctx context.Context, tenantID string) ([]model.User, error) {
	var users []model.User
	query := `SELECT * FROM users WHERE tenant_id = ? ORDER BY created_at DESC`
	query = r.db.Rebind(query)
	err := r.db.SelectContext(ctx, &users, query, tenantID)
	return users, err
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	user.UpdatedAt = time.Now()
	query := `UPDATE users SET email = :email, name = :name, role = :role, updated_at = :updated_at WHERE id = :id`
	_, err := r.db.NamedExecContext(ctx, query, user)
	return err
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = ?`
	query = r.db.Rebind(query)
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

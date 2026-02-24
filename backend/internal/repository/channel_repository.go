package repository

import (
	"context"
	"strings"
	"time"

	"backend/internal/model"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ChannelRepository struct {
	db *sqlx.DB
}

func NewChannelRepository(db *sqlx.DB) *ChannelRepository {
	return &ChannelRepository{db: db}
}

func (r *ChannelRepository) Create(ctx context.Context, ch *model.Channel) error {
	// Ensure slug defaults from name when missing
	if ch.Slug == "" && ch.Name != "" {
		ch.Slug = strings.ToLower(strings.ReplaceAll(ch.Name, " ", "-"))
	}
	// If name missing but slug provided, set name from slug
	if ch.Name == "" && ch.Slug != "" {
		ch.Name = ch.Slug
	}

	query := `INSERT INTO channels (id, tenant_id, name, slug, description, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`
	query = r.db.Rebind(query)
	if ch.ID == "" {
		ch.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	ch.CreatedAt = now
	ch.UpdatedAt = now
	_, err := r.db.ExecContext(ctx, query, ch.ID, ch.TenantID, ch.Name, ch.Slug, ch.Description, ch.CreatedAt, ch.UpdatedAt)
	return err
}

func (r *ChannelRepository) GetByID(ctx context.Context, id string) (*model.Channel, error) {
	query := `SELECT id, tenant_id, name, slug, description, created_at, updated_at FROM channels WHERE id=$1`
	query = r.db.Rebind(query)
	var ch model.Channel
	if err := r.db.GetContext(ctx, &ch, query, id); err != nil {
		return nil, err
	}
	return &ch, nil
}

func (r *ChannelRepository) List(ctx context.Context) ([]model.Channel, error) {
	query := `SELECT id, tenant_id, name, slug, description, created_at, updated_at FROM channels ORDER BY created_at DESC`
	query = r.db.Rebind(query)
	var out []model.Channel
	if err := r.db.SelectContext(ctx, &out, query); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *ChannelRepository) Update(ctx context.Context, ch *model.Channel) error {
	query := `UPDATE channels SET name=$1, slug=$2, description=$3, updated_at=$4 WHERE id=$5`
	query = r.db.Rebind(query)
	ch.UpdatedAt = time.Now().UTC()
	_, err := r.db.ExecContext(ctx, query, ch.Name, ch.Slug, ch.Description, ch.UpdatedAt, ch.ID)
	return err
}

func (r *ChannelRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM channels WHERE id=$1`
	query = r.db.Rebind(query)
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

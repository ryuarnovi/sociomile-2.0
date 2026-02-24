package repository

import (
	"context"
	"encoding/json"
	"time"

	"backend/internal/model"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type EventRepository struct {
	db *sqlx.DB
}

func NewEventRepository(db *sqlx.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(ctx context.Context, event *model.Event) error {
	event.ID = uuid.New().String()
	event.CreatedAt = time.Now()

	query := `INSERT INTO events (id, tenant_id, event_type, entity_type, entity_id, data, user_id, created_at)
			  VALUES (:id, :tenant_id, :event_type, :entity_type, :entity_id, :data, :user_id, :created_at)`

	_, err := r.db.NamedExecContext(ctx, query, event)
	return err
}

func (r *EventRepository) LogEvent(ctx context.Context, tenantID, eventType, entityType, entityID, userID string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		jsonData = []byte("{}")
	}

	event := &model.Event{
		TenantID:   tenantID,
		EventType:  eventType,
		EntityType: entityType,
		EntityID:   entityID,
		Data:       string(jsonData),
		UserID:     userID,
	}

	return r.Create(ctx, event)
}

func (r *EventRepository) GetByEntityID(ctx context.Context, entityType, entityID string) ([]model.Event, error) {
	var events []model.Event
	query := `SELECT * FROM events WHERE entity_type = ? AND entity_id = ? ORDER BY created_at DESC`
	query = r.db.Rebind(query)
	err := r.db.SelectContext(ctx, &events, query, entityType, entityID)
	return events, err
}

package repository

import (
	"context"
	"time"

	"backend/internal/model"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type CustomerRepository struct {
	db *sqlx.DB
}

func NewCustomerRepository(db *sqlx.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) Create(ctx context.Context, customer *model.Customer) error {
	customer.ID = uuid.New().String()
	customer.CreatedAt = time.Now()
	customer.UpdatedAt = time.Now()

	query := `INSERT INTO customers (id, tenant_id, external_id, name, channel, created_at, updated_at)
			  VALUES (:id, :tenant_id, :external_id, :name, :channel, :created_at, :updated_at)`

	_, err := r.db.NamedExecContext(ctx, query, customer)
	return err
}

func (r *CustomerRepository) GetByExternalID(ctx context.Context, externalID, tenantID string) (*model.Customer, error) {
	var customer model.Customer
	query := `SELECT * FROM customers WHERE external_id = ? AND tenant_id = ?`
	query = r.db.Rebind(query)
	err := r.db.GetContext(ctx, &customer, query, externalID, tenantID)
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *CustomerRepository) GetOrCreate(ctx context.Context, externalID, tenantID, channel string) (*model.Customer, error) {
	customer, err := r.GetByExternalID(ctx, externalID, tenantID)
	if err == nil {
		return customer, nil
	}

	// Create new customer
	newCustomer := &model.Customer{
		TenantID:   tenantID,
		ExternalID: externalID,
		Name:       "Customer " + externalID,
		Channel:    channel,
	}

	err = r.Create(ctx, newCustomer)
	if err != nil {
		return nil, err
	}

	return newCustomer, nil
}

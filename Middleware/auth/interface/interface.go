package _interface

import "context"

// Repository - интерфейс репозитория
type Repository interface {
	// Customers
	FindOrCreateCustomer(ctx context.Context, customer *Customer) (*Customer, error)

	// Cargos
	CreateCargo(ctx context.Context, cargo *Cargo) (int64, error)

	// Recipients
	FindOrCreateRecipient(ctx context.Context, recipient *Recipient) (*Recipient, error)

	// Applications
	CreateApplication(ctx context.Context, application *Application) (int64, error)
	GetApplicationByID(ctx context.Context, id int64) (*Application, error)
	ListApplications(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*Application, error)
	UpdateApplicationStatus(ctx context.Context, id int64, status ApplicationStatus) error
}

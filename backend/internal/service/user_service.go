package service

import (
	"context"
	"errors"

	"backend/internal/model"
	"backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) List(ctx context.Context, tenantID string) ([]model.User, error) {
	return s.userRepo.GetByTenantID(ctx, tenantID)
}

func (s *UserService) Create(ctx context.Context, tenantID string, req model.CreateUserRequest) (*model.User, error) {
	// Check if user exists
	existingUser, _ := s.userRepo.GetByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &model.User{
		TenantID: tenantID,
		Email:    req.Email,
		Password: string(hashedPassword),
		Name:     req.Name,
		Role:     req.Role,
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Update(ctx context.Context, id, tenantID string, req model.UpdateUserRequest) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if user.TenantID != tenantID {
		return nil, errors.New("user not found")
	}

	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Role != "" {
		user.Role = req.Role
	}

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Delete(ctx context.Context, id, tenantID string) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return errors.New("user not found")
	}

	if user.TenantID != tenantID {
		return errors.New("user not found")
	}

	return s.userRepo.Delete(ctx, id)
}

package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"context"
)

type ChannelService struct {
	repo *repository.ChannelRepository
}

func NewChannelService(r *repository.ChannelRepository) *ChannelService {
	return &ChannelService{repo: r}
}

func (s *ChannelService) Create(ctx context.Context, ch *model.Channel) error {
	return s.repo.Create(ctx, ch)
}

func (s *ChannelService) GetByID(ctx context.Context, id string) (*model.Channel, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ChannelService) List(ctx context.Context) ([]model.Channel, error) {
	return s.repo.List(ctx)
}

func (s *ChannelService) Update(ctx context.Context, ch *model.Channel) error {
	return s.repo.Update(ctx, ch)
}

func (s *ChannelService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

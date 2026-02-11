package service

import (
	"context"
	"errors"

	"github.com/hytonhan/certwatch/internal/model"
	"github.com/hytonhan/certwatch/internal/repository"
	dto "github.com/hytonhan/certwatch/internal/service/DTO"
)

var (
	ErrInvalidInput = errors.New("invalid_input")
)

type CertService interface {
	Create(ctx context.Context, input dto.CreateCertificateInput) (*model.Certificate, error)
	Get(ctx context.Context, id string) (*model.Certificate, error)
	List(ctx context.Context) ([]model.Certificate, error)
	Delete(ctx context.Context, id string) error
}

type CertificateService struct {
	repo        repository.CertificateRepository
	certService CertService
}

func New(repo repository.CertificateRepository) *CertificateService {
	return &CertificateService{repo: repo}
}

func (cs *CertificateService) Create(ctx context.Context, input dto.CreateCertificateInput) (*model.Certificate, error) {

}

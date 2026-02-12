package service

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hytonhan/certwatch/internal/model"
	"github.com/hytonhan/certwatch/internal/repository"
	dto "github.com/hytonhan/certwatch/internal/service/DTO"
)

var (
	ErrInvalidInput = errors.New("invalid_input")
)

type CertificateService interface {
	Create(ctx context.Context, input dto.CreateCertificateInput) (*model.Certificate, error)
	Get(ctx context.Context, id string) (*model.Certificate, error)
	List(ctx context.Context) ([]model.Certificate, error)
	Delete(ctx context.Context, id string) error
}

type certificateService struct {
	repo repository.CertificateRepository
}

func New(repo repository.CertificateRepository) CertificateService {
	return &certificateService{repo: repo}
}

func (cs *certificateService) Create(ctx context.Context, input dto.CreateCertificateInput) (*model.Certificate, error) {

	valerr := validateInput(input)
	if valerr != nil {
		return nil, ErrInvalidInput
	}

	nb := input.NotBefore.UTC()
	na := input.NotAfter.UTC()
	fp := strings.ToLower(input.FingerprintSHA256)
	id := uuid.NewString()
	createdAt := time.Now().UTC()

	cert := model.Certificate{
		Id:                id,
		CommonName:        input.CommonName,
		SerialNumber:      input.SerialNumber,
		Issuer:            input.Issuer,
		NotBefore:         nb,
		NotAfter:          na,
		FingerprintSHA256: fp,
		CreatedAt:         createdAt,
	}

	err := cs.repo.Create(ctx, &cert)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return nil, repository.ErrConflict
		}
		if errors.Is(err, repository.ErrNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("Creating cert: %w", err)
	}

	return &cert, nil
}

func (cs *certificateService) Get(ctx context.Context, id string) (*model.Certificate, error) {
	if id == "" {
		return nil, ErrInvalidInput
	}
	cert, err := cs.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("Getting cert: %w", err)
	}

	return cert, nil
}

func (cs *certificateService) List(ctx context.Context) ([]model.Certificate, error) {
	certs, err := cs.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("Getting many certs: %w", err)
	}

	return certs, nil
}

func (cs *certificateService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidInput
	}
	err := cs.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return repository.ErrNotFound
		}
		return fmt.Errorf("Deleting cert: %w", err)
	}

	return nil
}

func validateInput(input dto.CreateCertificateInput) error {
	if input.CommonName == "" || input.SerialNumber == "" || input.Issuer == "" || input.FingerprintSHA256 == "" {
		return ErrInvalidInput
	}
	if len(input.CommonName) > 255 {
		return ErrInvalidInput
	}
	if len(input.SerialNumber) > 128 {
		return ErrInvalidInput
	}
	if len(input.Issuer) > 255 {
		return ErrInvalidInput
	}
	if len(input.FingerprintSHA256) != 64 {
		return ErrInvalidInput
	}
	_, hexerr := hex.DecodeString(input.FingerprintSHA256)
	if hexerr != nil {
		return ErrInvalidInput
	}
	if input.NotBefore.IsZero() || input.NotAfter.IsZero() {
		return ErrInvalidInput
	}
	if !input.NotAfter.After(input.NotBefore) {
		return ErrInvalidInput
	}

	return nil
}

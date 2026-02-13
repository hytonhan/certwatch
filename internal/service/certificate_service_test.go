package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hytonhan/certwatch/internal/model"
	"github.com/hytonhan/certwatch/internal/repository"
	dto "github.com/hytonhan/certwatch/internal/service/DTO"
)

type FakeCertRepo struct {
}

func (fcr FakeCertRepo) Create(ctx context.Context, cert *model.Certificate) error {
	return nil
}

func (fcr FakeCertRepo) GetByID(ctx context.Context, id string) (*model.Certificate, error) {
	cert := model.Certificate{
		Id: id,
	}
	if id == "id1" {
		return &cert, nil
	}
	return nil, repository.ErrNotFound
}

func (fcr FakeCertRepo) List(ctx context.Context) ([]model.Certificate, error) {
	return []model.Certificate{}, nil
}

func (fcr FakeCertRepo) Delete(ctx context.Context, id string) error {
	if id == "id1" {
		return nil
	}
	return repository.ErrNotFound
}

func (fcr FakeCertRepo) ListExpiring(ctx context.Context, before time.Time, now time.Time) ([]model.Certificate, error) {
	return []model.Certificate{}, nil
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name     string
		input    dto.CreateCertificateInput
		expected error
	}{
		{"valid input", createInput("", "", "", time.Time{}, time.Time{}, ""), nil},
		{"common name too long", createInput("looooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooong name", "", "", time.Time{}, time.Time{}, ""), ErrInvalidInput},
		{"serial too long", createInput("", "looooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooong name", "", time.Time{}, time.Time{}, ""), ErrInvalidInput},
		{"issuer too long", createInput("", "", "looooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooong name", time.Time{}, time.Time{}, ""), ErrInvalidInput},
		{"fingerprint too long", createInput("", "", "", time.Time{}, time.Time{}, "bde4918f9e08256c787948908be7f5c1ebeead20ab4f596ecfccb62325009b22a"), ErrInvalidInput},
		{"fingerprint invalid", createInput("", "", "", time.Time{}, time.Time{}, "this is a string and not valid hex code but it is right length!!"), ErrInvalidInput},
		{"not before is after not after", createInput("", "", "", time.Now().Add(time.Hour), time.Now(), ""), ErrInvalidInput},
		{"empty ca", dto.CreateCertificateInput{SerialNumber: "test", Issuer: "test", FingerprintSHA256: "test"}, ErrInvalidInput},
		{"empty serial", dto.CreateCertificateInput{CommonName: "test", Issuer: "test", FingerprintSHA256: "test"}, ErrInvalidInput},
		{"empty iss", dto.CreateCertificateInput{CommonName: "test", SerialNumber: "test", FingerprintSHA256: "test"}, ErrInvalidInput},
		{"empty fingerprint", dto.CreateCertificateInput{CommonName: "test", SerialNumber: "test", Issuer: "test"}, ErrInvalidInput},
		{"empty not before", dto.CreateCertificateInput{CommonName: "test", SerialNumber: "test", Issuer: "", FingerprintSHA256: "bde4918f9e08256c787948908be7f5c1ebeead20ab4f596ecfccb62325009b22", NotBefore: time.Time{}}, ErrInvalidInput},
		{"empty not after", dto.CreateCertificateInput{CommonName: "test", SerialNumber: "test", Issuer: "", FingerprintSHA256: "bde4918f9e08256c787948908be7f5c1ebeead20ab4f596ecfccb62325009b22", NotBefore: time.Now(), NotAfter: time.Time{}}, ErrInvalidInput},
	}
	repo := FakeCertRepo{}
	srv := New(repo)
	ctx := context.Background()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cert, err := srv.Create(ctx, test.input)
			if !errors.Is(err, test.expected) {
				t.Errorf("Create(%q) = %v; want %v", test.input, cert, test.expected)
			}
			if err != nil {
				return
			}
			if uuid.Validate(cert.Id) != nil {
				t.Errorf("invalid uuid: %v", cert.Id)
			}
			if test.input.CommonName != cert.CommonName {
				t.Errorf("Create(%q) = %v; want %v", test.input, cert.CommonName, test.input.CommonName)
			}
			if test.input.SerialNumber != cert.SerialNumber {
				t.Errorf("Create(%q) = %v; want %v", test.input, cert.SerialNumber, test.input.SerialNumber)
			}
			if test.input.Issuer != cert.Issuer {
				t.Errorf("Create(%q) = %v; want %v", test.input, cert.Issuer, test.input.Issuer)
			}
			if test.input.NotBefore.UTC() != cert.NotBefore {
				t.Errorf("Create(%q) = %v; want %v", test.input, cert.NotBefore.UTC(), test.input.NotBefore)
			}
			if test.input.NotAfter.UTC() != cert.NotAfter {
				t.Errorf("Create(%q) = %v; want %v", test.input, cert.NotAfter.UTC(), test.input.NotAfter)
			}
			if test.input.FingerprintSHA256 != cert.FingerprintSHA256 {
				t.Errorf("Create(%q) = %v; want %v", test.input, cert.FingerprintSHA256, test.input.FingerprintSHA256)
			}
			if cert.CreatedAt.IsZero() {
				t.Errorf("Create(%q) = %v; want %v", test.input, cert.CreatedAt, "\"Valid time\"")
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected error
	}{
		{"valid input", "id1", nil},
		{"empty input", "", ErrInvalidInput},
		{"ErrNotFound bubbles", "doesn't exists", repository.ErrNotFound},
	}
	repo := FakeCertRepo{}
	srv := New(repo)
	ctx := context.Background()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cert, err := srv.Get(ctx, test.input)
			if !errors.Is(err, test.expected) {
				t.Errorf("Get(%q) = %v; want %v", test.input, cert, test.expected)
			}
		})
	}
}

func TestList(t *testing.T) {
	tests := []struct {
		name     string
		expected error
	}{
		{"valid input", nil},
	}
	repo := FakeCertRepo{}
	srv := New(repo)
	ctx := context.Background()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cert, err := srv.List(ctx)
			if !errors.Is(err, test.expected) {
				t.Errorf("List(); want %v", test.expected)
			}
			if len(cert) != 0 {
				t.Errorf("List(); want %v", "len(cert) != 0")
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected error
	}{
		{"valid input", "id1", nil},
		{"empty input", "", ErrInvalidInput},
		{"ErrNotFound bubbles", "doesn't exists", repository.ErrNotFound},
	}
	repo := FakeCertRepo{}
	srv := New(repo)
	ctx := context.Background()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := srv.Delete(ctx, test.input)
			if !errors.Is(err, test.expected) {
				t.Errorf("Delete(%q); want %v", test.input, test.expected)
			}
		})
	}
}

func createInput(
	commonName string,
	serial string,
	iss string,
	nb time.Time,
	na time.Time,
	fingerprint string) dto.CreateCertificateInput {

	if commonName == "" {
		commonName = "SomeName"
	}
	if serial == "" {
		serial = "e4d530d0-00d1-49ab-b579-f58efbcdc521"
	}
	if iss == "" {
		iss = "SomeIssuer"
	}
	if nb.IsZero() {
		nb = time.Now().Add(time.Hour * -1)
	}
	if na.IsZero() {
		na = time.Now().Add(time.Hour)
	}
	if fingerprint == "" {
		fingerprint = "bde4918f9e08256c787948908be7f5c1ebeead20ab4f596ecfccb62325009b22"
	}

	return dto.CreateCertificateInput{
		CommonName:        commonName,
		SerialNumber:      serial,
		Issuer:            iss,
		NotBefore:         nb,
		NotAfter:          na,
		FingerprintSHA256: fingerprint,
	}
}

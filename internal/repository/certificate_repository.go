package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/hytonhan/certwatch/internal/model"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

var (
	ErrNotFound = errors.New("not_found")
	ErrConflict = errors.New("conflict")
)

type CertificateRepository struct {
	db *sql.DB
}

func NewCertificateRepository(db *sql.DB) *CertificateRepository {
	return &CertificateRepository{db: db}
}

func (cr *CertificateRepository) Create(ctx context.Context, cert *model.Certificate) error {

	_, err := cr.db.ExecContext(ctx, "INSERT INTO certificates (id, common_name, serial_number, issuer, not_before, not_after, fingerprint_sha256, created_at)  VALUES(?,?,?,?,?,?,?,?)",
		cert.Id,
		cert.CommonName,
		cert.SerialNumber,
		cert.Issuer,
		cert.NotBefore,
		cert.NotAfter,
		cert.FingerPrintSHA256,
		cert.CreatedAt)
	if err != nil {
		var sqlErr *sqlite.Error
		if errors.As(err, &sqlErr) {
			if sqlErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
				return ErrConflict
			}
		}
		return fmt.Errorf("Creating cert: %w", err)
	}
	return nil
}

func (cr *CertificateRepository) GetByID(ctx context.Context, id string) (*model.Certificate, error) {

	result := cr.db.QueryRowContext(
		ctx,
		"SELECT id, common_name, serial_number, issuer, not_before, not_after, fingerprint_sha256, created_at FROM certificates WHERE id = ?",
		id)

	var returnVal model.Certificate
	err := result.Scan(&returnVal.Id,
		&returnVal.CommonName,
		&returnVal.SerialNumber,
		&returnVal.Issuer,
		&returnVal.NotBefore,
		&returnVal.NotAfter,
		&returnVal.FingerPrintSHA256,
		&returnVal.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("Get cert: %w", err)
	}

	return &returnVal, nil
}

func (cr *CertificateRepository) List(ctx context.Context) ([]model.Certificate, error) {

	result, err := cr.db.QueryContext(
		ctx,
		"SELECT id, common_name, serial_number, issuer, not_before, not_after, fingerprint_sha256, created_at FROM certificates",
	)
	if err != nil {
		return nil, fmt.Errorf("Querying for certs: %w", err)
	}
	defer result.Close()

	var retValue []model.Certificate
	for result.Next() {
		item := model.Certificate{}
		err2 := result.Scan(
			&item.Id,
			&item.CommonName,
			&item.SerialNumber,
			&item.Issuer,
			&item.NotBefore,
			&item.NotAfter,
			&item.FingerPrintSHA256,
			&item.CreatedAt)
		if err2 != nil {
			return nil, fmt.Errorf("Querying for certs: %w", err2)
		}
		retValue = append(retValue, item)
	}
	if er := result.Err(); er != nil {
		return nil, fmt.Errorf("Querying for certs: %w", er)
	}
	return retValue, nil
}

func (cr *CertificateRepository) Delete(ctx context.Context, id string) error {

	result, err := cr.db.ExecContext(
		ctx,
		"DELETE FROM certificates WHERE id = ?",
		id,
	)
	if err != nil {
		return fmt.Errorf("Deleting cert: %w", err)
	}
	rows, rowerr := result.RowsAffected()
	if rowerr != nil {
		return fmt.Errorf("Deleting cert: %w", rowerr)
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

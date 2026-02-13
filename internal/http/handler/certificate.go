package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/hytonhan/certwatch/internal/model"
	"github.com/hytonhan/certwatch/internal/repository"
	"github.com/hytonhan/certwatch/internal/service"
	dto "github.com/hytonhan/certwatch/internal/service/DTO"
)

type CertificateHandler struct {
	service service.CertificateService
	logger  *slog.Logger
}

type CreateRequest struct {
	CommonName        string    `json:"common_name"`
	SerialNumber      string    `json:"serial_number"`
	Issuer            string    `json:"issuer"`
	NotBefore         time.Time `json:"not_before"`
	NotAfter          time.Time `json:"not_after"`
	FingerprintSHA256 string    `json:"fingerprintsha256"`
}

func NewCertificateHandler(s service.CertificateService, log *slog.Logger) *CertificateHandler {
	return &CertificateHandler{service: s, logger: log}
}

func (h *CertificateHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {

	h.logger.InfoContext(r.Context(), "Received Create request")
	var req CreateRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	input := dto.CreateCertificateInput{
		CommonName:        req.CommonName,
		SerialNumber:      req.SerialNumber,
		Issuer:            req.Issuer,
		NotBefore:         req.NotBefore,
		NotAfter:          req.NotAfter,
		FingerprintSHA256: req.FingerprintSHA256,
	}

	cert, err := h.service.Create(r.Context(), input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			h.logger.InfoContext(r.Context(), "Create failed: invalid input")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, repository.ErrConflict) {
			h.logger.InfoContext(r.Context(), "Create failed: conflict")
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		h.logger.WarnContext(r.Context(), "Create failed for unknown reason")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	h.logger.InfoContext(r.Context(), "Created certificate with id "+cert.Id)
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Contect-Type", "application/json")
	json.NewEncoder(w).Encode(cert.Id)
}

func (h *CertificateHandler) HandleGet(w http.ResponseWriter, r *http.Request) {

	h.logger.InfoContext(r.Context(), "Received get request")
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB

	id := r.PathValue("id")

	cert, err := h.service.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			h.logger.InfoContext(r.Context(), "Get failed: invalid input")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, repository.ErrNotFound) {
			h.logger.InfoContext(r.Context(), "Get failed: not found with id "+id)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		h.logger.InfoContext(r.Context(), "Get failed for unkown reason")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	h.logger.InfoContext(r.Context(), "Found cert with id "+id)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Contect-Type", "application/json")
	json.NewEncoder(w).Encode(cert)
}

func (h *CertificateHandler) HandleList(w http.ResponseWriter, r *http.Request) {

	h.logger.InfoContext(r.Context(), "Received List request")
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB

	var certs []model.Certificate
	var err error

	within := r.URL.Query().Get("expiring_within")
	if within == "" {
		certs, err = h.service.List(r.Context())
		if err != nil {
			h.logger.WarnContext(r.Context(), "List failed for unknown reason")
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	} else {
		d, err := time.ParseDuration(within)
		if err != nil {
			http.Error(w, "invalid input", http.StatusBadRequest)
			return
		}
		expired := r.URL.Query().Get("expired")
		var option service.ExpiryOption
		if expired == "" {
			option = service.IncludeExpired
		} else {
			parsed, berr := strconv.ParseBool(expired)
			if berr != nil {
				http.Error(w, "invalid input", http.StatusBadRequest)
				return
			}
			if parsed == true {
				option = service.IncludeExpired
			} else {
				option = service.ExcludeExpired
			}
		}
		h.logger.InfoContext(r.Context(), "Params: window: "+within+". expired: "+expired)
		certs, err = h.service.ListExpiring(r.Context(), d, option)
		if err != nil {
			h.logger.WarnContext(r.Context(), "List failed for unknown reason")
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}

	h.logger.InfoContext(r.Context(), "Fetched "+strconv.Itoa(len(certs))+" certs")

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Contect-Type", "application/json")
	json.NewEncoder(w).Encode(certs)
}

func (h *CertificateHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {

	h.logger.InfoContext(r.Context(), "Received delete request")
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB

	id := r.PathValue("id")

	err := h.service.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.logger.InfoContext(r.Context(), "Delete failed: cert not found with id "+id)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		h.logger.WarnContext(r.Context(), "Delete failed for unknown reason")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	h.logger.InfoContext(r.Context(), "Deleted cert with id "+id)
	w.WriteHeader(http.StatusNoContent)
}

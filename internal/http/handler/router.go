package handler

import (
	"net/http"
)

func (h *CertificateHandler) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("POST /certificates", h.HandleCreate)
	mux.HandleFunc("GET /certificates", h.HandleList)
	mux.HandleFunc("GET /certificates/{id}", h.HandleGet)
	mux.HandleFunc("DELETE /certificates/{id}", h.HandleDelete)

	return mux
}

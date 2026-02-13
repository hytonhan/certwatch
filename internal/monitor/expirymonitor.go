package monitor

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/hytonhan/certwatch/internal/model"
	"github.com/hytonhan/certwatch/internal/service"
)

type ExpiryMonitor struct {
	service  service.CertificateService
	interval time.Duration
	window   time.Duration
	logger   *slog.Logger
}

func NewMonitor(service service.CertificateService, interval time.Duration, window time.Duration, logger *slog.Logger) *ExpiryMonitor {
	return &ExpiryMonitor{service: service, interval: interval, window: window, logger: logger}
}

func (m *ExpiryMonitor) Start(ctx context.Context) {

	m.logger.InfoContext(ctx, "expiry monitor starter",
		"interval", m.interval,
		"window", m.window)
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	reported := map[string]model.Certificate{}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			certs, err := m.service.ListExpiring(ctx, m.window, service.ExcludeExpired)
			if err != nil {
				m.logger.WarnContext(ctx, "unknown error occured")
				continue
			}
			if len(certs) > 0 {
				m.logger.InfoContext(ctx, "Found "+strconv.Itoa(len(certs))+" expiring certs!")
				for _, cert := range certs {
					_, alreadyReported := reported[cert.Id]
					if alreadyReported {
						continue
					}
					m.logger.WarnContext(ctx,
						"Expiring.",
						"id", cert.Id,
						"common_name", cert.CommonName,
						"expires_at", cert.NotAfter)
					reported[cert.Id] = cert
				}
			}
		}
	}
}

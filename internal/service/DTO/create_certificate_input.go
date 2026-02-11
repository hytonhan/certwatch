package dto

import "time"

type CreateCertificateInput struct {
	CommonName        string
	SerialNumber      string
	Issuer            string
	NotBefore         time.Time
	NotAfter          time.Time
	FingerprintSHA256 string
}

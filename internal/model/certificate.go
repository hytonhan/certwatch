package model

import "time"

type CertificateId = string

type Certificate struct {
	Id                CertificateId
	CommonName        string
	SerialNumber      string
	Issuer            string
	NotBefore         time.Time
	NotAfter          time.Time
	FingerPrintSha256 string
	CreatedAt         time.Time
}

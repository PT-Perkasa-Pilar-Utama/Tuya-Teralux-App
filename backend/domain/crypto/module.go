package crypto

import (
	cryptoservices "sensio/domain/crypto/crypto/services"
)

type CryptoModule struct {
	PDFProtector *cryptoservices.PDFProtector
	ZIPEncryptor *cryptoservices.ZIPEncryptor
}

func NewCryptoModule() *CryptoModule {
	return &CryptoModule{
		PDFProtector: cryptoservices.NewPDFProtector(),
		ZIPEncryptor: cryptoservices.NewZIPEncryptor(),
	}
}
package autocert

import (
	"crypto/tls"
	"sync"
	"time"
)

type CertManager struct {
	lock     sync.RWMutex
	cert     *tls.Certificate
	certPath string
	keyPath  string
}

// NewManager refreshes certificates before they expire
func NewManager(certPath, keyPath string) (*CertManager, error) {
	m := &CertManager{
		certPath: certPath,
		keyPath:  keyPath,
	}
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}
	m.cert = &cert
	go m.updater()

	return m, nil
}

func (m *CertManager) tryRefresh() time.Duration {
	newCert, err := tls.LoadX509KeyPair(m.certPath, m.keyPath)
	if err != nil {
		return time.Hour
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	m.cert = &newCert
	return time.Hour * 24
}

func (m *CertManager) updater() {
	for {
		sleep := m.tryRefresh()
		time.Sleep(sleep)
	}
}

func (m *CertManager) GetCertificateFunc() func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		m.lock.RLock()
		defer m.lock.RUnlock()
		return m.cert, nil
	}
}

package session

import (
	"fmt"
	"log"

	"zeng_bot/internal/models"

	http "github.com/bogdanfinn/tls-client"
)

// TargetSession is the production Session implementation backed by
// bogdanfinn/tls-client for TLS fingerprint spoofing.
type TargetSession struct {
	client http.HttpClient
	proxy  *models.Proxy
}

// NewTargetSession creates a new session with a spoofed TLS client.
// This is a skeleton — it initializes the client but does not yet
// implement real Target API logic.
func NewTargetSession() (*TargetSession, error) {
	opts := []http.HttpClientOption{
		http.WithTimeoutSeconds(30),
		http.WithClientProfile(http.Chrome_120),
		http.WithNotFollowRedirects(),
	}

	client, err := http.NewHttpClient(http.NewNoopLogger(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create tls client: %w", err)
	}

	return &TargetSession{client: client}, nil
}

// SetProxy assigns a proxy to the underlying TLS client.
func (s *TargetSession) SetProxy(proxy models.Proxy) error {
	s.proxy = &proxy
	if err := s.client.SetProxy(proxy.URL()); err != nil {
		return fmt.Errorf("failed to set proxy: %w", err)
	}
	return nil
}

// GetCookies returns the current PerimeterX cookies. Skeleton returns empty.
func (s *TargetSession) GetCookies() map[string]string {
	log.Println("[session] GetCookies called (skeleton — returning empty)")
	return map[string]string{}
}

// Do executes an HTTP request via the TLS client. This skeleton implementation
// logs the request but does not perform actual Target API calls.
func (s *TargetSession) Do(method, url string, headers map[string]string, body []byte) (int, []byte, error) {
	log.Printf("[session] %s %s (skeleton — no-op)", method, url)
	return 200, []byte(`{"status":"skeleton"}`), nil
}

// compile-time check: TargetSession must satisfy the Session interface.
var _ Session = (*TargetSession)(nil)

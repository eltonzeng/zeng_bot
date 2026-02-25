// Package session manages TLS client instances and proxy rotation.
package session

import "zeng_bot/internal/models"

// Session represents an isolated HTTP session with its own TLS client,
// proxy, and PerimeterX cookies. Each checkout task gets its own Session.
type Session interface {
	// SetProxy assigns a proxy to this session's underlying TLS client.
	SetProxy(proxy models.Proxy) error

	// GetCookies returns the current PerimeterX cookie values.
	GetCookies() map[string]string

	// Do executes an HTTP request and returns the raw response body.
	// The underlying client handles TLS fingerprint spoofing.
	Do(method, url string, headers map[string]string, body []byte) (statusCode int, respBody []byte, err error)

	// WarmUp performs an initial request to target.com to populate the
	// cookie jar with PerimeterX cookies and extract a visitorId.
	// Must be called before Login or any checkout API calls.
	WarmUp() error

	// Login authenticates with a Target account and persists the resulting
	// auth cookies/token in the session for subsequent API calls.
	Login(email, password string) error
}

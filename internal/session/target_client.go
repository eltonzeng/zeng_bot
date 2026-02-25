package session

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	stdhttp "net/http"
	"net/url"
	"regexp"
	"strings"

	"zeng_bot/internal/models"

	fhttp "github.com/bogdanfinn/fhttp"
	http "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

const (
	targetBaseURL       = "https://www.target.com"
	targetLoginPageURL  = "https://www.target.com/login"
	gspBaseURL          = "https://gsp.target.com"
	credValidationsURL  = gspBaseURL + "/gsp/authentications/v1/credential_validations?client_id=ecom-web-1.0.0"
	skipPhoneURL        = gspBaseURL + "/gsp/authentications/v1/skip_phone_verifications"
	clientTokensURL     = gspBaseURL + "/gsp/oauth_tokens/v2/client_tokens"
	tokenValidationsURL = gspBaseURL + "/gsp/oauth_validations/v3/token_validations"
)

// pxCookieNames lists the PerimeterX cookies we need to track.
var pxCookieNames = []string{"_px3", "_pxvid", "_pxhd"}

// visitorIDPattern matches the visitorId value embedded in Target's HTML/scripts.
var visitorIDPattern = regexp.MustCompile(`"visitorId"\s*:\s*"([A-Za-z0-9-]+)"`)

// TargetSession is the production Session implementation backed by
// bogdanfinn/tls-client for TLS fingerprint spoofing.
type TargetSession struct {
	client    http.HttpClient
	proxy     *models.Proxy
	deviceID  string
	tealeafID string
	VisitorID string
}

// NewTargetSession creates a new session with a spoofed TLS client
// and an automatic cookie jar for persisting PerimeterX cookies.
func NewTargetSession() (*TargetSession, error) {
	jar := http.NewCookieJar()

	opts := []http.HttpClientOption{
		http.WithTimeoutSeconds(30),
		http.WithClientProfile(profiles.Chrome_120),
		http.WithNotFollowRedirects(),
		http.WithCookieJar(jar),
	}

	client, err := http.NewHttpClient(http.NewNoopLogger(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create tls client: %w", err)
	}

	return &TargetSession{client: client, deviceID: generateDeviceID()}, nil
}

// generateDeviceID returns a random 64-char hex string matching the format
// observed in Target's accessToken JWT `did` claim.
func generateDeviceID() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return strings.Repeat("0", 64)
	}
	return hex.EncodeToString(b)
}

// SetProxy assigns a proxy to the underlying TLS client.
func (s *TargetSession) SetProxy(proxy models.Proxy) error {
	s.proxy = &proxy
	if err := s.client.SetProxy(proxy.URL()); err != nil {
		return fmt.Errorf("failed to set proxy: %w", err)
	}
	return nil
}

// GetCookies returns the current PerimeterX cookies from the jar.
func (s *TargetSession) GetCookies() map[string]string {
	result := make(map[string]string)

	u, err := url.Parse(targetBaseURL)
	if err != nil {
		return result
	}

	cookies := s.client.GetCookies(u)
	for _, c := range cookies {
		for _, name := range pxCookieNames {
			if c.Name == name {
				result[name] = c.Value
			}
		}
	}
	return result
}

// getCookieValue returns the value of a named cookie for the given URL,
// or an empty string if not found.
func (s *TargetSession) getCookieValue(rawURL, name string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	for _, c := range s.client.GetCookies(u) {
		if c.Name == name {
			return c.Value
		}
	}
	return ""
}

// Do executes an HTTP request via the TLS client and returns the
// status code and response body.
func (s *TargetSession) Do(method, rawURL string, headers map[string]string, body []byte) (int, []byte, error) {
	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewReader(body)
	}

	req, err := fhttp.NewRequest(method, rawURL, bodyReader)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to build request: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return resp.StatusCode, respBody, nil
}

// gspHeaders returns the standard headers for requests to gsp.target.com.
// sec-fetch-site is same-site (not same-origin) because gsp is a different
// subdomain from www.
func gspHeaders() map[string]string {
	h := models.DefaultHeaders()
	h["Content-Type"] = "application/json"
	h["Accept"] = "application/json"
	h["Origin"] = "https://www.target.com"
	h["Referer"] = "https://www.target.com/cart"
	h["Sec-Fetch-Site"] = "same-site"
	return h
}

// WarmUp performs the anonymous session bootstrap:
//  1. GET target.com — populate cookie jar, extract visitorId
//  2. POST client_tokens (client_credentials grant) — establish login-session
//     and TealeafAkaSid cookies needed before credential_validations
func (s *TargetSession) WarmUp() error {
	headers := models.DefaultHeaders()
	status, body, err := s.Do("GET", targetBaseURL, headers, nil)
	if err != nil {
		return fmt.Errorf("warm-up request failed: %w", err)
	}
	log.Printf("[session] warm-up response: status=%d bodyLen=%d", status, len(body))

	if m := visitorIDPattern.FindSubmatch(body); len(m) > 1 {
		s.VisitorID = string(m[1])
		log.Printf("[session] extracted visitorId: %s", s.VisitorID)
	} else {
		log.Println("[session] warning: visitorId not found in warm-up response")
	}

	px := s.GetCookies()
	captured := make([]string, 0, len(px))
	for k := range px {
		captured = append(captured, k)
	}
	if len(captured) > 0 {
		log.Printf("[session] captured PX cookies: %s", strings.Join(captured, ", "))
	} else {
		log.Println("[session] warning: no PX cookies captured during warm-up")
	}

	// Visit the login page to collect cookies Target sets before accepting
	// credential_validations: login-session, 3YCzT93n, sapphire, etc.
	loginHeaders := models.DefaultHeaders()
	loginHeaders["Sec-Fetch-Site"] = "same-origin"
	loginHeaders["Sec-Fetch-Mode"] = "navigate"
	loginHeaders["Sec-Fetch-Dest"] = "document"
	loginHeaders["Accept"] = "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"

	loginStatus, _, err := s.Do("GET", targetLoginPageURL, loginHeaders, nil)
	if err != nil {
		return fmt.Errorf("login page request failed: %w", err)
	}
	log.Printf("[session] login page: status=%d", loginStatus)

	// Extract TealeafAkaSid for use in device_info fingerprints.
	s.tealeafID = s.getCookieValue("https://target.com", "TealeafAkaSid")
	if s.tealeafID != "" {
		log.Printf("[session] tealeafId: %s", s.tealeafID)
	}

	loginSession := s.getCookieValue("https://target.com", "login-session")
	log.Printf("[session] login-session present after warm-up: %v", loginSession != "")

	return nil
}

// Login authenticates with a Target account using a real headless browser
// (go-rod) to bypass PerimeterX, then injects the resulting cookies into
// the TLS client for fast ATC/checkout API calls.
func (s *TargetSession) Login(email, password string) error {
	result, err := BrowserLogin(email, password)
	if err != nil {
		return fmt.Errorf("browser login failed: %w", err)
	}

	s.injectCookies(result.Cookies)

	if result.VisitorID != "" {
		s.VisitorID = result.VisitorID
		log.Printf("[session] using browser visitorId: %s", s.VisitorID)
	}

	log.Println("[session] login complete, browser cookies injected into TLS client")
	return nil
}

// cookieInjectionURLs lists the Target domain URLs where browser cookies
// need to be set in the TLS client's cookie jar.
var cookieInjectionURLs = []string{
	"https://www.target.com",
	"https://gsp.target.com",
	"https://carts.target.com",
}

// injectCookies takes standard library cookies (from the browser) and sets
// them on the TLS client's cookie jar for each relevant Target domain.
func (s *TargetSession) injectCookies(cookies []*stdhttp.Cookie) {
	for _, rawURL := range cookieInjectionURLs {
		u, err := url.Parse(rawURL)
		if err != nil {
			continue
		}

		fhttpCookies := make([]*fhttp.Cookie, 0, len(cookies))
		for _, c := range cookies {
			fhttpCookies = append(fhttpCookies, &fhttp.Cookie{
				Name:     c.Name,
				Value:    c.Value,
				Domain:   c.Domain,
				Path:     c.Path,
				Expires:  c.Expires,
				Secure:   c.Secure,
				HttpOnly: c.HttpOnly,
			})
		}

		s.client.SetCookies(u, fhttpCookies)
	}

	log.Printf("[session] injected %d cookies across %d domains", len(cookies), len(cookieInjectionURLs))
}

// credentialValidation submits email and password to Target's auth service.
// Returns the OAuth authorization code from the 202 response body.
func (s *TargetSession) credentialValidation(email, password string) (string, error) {
	payload := models.LoginRequest{
		Username:       email,
		Password:       password,
		DeviceInfo:     models.DefaultDeviceInfo(s.VisitorID, s.tealeafID),
		KeepMeSignedIn: false,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	status, respBody, err := s.Do("POST", credValidationsURL, gspHeaders(), body)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	log.Printf("[session] credential_validations: status=%d body=%s", status, string(respBody))

	if status != 202 {
		return "", fmt.Errorf("unexpected status %d: %s", status, string(respBody))
	}

	var resp models.CredValidationResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	if resp.Code == "" {
		return "", fmt.Errorf("no authorization code in response: %s", string(respBody))
	}

	log.Printf("[session] credential_validations: got auth code")
	return resp.Code, nil
}

// skipPhoneVerification posts to the skip-2FA endpoint. Target shows this
// prompt on accounts that haven't registered a phone number.
func (s *TargetSession) skipPhoneVerification() error {
	status, respBody, err := s.Do("POST", skipPhoneURL, gspHeaders(), nil)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	log.Printf("[session] skip_phone_verifications: status=%d", status)
	if status != 200 {
		return fmt.Errorf("unexpected status %d: %s", status, string(respBody))
	}
	return nil
}

// fetchClientTokens posts to the OAuth token endpoint.
// grantType is either "client_credentials" (anonymous warm-up) or
// "authorization_code" (authenticated login, requires code).
func (s *TargetSession) fetchClientTokens(grantType, code string) error {
	payload := models.ClientTokensRequest{
		GrantType:        grantType,
		ClientCredential: models.ClientCredential{ClientID: "ecom-web-1.0.0"},
		Merge:            "cart",
		DeviceInfo:       models.DefaultDeviceInfo(s.VisitorID, s.tealeafID),
		Code:             code,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	status, respBody, err := s.Do("POST", clientTokensURL, gspHeaders(), body)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	log.Printf("[session] client_tokens (%s): status=%d body=%s", grantType, status, string(respBody))

	if status != 201 {
		return fmt.Errorf("unexpected status %d: %s", status, string(respBody))
	}
	return nil
}

// validateTokens confirms the issued tokens are valid. The server reads the
// accessToken cookie; the request body is an empty JSON object.
func (s *TargetSession) validateTokens() error {
	status, respBody, err := s.Do("POST", tokenValidationsURL, gspHeaders(), []byte("{}"))
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	log.Printf("[session] token_validations: status=%d", status)
	if status != 200 {
		return fmt.Errorf("unexpected status %d: %s", status, string(respBody))
	}
	return nil
}

// compile-time check: TargetSession must satisfy the Session interface.
var _ Session = (*TargetSession)(nil)

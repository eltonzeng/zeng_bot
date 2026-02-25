package session

import (
	"fmt"
	"log"
	stdhttp "net/http"
	"os"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// BrowserLoginResult holds the cookies and visitor ID extracted from a
// successful browser-based login to Target.
type BrowserLoginResult struct {
	Cookies   []*stdhttp.Cookie
	VisitorID string
}

// BrowserLogin uses a real headless browser (go-rod) to log into Target.
// This bypasses PerimeterX because the browser executes PX's JavaScript sensor
// natively. The resulting cookies are returned for injection into the TLS client.
//
// Set HEADLESS=1 to run without a visible browser window.
func BrowserLogin(email, password string) (*BrowserLoginResult, error) {
	headless := os.Getenv("HEADLESS") == "1"

	l := launcher.New().Headless(headless)
	controlURL, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("failed to launch browser: %w", err)
	}

	browser := rod.New().ControlURL(controlURL)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to browser: %w", err)
	}
	defer browser.MustClose()

	page, err := browser.Page(proto.TargetCreateTarget{URL: "https://www.target.com/login"})
	if err != nil {
		return nil, fmt.Errorf("failed to open login page: %w", err)
	}

	if err := page.WaitLoad(); err != nil {
		return nil, fmt.Errorf("login page did not load: %w", err)
	}
	log.Println("[browser-login] login page loaded")

	// Fill email field.
	emailEl, err := findElement(page, []string{"#username", `input[name="username"]`}, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("email field not found: %w", err)
	}
	if err := emailEl.SelectAllText(); err == nil {
		emailEl.MustInput(email)
	} else {
		emailEl.MustInput(email)
	}
	log.Println("[browser-login] filled email")

	// Fill password field.
	passEl, err := findElement(page, []string{"#password", `input[name="password"]`}, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("password field not found: %w", err)
	}
	passEl.MustInput(password)
	log.Println("[browser-login] filled password")

	// Click sign-in button.
	loginBtn, err := findElement(page, []string{"#login", `button[data-test="login-button"]`}, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("login button not found: %w", err)
	}
	loginBtn.MustClick()
	log.Println("[browser-login] clicked sign-in")

	// Wait for network to settle after login.
	time.Sleep(3 * time.Second)
	page.MustWaitStable()

	// Best-effort: dismiss phone verification modal.
	handlePhoneVerification(page)

	// Verify login succeeded by checking for a logged-in indicator.
	loggedIn := verifyLoggedIn(page)
	if !loggedIn {
		log.Println("[browser-login] warning: could not confirm login success via DOM indicator")
	} else {
		log.Println("[browser-login] login confirmed via DOM indicator")
	}

	// Extract cookies from the browser.
	browserCookies, err := browser.GetCookies()
	if err != nil {
		return nil, fmt.Errorf("failed to get browser cookies: %w", err)
	}
	log.Printf("[browser-login] extracted %d cookies from browser", len(browserCookies))

	cookies := convertBrowserCookies(browserCookies)

	// Attempt to extract visitorId from localStorage.
	visitorID := extractVisitorID(page)
	if visitorID != "" {
		log.Printf("[browser-login] extracted visitorId: %s", visitorID)
	}

	return &BrowserLoginResult{
		Cookies:   cookies,
		VisitorID: visitorID,
	}, nil
}

// findElement tries multiple CSS selectors in order, returning the first
// element found within the timeout.
func findElement(page *rod.Page, selectors []string, timeout time.Duration) (*rod.Element, error) {
	for _, sel := range selectors {
		el, err := page.Timeout(timeout).Element(sel)
		if err == nil {
			return el, nil
		}
	}
	return nil, fmt.Errorf("none of the selectors matched: %v", selectors)
}

// handlePhoneVerification dismisses the phone verification modal if it
// appears after login. This is best-effort with a short timeout.
func handlePhoneVerification(page *rod.Page) {
	noThanksSelectors := []string{
		`button[data-test="no-thanks-button"]`,
		`button:has-text("No Thanks")`,
		`button:has-text("Not now")`,
	}
	el, err := findElement(page, noThanksSelectors, 3*time.Second)
	if err != nil {
		return // No modal appeared, that's fine.
	}
	el.MustClick()
	log.Println("[browser-login] dismissed phone verification modal")
}

// verifyLoggedIn checks for DOM elements that indicate a successful login.
func verifyLoggedIn(page *rod.Page) bool {
	indicators := []string{
		`[data-test="accountNav-toggle"]`,
		`[data-test="@web/AccountLink"]`,
		`#account`,
	}
	_, err := findElement(page, indicators, 8*time.Second)
	return err == nil
}

// extractVisitorID attempts to read visitorId from the page's localStorage.
func extractVisitorID(page *rod.Page) string {
	val, err := page.Eval(`() => {
		try {
			return localStorage.getItem('visitorId') || '';
		} catch(e) {
			return '';
		}
	}`)
	if err != nil || val.Value.Str() == "" {
		return ""
	}
	return val.Value.Str()
}

// convertBrowserCookies maps rod's proto.NetworkCookie slice to standard
// library []*net/http.Cookie values.
func convertBrowserCookies(rodCookies []*proto.NetworkCookie) []*stdhttp.Cookie {
	result := make([]*stdhttp.Cookie, 0, len(rodCookies))
	for _, rc := range rodCookies {
		c := &stdhttp.Cookie{
			Name:     rc.Name,
			Value:    rc.Value,
			Domain:   rc.Domain,
			Path:     rc.Path,
			Secure:   rc.Secure,
			HttpOnly: rc.HTTPOnly,
		}
		if rc.Expires > 0 {
			c.Expires = time.Unix(int64(rc.Expires), 0)
		}
		result = append(result, c)
	}
	return result
}

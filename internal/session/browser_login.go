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
// Target uses a multi-step login flow:
//  1. Click "Sign in or create account" to open the login modal
//  2. Enter email → click "Continue"
//  3. Select "password" auth factor (vs passkey)
//  4. Enter password → submit
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

	// Wait for initial page load.
	time.Sleep(5 * time.Second)
	log.Println("[browser-login] login page loaded")

	// Step 1: Click "Sign in or create account" to open the login modal.
	// This button triggers the modal with the email input.
	signinTrigger, err := findElement(page, []string{
		`button[class*="ndsButton"][class*="filled"]`,
		`button[data-test="accountNav-signIn"]`,
	}, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("sign-in trigger button not found: %w", err)
	}
	signinTrigger.MustClick()
	log.Println("[browser-login] opened sign-in modal")
	time.Sleep(2 * time.Second)

	// Step 2: Fill email and click "Continue".
	emailEl, err := findElement(page, []string{"#username", `input[name="username"]`}, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("email field not found: %w", err)
	}
	emailEl.MustInput(email)
	log.Println("[browser-login] filled email")

	continueBtn, err := findElement(page, []string{"#login", `button[type="submit"]`}, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("continue button not found: %w", err)
	}
	continueBtn.MustClick()
	log.Println("[browser-login] clicked continue")
	time.Sleep(3 * time.Second)

	// Step 3: Select the "password" auth factor radio if it appears.
	// (Target may offer passkey vs password choice for existing accounts.)
	if pwRadio, err := page.Timeout(3 * time.Second).Element(`#password-checkbox`); err == nil {
		pwRadio.MustClick()
		log.Println("[browser-login] selected password auth factor")
		time.Sleep(2 * time.Second)
	}

	// Step 4: Fill the password field and submit.
	passEl, err := findElement(page, []string{
		`input[type="password"]`,
		`#password`,
		`input[name="password"]`,
	}, 8*time.Second)
	if err != nil {
		return nil, fmt.Errorf("password field not found: %w", err)
	}
	passEl.MustInput(password)
	log.Println("[browser-login] filled password")

	submitBtn, err := findElement(page, []string{
		`#login`,
		`button[data-test="form-submit-button"]`,
		`button[type="submit"]`,
	}, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("submit button not found: %w", err)
	}
	submitBtn.MustClick()
	log.Println("[browser-login] submitted login form")

	// Wait for post-login navigation to settle.
	time.Sleep(5 * time.Second)

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
		`button[id="no-thanks"]`,
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

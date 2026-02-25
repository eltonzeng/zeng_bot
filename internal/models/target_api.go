package models

// ATCRequest is the JSON payload for adding an item to the Target cart.
// Endpoint: POST https://carts.target.com/web_checkouts/v1/cart_items?field_groups=CART,CART_ITEMS,SUMMARY&key=...
type ATCRequest struct {
	CartItem        ATCCartItem `json:"cart_item"`
	CartType        string      `json:"cart_type"`
	ChannelID       string      `json:"channel_id"`
	ShoppingContext string      `json:"shopping_context"`
}

// ATCCartItem is the item payload within an ATCRequest.
type ATCCartItem struct {
	TCIN          string `json:"tcin"`
	Quantity      int    `json:"quantity"`
	ItemChannelID string `json:"item_channel_id"`
}

// ATCResponse is the parsed response from the Target cart API.
type ATCResponse struct {
	CartID string `json:"cart_id"`
}

// OrderRequest is the JSON payload for submitting a Target order.
type OrderRequest struct {
	CartID          string       `json:"cart_id"`
	ChannelID       string       `json:"channel_id"`
	ShippingAddress Address      `json:"shipping_address"`
	BillingAddress  Address      `json:"billing_address"`
	PaymentInfo     OrderPayment `json:"payment_info"`
	ContactInfo     OrderContact `json:"contact_info"`
}

// OrderPayment holds the payment details within an OrderRequest.
type OrderPayment struct {
	CardNumber string `json:"card_number"`
	ExpMonth   string `json:"exp_month"`
	ExpYear    string `json:"exp_year"`
	CVV        string `json:"cvv"`
	CardType   string `json:"card_type"`
}

// OrderContact holds the contact info within an OrderRequest.
type OrderContact struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

// OrderResponse is the parsed response from the Target orders API.
type OrderResponse struct {
	OrderID string `json:"order_id"`
}

// DeviceInfo is the full browser fingerprint included in Target auth requests.
// All fields are string-encoded to match Target's exact wire format.
type DeviceInfo struct {
	UserAgent               string `json:"user_agent"`
	Language                string `json:"language"`
	Canvas                  string `json:"canvas"`
	ColorDepth              string `json:"color_depth"`
	DeviceMemory            string `json:"device_memory"`
	PixelRatio              string `json:"pixel_ratio"`
	HardwareConcurrency     string `json:"hardware_concurrency"`
	Resolution              string `json:"resolution"`
	AvailableResolution     string `json:"available_resolution"`
	TimezoneOffset          string `json:"timezone_offset"`
	SessionStorage          string `json:"session_storage"`
	LocalStorage            string `json:"local_storage"`
	IndexedDB               string `json:"indexed_db"`
	AddBehavior             string `json:"add_behavior"`
	OpenDatabase            string `json:"open_database"`
	CPUClass                string `json:"cpu_class"`
	NavigatorPlatform       string `json:"navigator_platform"`
	DoNotTrack              string `json:"do_not_track"`
	RegularPlugins          string `json:"regular_plugins"`
	Adblock                 string `json:"adblock"`
	HasLiedLanguages        string `json:"has_lied_languages"`
	HasLiedResolution       string `json:"has_lied_resolution"`
	HasLiedOS               string `json:"has_lied_os"`
	HasLiedBrowser          string `json:"has_lied_browser"`
	TouchSupport            string `json:"touch_support"`
	JSFonts                 string `json:"js_fonts"`
	NavigatorVendor         string `json:"navigator_vendor"`
	NavigatorWebdriver      string `json:"navigator_webdriver"`
	NavigatorAppName        string `json:"navigator_app_name"`
	NavigatorAppCodeName    string `json:"navigator_app_code_name"`
	NavigatorAppVersion     string `json:"navigator_app_version"`
	NavigatorLanguages      string `json:"navigator_languages"`
	NavigatorCookiesEnabled string `json:"navigator_cookies_enabled"`
	NavigatorJavaEnabled    string `json:"navigator_java_enabled"`
	VisitorID               string `json:"visitor_id"`
	TealeafID               string `json:"tealeaf_id"`
	WebGL                   string `json:"webgl"`
	WebGLVendor             string `json:"webgl_vendor"`
	BrowserName             string `json:"browser_name"`
	BrowserVersion          string `json:"browser_version"`
	CPUArchitecture         string `json:"cpu_architecture"`
	DeviceVendor            string `json:"device_vendor"`
	DeviceModel             string `json:"device_model"`
	DeviceType              string `json:"device_type"`
	EngineName              string `json:"engine_name"`
	EngineVersion           string `json:"engine_version"`
	OSName                  string `json:"os_name"`
	OSVersion               string `json:"os_version"`
}

// DefaultDeviceInfo returns a hardcoded Chrome 144 / macOS fingerprint,
// parameterised by the session-specific visitorId and tealeafId values.
func DefaultDeviceInfo(visitorID, tealeafID string) DeviceInfo {
	return DeviceInfo{
		UserAgent:               "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36",
		Language:                "en",
		Canvas:                  "f418b62b527756dce2ba14edd74195ce",
		ColorDepth:              "30",
		DeviceMemory:            "8",
		PixelRatio:              "unknown",
		HardwareConcurrency:     "8",
		Resolution:              "[1440,900]",
		AvailableResolution:     "[1440,875]",
		TimezoneOffset:          "480",
		SessionStorage:          "1",
		LocalStorage:            "1",
		IndexedDB:               "1",
		AddBehavior:             "unknown",
		OpenDatabase:            "unknown",
		CPUClass:                "unknown",
		NavigatorPlatform:       "MacIntel",
		DoNotTrack:              "unknown",
		RegularPlugins:          `["PDF Viewer::Portable Document Format::application/pdf~pdf,text/pdf~pdf","Chrome PDF Viewer::Portable Document Format::application/pdf~pdf,text/pdf~pdf","Chromium PDF Viewer::Portable Document Format::application/pdf~pdf,text/pdf~pdf","Microsoft Edge PDF Viewer::Portable Document Format::application/pdf~pdf,text/pdf~pdf","WebKit built-in PDF::Portable Document Format::application/pdf~pdf,text/pdf~pdf"]`,
		Adblock:                 "false",
		HasLiedLanguages:        "false",
		HasLiedResolution:       "false",
		HasLiedOS:               "false",
		HasLiedBrowser:          "false",
		TouchSupport:            "[0,false,false]",
		JSFonts:                 `["Andale Mono","Arial","Arial Black","Arial Hebrew","Arial Narrow","Arial Rounded MT Bold","Arial Unicode MS","Comic Sans MS","Courier","Courier New","Geneva","Georgia","Helvetica","Helvetica Neue","Impact","LUCIDA GRANDE","Microsoft Sans Serif","Monaco","Palatino","Tahoma","Times","Times New Roman","Trebuchet MS","Verdana","Wingdings","Wingdings 2","Wingdings 3"]`,
		NavigatorVendor:         "Google Inc.",
		NavigatorWebdriver:      "false",
		NavigatorAppName:        "Netscape",
		NavigatorAppCodeName:    "Mozilla",
		NavigatorAppVersion:     "5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36",
		NavigatorLanguages:      `["en","zh","zh-CN"]`,
		NavigatorCookiesEnabled: "true",
		NavigatorJavaEnabled:    "false",
		VisitorID:               visitorID,
		TealeafID:               tealeafID,
		WebGL:                   "unknown",
		WebGLVendor:             "unknown",
		BrowserName:             "Unknown",
		BrowserVersion:          "Unknown",
		CPUArchitecture:         "Unknown",
		DeviceVendor:            "Unknown",
		DeviceModel:             "Unknown",
		DeviceType:              "Unknown",
		EngineName:              "Unknown",
		EngineVersion:           "Unknown",
		OSName:                  "Unknown",
		OSVersion:               "Unknown",
	}
}

// LoginRequest is the JSON payload for credential_validations.
// Endpoint: POST https://gsp.target.com/gsp/authentications/v1/credential_validations?client_id=ecom-web-1.0.0
type LoginRequest struct {
	Username       string     `json:"username"`
	Password       string     `json:"password"`
	DeviceInfo     DeviceInfo `json:"device_info"`
	KeepMeSignedIn bool       `json:"keep_me_signed_in"`
}

// CredValidationResponse is the parsed 202 response from credential_validations.
// The Code is an OAuth authorization code passed to client_tokens.
type CredValidationResponse struct {
	Code string `json:"code"`
}

// ClientCredential identifies the application in an OAuth token request.
type ClientCredential struct {
	ClientID string `json:"client_id"`
}

// ClientTokensRequest is the OAuth payload for client_tokens.
// Use grant_type "client_credentials" for anonymous sessions and
// "authorization_code" (with Code) to upgrade to an authenticated session.
// Endpoint: POST https://gsp.target.com/gsp/oauth_tokens/v2/client_tokens
type ClientTokensRequest struct {
	GrantType        string           `json:"grant_type"`
	ClientCredential ClientCredential `json:"client_credential"`
	Merge            string           `json:"merge"`
	DeviceInfo       DeviceInfo       `json:"device_info"`
	Code             string           `json:"code,omitempty"`
}

// DefaultHeaders returns the standard browser-mimicking headers required
// for Target API requests to pass PerimeterX validation.
func DefaultHeaders() map[string]string {
	return map[string]string{
		"User-Agent":         "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36",
		"Accept":             "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
		"Accept-Language":    "en,zh;q=0.9,zh-CN;q=0.8",
		"Accept-Encoding":    "gzip, deflate, br",
		"sec-ch-ua":          `"Not(A:Brand";v="8", "Chromium";v="144", "Google Chrome";v="144"`,
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": `"macOS"`,
		"Sec-Fetch-Site":     "same-origin",
		"Sec-Fetch-Mode":     "cors",
		"Sec-Fetch-Dest":     "empty",
	}
}

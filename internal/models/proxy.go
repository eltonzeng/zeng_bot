package models

import "fmt"

// Proxy represents a single proxy endpoint used for HTTP requests.
type Proxy struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// URL returns the proxy formatted as an HTTP URL string.
func (p Proxy) URL() string {
	if p.Username != "" && p.Password != "" {
		return fmt.Sprintf("http://%s:%s@%s:%s", p.Username, p.Password, p.Host, p.Port)
	}
	return fmt.Sprintf("http://%s:%s", p.Host, p.Port)
}

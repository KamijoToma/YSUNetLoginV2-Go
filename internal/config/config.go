package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/viper"
	"golang.org/x/term"
)

// Config holds all configuration for the application
type Config struct {
	Username string
	Password string
	Service  string
	Proxies  map[string]string
	Verbose  bool
}

// ServiceMapping maps aliases to actual service names
var ServiceMapping = map[string]string{
	"campus":  "校园网",
	"unicom":  "中国联通",
	"telecom": "中国电信",
	"mobile":  "中国移动",
	"1":       "校园网",
	"2":       "中国联通",
	"3":       "中国电信",
	"4":       "中国移动",
}

// NewConfig creates a new configuration instance
func NewConfig() *Config {
	return &Config{
		Service: "校园网",
		Proxies: make(map[string]string),
	}
}

// LoadFromViper loads configuration from viper (environment variables and config files)
func (c *Config) LoadFromViper() {
	c.Username = viper.GetString("username")
	c.Password = viper.GetString("password")
	c.Service = viper.GetString("service")
	c.Verbose = viper.GetBool("verbose")

	// Set default service if empty
	if c.Service == "" {
		c.Service = "校园网"
	}

	// Load proxy settings
	if httpProxy := viper.GetString("http_proxy"); httpProxy != "" {
		c.Proxies["http"] = httpProxy
	}
	if httpsProxy := viper.GetString("https_proxy"); httpsProxy != "" {
		c.Proxies["https"] = httpsProxy
	}
	if proxy := viper.GetString("proxy"); proxy != "" {
		c.Proxies["http"] = proxy
		c.Proxies["https"] = proxy
	}
}

// UpdateFromFlags updates configuration from command line flags
func (c *Config) UpdateFromFlags(username, password, service, proxy string, verbose bool) {
	if username != "" {
		c.Username = username
	}
	if password != "" {
		c.Password = password
	}
	if service != "" {
		c.Service = service
	}
	if proxy != "" {
		c.Proxies["http"] = proxy
		c.Proxies["https"] = proxy
	}
	if verbose {
		c.Verbose = verbose
	}
}

// ValidateCredentials checks if username and password are provided
func (c *Config) ValidateCredentials() bool {
	return c.Username != "" && c.Password != ""
}

// GetCredentialsInteractive prompts user for credentials if not provided
func (c *Config) GetCredentialsInteractive() error {
	reader := bufio.NewReader(os.Stdin)

	if c.Username == "" {
		fmt.Print("Username: ")
		username, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read username: %w", err)
		}
		c.Username = strings.TrimSpace(username)
	}

	if c.Password == "" {
		fmt.Print("Password: ")
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		fmt.Println() // Print newline after password input
		c.Password = strings.TrimSpace(string(bytePassword))
	}

	return nil
}

// ResolveServiceName resolves service name from aliases
func (c *Config) ResolveServiceName(serviceInput string) string {
	if serviceInput == "" {
		return c.Service
	}

	// Direct Chinese service names
	chineseServices := []string{"校园网", "中国联通", "中国电信", "中国移动"}
	for _, service := range chineseServices {
		if serviceInput == service {
			return serviceInput
		}
	}

	// Check mapping
	if mapped, exists := ServiceMapping[strings.ToLower(serviceInput)]; exists {
		return mapped
	}

	// Return original input if no mapping found
	return serviceInput
}

// GetErrorMessage converts errors to user-friendly messages
func GetErrorMessage(err error) string {
	errMsg := strings.ToLower(err.Error())

	// Network related errors
	if strings.Contains(errMsg, "connection") || strings.Contains(errMsg, "timeout") {
		return "Network connection failed. Please check your internet connection."
	}

	// Authentication related errors
	if strings.Contains(errMsg, "authentication failed") || strings.Contains(errMsg, "cas") {
		return fmt.Sprintf("Authentication failed. Detail: %s", errMsg)
	}

	// API related errors
	if strings.Contains(errMsg, "api error") {
		return fmt.Sprintf("Server error: %s", err.Error())
	}

	// Portal redirection errors
	if strings.Contains(errMsg, "portal redirection failed") {
		return "Portal access failed. You may not be connected to the campus network."
	}

	// CAS redirection errors
	if strings.Contains(errMsg, "cas redirection failed") {
		return "CAS authentication failed. Please try again."
	}

	// Captcha related errors
	if strings.Contains(errMsg, "captcha") || strings.Contains(errMsg, "验证码") {
		return "Captcha verification failed. Please try again."
	}

	// Default error message
	return fmt.Sprintf("Operation failed: %s", err.Error())
}

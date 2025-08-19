package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"ruijie-go/internal/utils"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
)

// CASClient handles YSU CAS authentication
type CASClient struct {
	client      *resty.Client
	username    string
	password    string
	displayMode utils.CaptchaDisplayMode
	verbose     bool

	// Form parameters
	lt        string
	execution string
	salt      string
	cllt      string
	dllt      string
	eventID   string
	captcha   string
}

// CAS URLs
const (
	LoginURL        = "https://cer.ysu.edu.cn/authserver/login?service=https%3A%2F%2Fehall.ysu.edu.cn%2Flogin"
	CheckCaptchaURL = "https://cer.ysu.edu.cn/authserver/checkNeedCaptcha.htl"
	CaptchaURL      = "https://cer.ysu.edu.cn/authserver/getCaptcha.htl"
)

// NewCASClient creates a new CAS authentication client
func NewCASClient(username, password string, client *resty.Client, displayMode utils.CaptchaDisplayMode, verbose bool) *CASClient {
	if client == nil {
		client = resty.New()
		client.SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	}

	return &CASClient{
		client:      client,
		username:    username,
		password:    password,
		displayMode: displayMode,
		verbose:     verbose,
	}
}

// log outputs debug information if verbose mode is enabled
func (c *CASClient) log(message string) {
	if c.verbose {
		fmt.Printf("[DEBUG] %s\n", message)
	}
}

// fetchLoginPage fetches the login page and extracts form parameters
func (c *CASClient) fetchLoginPage() error {
	c.log("Fetching login page...")

	resp, err := c.client.R().Get(LoginURL)
	if err != nil {
		return fmt.Errorf("failed to fetch login page: %w", err)
	}

	// Parse HTML to extract form parameters
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
	if err != nil {
		return fmt.Errorf("failed to parse login page: %w", err)
	}

	// Find the password login form
	form := doc.Find("form#pwdFromId")
	if form.Length() == 0 {
		return fmt.Errorf("login form with ID 'pwdFromId' not found")
	}

	// Extract form parameters
	c.lt = form.Find("input[name='lt']").AttrOr("value", "")
	c.execution = form.Find("input[name='execution']").AttrOr("value", "")
	c.salt = form.Find("input#pwdEncryptSalt").AttrOr("value", "")
	c.cllt = form.Find("input[name='cllt']").AttrOr("value", "userNameLogin")
	c.dllt = form.Find("input[name='dllt']").AttrOr("value", "generalLogin")
	c.eventID = form.Find("input[name='_eventId']").AttrOr("value", "submit")

	if c.execution == "" || c.salt == "" {
		return fmt.Errorf("failed to extract required form parameters")
	}

	c.log(fmt.Sprintf("Extracted form parameters: lt=%s, execution=%s, salt=%s", c.lt, c.execution, c.salt))
	return nil
}

// needCaptcha checks if captcha is required for the username
func (c *CASClient) needCaptcha() (bool, error) {
	c.log("Checking if captcha is required...")

	resp, err := c.client.R().
		SetFormData(map[string]string{
			"username": c.username,
		}).
		Post(CheckCaptchaURL)

	if err != nil {
		c.log(fmt.Sprintf("Warning: Failed to check captcha requirement: %v", err))
		return false, nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		c.log(fmt.Sprintf("Warning: Failed to parse captcha check response: %v", err))
		return false, nil
	}

	isNeed, ok := result["isNeed"].(bool)
	if !ok {
		return false, nil
	}

	c.log(fmt.Sprintf("Captcha required: %v", isNeed))
	return isNeed, nil
}

// fetchCaptcha downloads and displays captcha, then prompts for user input
func (c *CASClient) fetchCaptcha() error {
	c.log("Fetching captcha...")

	resp, err := c.client.R().Get(CaptchaURL)
	if err != nil {
		return fmt.Errorf("failed to fetch captcha: %w", err)
	}

	imageData := resp.Body()
	captcha, err := utils.DisplayCaptcha(imageData, c.displayMode)
	if err != nil {
		return fmt.Errorf("failed to display captcha: %w", err)
	}

	c.captcha = captcha
	return nil
}

// Login performs the complete CAS login process
func (c *CASClient) Login() error {
	// Step 1: Fetch login page and extract parameters
	if err := c.fetchLoginPage(); err != nil {
		return err
	}

	// Step 2: Check if captcha is needed
	needCaptcha, err := c.needCaptcha()
	if err != nil {
		return err
	}

	// Step 3: Handle captcha if required
	if needCaptcha {
		if err := c.fetchCaptcha(); err != nil {
			return fmt.Errorf("captcha required but failed to fetch: %w", err)
		}
	}

	// Step 4: Encrypt password
	encryptedPassword, err := utils.EncryptPassword(c.password, c.salt)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}

	// Step 5: Submit login form
	c.log("Submitting login form...")
	formData := map[string]string{
		"username":  c.username,
		"password":  encryptedPassword,
		"captcha":   c.captcha,
		"lt":        c.lt,
		"execution": c.execution,
		"_eventId":  c.eventID,
		"cllt":      c.cllt,
		"dllt":      c.dllt,
	}

	resp, err := c.client.R().
		SetFormData(formData).
		SetDoNotParseResponse(true).
		Post(LoginURL)

	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}

	// Step 6: Check login result
	if resp.StatusCode() == 302 {
		// Success - should redirect
		location := resp.Header().Get("Location")
		if location != "" {
			c.log(fmt.Sprintf("Login successful! Redirecting to: %s", location))

			// Follow the redirect to confirm
			finalResp, err := c.client.R().Get(location)
			if err != nil {
				return fmt.Errorf("failed to follow redirect: %w", err)
			}

			if !strings.Contains(finalResp.String(), "统一身份认证") {
				c.log("Login confirmed successfully")
				return nil
			} else {
				return fmt.Errorf("login appeared successful but redirected back to login page")
			}
		}
	}

	// Login failed - extract error message
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body()))
	if err != nil {
		return fmt.Errorf("login failed and could not parse error response")
	}

	errorSpan := doc.Find("span#showErrorTip")
	if errorSpan.Length() > 0 {
		errorMsg := strings.TrimSpace(errorSpan.Text())
		return fmt.Errorf("login failed: %s", errorMsg)
	}

	return fmt.Errorf("login failed with no clear error message")
}

package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"ruijie-go/internal/utils"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
)

// RuijieClient handles Ruijie network authentication
type RuijieClient struct {
	client  *resty.Client
	proxies map[string]string
	verbose bool
}

// NewRuijieClient creates a new Ruijie client
func NewRuijieClient(proxies map[string]string, verbose bool) *RuijieClient {
	client := resty.New()
	client.SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	// Set proxy if provided
	if len(proxies) > 0 {
		if httpProxy, ok := proxies["http"]; ok {
			client.SetProxy(httpProxy)
		} else if httpsProxy, ok := proxies["https"]; ok {
			client.SetProxy(httpsProxy)
		}
	}

	return &RuijieClient{
		client:  client,
		proxies: proxies,
		verbose: verbose,
	}
}

// log outputs debug information if verbose mode is enabled
func (r *RuijieClient) log(message string) {
	if r.verbose {
		fmt.Printf("[DEBUG] %s\n", message)
	}
}

// unwrapResponse processes API responses and handles errors
func (r *RuijieClient) unwrapResponse(resp *resty.Response, jsonResponse bool) (interface{}, error) {
	if resp.IsError() {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status())
	}

	if jsonResponse {
		var data map[string]interface{}
		if err := json.Unmarshal(resp.Body(), &data); err != nil {
			return nil, fmt.Errorf("failed to parse JSON response: %w", err)
		}

		if code, ok := data["code"].(float64); ok && code == 200 {
			return data["data"], nil
		}

		message, _ := data["message"].(string)
		return nil, fmt.Errorf("API error: %s", message)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return result, nil
}

// GetOnlineUserInfo gets current online user information
func (r *RuijieClient) GetOnlineUserInfo(sessionID string) (map[string]interface{}, error) {
	if sessionID == "" {
		sessionID = "114514"
	}

	timestamp := time.Now().UnixMilli()
	url := fmt.Sprintf("https://auth1.ysu.edu.cn/eportal/adaptor/getOnlineUserInfo?sessionId=%s&%d&version=this%%20is%%20a%%20git-commit", sessionID, timestamp)

	resp, err := r.client.R().Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get online user info: %w", err)
	}

	data, err := r.unwrapResponse(resp, true)
	if err != nil {
		return nil, err
	}

	return data.(map[string]interface{}), nil
}

// RedirectToPortal redirects to portal and extracts session information
func (r *RuijieClient) RedirectToPortal(redirectURL string) (map[string]string, error) {
	if redirectURL == "" {
		redirectURL = "https://auth1.ysu.edu.cn/eportal/redirect.jsp?mode=history"
	}

	resp, err := r.client.R().Get(redirectURL)
	if err != nil {
		return nil, fmt.Errorf("failed to redirect to portal: %w", err)
	}

	// Get final URL after redirects
	finalURL := resp.RawResponse.Request.URL.String()
	r.log(fmt.Sprintf("Portal redirect final URL: %s", finalURL))

	// Handle JavaScript redirect (e.g. top.self.location.href='...')
	if strings.Contains(resp.String(), "location.href=") {
		content := resp.String()
		start := strings.Index(content, "location.href='")
		if start >= 0 {
			start += len("location.href='")
			end := strings.Index(content[start:], "'")
			if end > 0 {
				redirectURL2 := content[start : start+end]
				r.log(fmt.Sprintf("Following JS redirect to: %s", redirectURL2))
				resp, err = r.client.R().Get(redirectURL2)
				if err != nil {
					return nil, fmt.Errorf("failed to follow JavaScript redirect: %w", err)
				}
				finalURL = resp.RawResponse.Request.URL.String()
				r.log(fmt.Sprintf("JS redirect final URL: %s", finalURL))
			}
		}
	}

	if !strings.Contains(finalURL, "portal-main") {
		return nil, fmt.Errorf("portal redirection failed. Expected URL to contain 'portal-main', but got: %s", finalURL)
	}

	// Parse URL parameters
	parsedURL, err := url.Parse(finalURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse portal URL: %w", err)
	}

	params := make(map[string]string)
	for key, values := range parsedURL.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	return params, nil
}

// getCurrentNode gets current workflow node
func (r *RuijieClient) getCurrentNode(sessionInfo map[string]string, flowKey string) (map[string]interface{}, error) {
	if flowKey == "" {
		flowKey = "portal_auth"
	}

	nodeURL := "https://auth1.ysu.edu.cn/eportal/workFlow/getCurrentNode"
	requestData := map[string]interface{}{
		"sessionId": sessionInfo["sessionId"],
		"flowKey":   flowKey,
	}

	resp, err := r.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(requestData).
		Post(nodeURL)

	if err != nil {
		return nil, fmt.Errorf("failed to get current node: %w", err)
	}

	var nodeResp map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &nodeResp); err != nil {
		return nil, fmt.Errorf("failed to parse node response: %w", err)
	}

	if data, ok := nodeResp["data"].(map[string]interface{}); ok {
		if currentNode, ok := data["currentNodePath"].(string); ok {
			r.log(fmt.Sprintf("Current Node: %s", currentNode))
		}
	}

	return nodeResp, nil
}

// CasSSOLogin performs direct CAS-SSO authentication (new method replacing CAS+SAM)
func (r *RuijieClient) CasSSOLogin(username, password string, sessionInfo map[string]string) error {
	sessionID := sessionInfo["sessionId"]
	customPageID := sessionInfo["customPageId"]
	nasIP := sessionInfo["nasIp"]
	userIP := sessionInfo["userIp"]
	ssid := sessionInfo["ssid"]
	mode := sessionInfo["mode"]

	timer := fmt.Sprintf("%d", time.Now().UnixMilli())
	casSSOURL := fmt.Sprintf(
		"https://auth1.ysu.edu.cn/cas-sso/login?flowSessionId=%s&customPageId=%s&preview=false&appType=normal&language=zh-CN&mode=%s&timer=%s&nasIp=%s&userIp=%s&ssid=%s",
		sessionID, customPageID, mode, timer, nasIP, userIP, ssid,
	)

	// Step 1: GET cas-sso/login page to extract croypto and execution
	r.log("Fetching cas-sso login page...")
	resp, err := r.client.R().Get(casSSOURL)
	if err != nil {
		return fmt.Errorf("failed to fetch cas-sso page: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
	if err != nil {
		return fmt.Errorf("failed to parse cas-sso page: %w", err)
	}

	croypto := strings.TrimSpace(doc.Find("p#login-croypto").Text())
	execution := strings.TrimSpace(doc.Find("p#login-page-flowkey").Text())
	if croypto == "" || execution == "" {
		return fmt.Errorf("failed to extract croypto/execution from cas-sso page")
	}
	r.log(fmt.Sprintf("Got croypto: %s..., execution length: %d", croypto[:20], len(execution)))

	// Step 2: Encrypt password with AES-ECB
	encryptedPassword, err := utils.AESEncryptECB(croypto, password)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}
	encryptedCaptcha, err := utils.AESEncryptECB(croypto, "{}")
	if err != nil {
		return fmt.Errorf("failed to encrypt captcha payload: %w", err)
	}

	// Step 3: POST login form
	postURL := casSSOURL + "&accept-language=zh-CN"
	r.log("Submitting cas-sso login form...")
	resp, err = r.client.R().
		SetFormData(map[string]string{
			"username":        username,
			"type":            "UsernamePassword",
			"_eventId":        "submit",
			"geolocation":     "",
			"execution":       execution,
			"captcha_code":    "",
			"croypto":         croypto,
			"password":        encryptedPassword,
			"captcha_payload": encryptedCaptcha,
		}).
		Post(postURL)
	if err != nil {
		return fmt.Errorf("cas-sso login request failed: %w", err)
	}

	finalURL := resp.RawResponse.Request.URL.String()
	r.log(fmt.Sprintf("Login response URL: %s", finalURL))
	if strings.Contains(finalURL, "auth-success") || strings.Contains(finalURL, "ticket=") {
		r.log("CAS-SSO login succeeded (got ticket)")
		return nil
	}

	// Check for error message in response
	errorDoc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
	if err == nil {
		if errorMsg := strings.TrimSpace(errorDoc.Find("#errorMessage").Text()); errorMsg != "" {
			return fmt.Errorf("login failed: %s", errorMsg)
		}
	}

	return fmt.Errorf("CAS-SSO login failed, final URL: %s", finalURL)
}

// ServiceSelection gets available services
func (r *RuijieClient) ServiceSelection(sessionInfo map[string]string) (interface{}, error) {
	serviceURL := "https://auth1.ysu.edu.cn/eportal/network/serviceSelection"
	requestData := map[string]interface{}{
		"sessionId": sessionInfo["sessionId"],
	}

	resp, err := r.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(requestData).
		Post(serviceURL)

	if err != nil {
		return nil, fmt.Errorf("service selection failed: %w", err)
	}

	r.getCurrentNode(sessionInfo, "portal_auth")

	data, err := r.unwrapResponse(resp, true)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// ServiceLogin logs into specified service
func (r *RuijieClient) ServiceLogin(sessionInfo map[string]string, service string) (map[string]interface{}, error) {
	serviceURL := "https://auth1.ysu.edu.cn/eportal/network/serviceLogin"
	requestData := map[string]interface{}{
		"sessionId": sessionInfo["sessionId"],
		"service":   service,
	}

	resp, err := r.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(requestData).
		Post(serviceURL)

	if err != nil {
		return nil, fmt.Errorf("service login failed: %w", err)
	}

	r.getCurrentNode(sessionInfo, "portal_auth")

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse service login response: %w", err)
	}

	return result, nil
}

// UserOnline checks if user is online
func (r *RuijieClient) UserOnline(sessionInfo map[string]string) (map[string]interface{}, error) {
	onlineURL := "https://auth1.ysu.edu.cn/eportal/network/userOnline"
	requestData := map[string]interface{}{
		"sessionId": sessionInfo["sessionId"],
	}

	resp, err := r.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(requestData).
		Post(onlineURL)

	if err != nil {
		return nil, fmt.Errorf("user online check failed: %w", err)
	}

	data, err := r.unwrapResponse(resp, true)
	if err != nil {
		return nil, err
	}

	return data.(map[string]interface{}), nil
}

// GetAccountInfo gets account information
func (r *RuijieClient) GetAccountInfo(sessionInfo map[string]string) (map[string]interface{}, error) {
	accountURL := "https://auth1.ysu.edu.cn/eportal/operator/getAccountInfo"
	requestData := map[string]interface{}{
		"sessionId": sessionInfo["sessionId"],
	}

	resp, err := r.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(requestData).
		Post(accountURL)

	if err != nil {
		return nil, fmt.Errorf("get account info failed: %w", err)
	}

	data, err := r.unwrapResponse(resp, true)
	if err != nil {
		return nil, err
	}

	return data.(map[string]interface{}), nil
}

// Offline logs user out
func (r *RuijieClient) Offline(sessionInfo map[string]string) (map[string]interface{}, error) {
	offlineURL := "https://auth1.ysu.edu.cn/eportal/network/offline"
	requestData := map[string]interface{}{
		"sessionId": sessionInfo["sessionId"],
	}

	resp, err := r.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(requestData).
		Post(offlineURL)

	if err != nil {
		return nil, fmt.Errorf("offline failed: %w", err)
	}

	data, err := r.unwrapResponse(resp, true)
	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}
	return data.(map[string]interface{}), nil
}

// CheckLoginStatus checks current login status
func (r *RuijieClient) CheckLoginStatus() (bool, interface{}, error) {
	userInfo, err := r.GetOnlineUserInfo("")
	if err != nil {
		r.log(fmt.Sprintf("Error checking login status: %v", err))
		return false, nil, err
	}

	if portalInfo, ok := userInfo["portalOnlineUserInfo"].(map[string]interface{}); ok {
		if redirectURL, exists := portalInfo["redirectUrl"]; exists && redirectURL != nil {
			// Not logged in
			return false, redirectURL, nil
		}
	}

	// Logged in
	return true, userInfo, nil
}

// GetAvailableServices gets available services without logging in
func (r *RuijieClient) GetAvailableServices(username, password string) (interface{}, error) {
	// Check current status
	isLoggedIn, _, err := r.CheckLoginStatus()
	if err != nil {
		return nil, err
	}

	if isLoggedIn {
		// If already logged in, get session info and query services
		sessionInfo, err := r.RedirectToPortal("")
		if err != nil {
			return nil, err
		}
		return r.ServiceSelection(sessionInfo)
	}

	// Redirect to portal to get session info
	sessionInfo, err := r.RedirectToPortal("")
	if err != nil {
		return nil, err
	}
	r.log(fmt.Sprintf("Got session info: %v", sessionInfo))

	// CAS-SSO login
	if err := r.CasSSOLogin(username, password, sessionInfo); err != nil {
		return nil, fmt.Errorf("CAS-SSO authentication failed: %w", err)
	}

	// Get services
	services, err := r.ServiceSelection(sessionInfo)
	if err != nil {
		return nil, err
	}
	r.log(fmt.Sprintf("Available services: %v", services))

	return services, nil
}

// Login performs complete login flow
func (r *RuijieClient) Login(username, password, service string) error {
	// Check current status
	isLoggedIn, _, err := r.CheckLoginStatus()
	if err != nil {
		return err
	}

	if isLoggedIn {
		r.log("Already logged in")
		return nil
	}

	// Redirect to portal to get session info
	sessionInfo, err := r.RedirectToPortal("")
	if err != nil {
		return err
	}
	r.log(fmt.Sprintf("Got session info: %v", sessionInfo))

	// CAS-SSO login
	if err := r.CasSSOLogin(username, password, sessionInfo); err != nil {
		return fmt.Errorf("CAS-SSO authentication failed: %w", err)
	}

	// Get services
	services, err := r.ServiceSelection(sessionInfo)
	if err != nil {
		return err
	}
	r.log(fmt.Sprintf("Available services: %v", services))

	// Login to specified service
	loginResult, err := r.ServiceLogin(sessionInfo, service)
	if err != nil {
		return err
	}
	r.log(fmt.Sprintf("Service login result: %v", loginResult))

	// Verify login status
	onlineStatus, err := r.UserOnline(sessionInfo)
	if err != nil {
		return err
	}
	r.log(fmt.Sprintf("User online status: %v", onlineStatus))

	// Check authentication result
	if code, ok := loginResult["code"].(float64); ok && code == 200 {
		if data, ok := loginResult["data"].(map[string]interface{}); ok {
			if authResult, ok := data["authResult"].(string); ok {
				if authResult == "fail" {
					authMessage, _ := data["authMessage"].(string)
					if authMessage == "" {
						authMessage = "Unknown authentication error"
					}
					return fmt.Errorf("authentication failed: %s", authMessage)
				} else if authResult != "success" {
					return fmt.Errorf("unexpected authentication result: %s", authResult)
				}
			}
		}
	} else {
		return fmt.Errorf("invalid service login response: %v", loginResult)
	}

	// Check online status
	if online, ok := onlineStatus["online"].(bool); !ok || !online {
		message, _ := onlineStatus["message"].(string)
		if message == "" {
			message = "User is not online after authentication"
		}
		return fmt.Errorf("login verification failed: %s", message)
	}

	return nil
}

// Logout performs logout operation
func (r *RuijieClient) Logout() error {
	// Check current status
	isLoggedIn, _, err := r.CheckLoginStatus()
	if err != nil {
		return err
	}

	if !isLoggedIn {
		r.log("Already logged out")
		return nil
	}

	// Redirect to portal to get session info
	sessionInfo, err := r.RedirectToPortal("")
	if err != nil {
		return err
	}
	r.log(fmt.Sprintf("Got session info for logout: %v", sessionInfo))

	// Execute logout
	offlineResult, err := r.Offline(sessionInfo)
	if err != nil {
		return err
	}
	r.log(fmt.Sprintf("Offline result: %v", offlineResult))

	// Verify logout status
	finalStatus, err := r.UserOnline(sessionInfo)
	if err != nil {
		return err
	}
	r.log(fmt.Sprintf("Final user status: %v", finalStatus))

	return nil
}

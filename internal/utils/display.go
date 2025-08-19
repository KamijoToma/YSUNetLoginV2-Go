package utils

import (
	"fmt"
	"strings"
)

// PrintStatusInfo prints user status information
func PrintStatusInfo(userInfo map[string]interface{}) {
	portalInfo, _ := userInfo["portalOnlineUserInfo"].(map[string]interface{})
	onlineInfo, _ := userInfo["onlineUser"].(map[string]interface{})

	var username string
	if portalInfo != nil {
		if name, ok := portalInfo["userName"].(string); ok && name != "" {
			username = name
		} else if userID, ok := portalInfo["userId"].(string); ok {
			username = userID
		}
	}

	if username != "" {
		fmt.Printf("Online: %s", username)

		if portalInfo != nil {
			if service, ok := portalInfo["service"].(string); ok && service != "" {
				fmt.Printf(" (%s)", service)
			}
		}
		fmt.Println()

		if portalInfo != nil {
			if userIP, ok := portalInfo["userIp"].(string); ok && userIP != "" {
				fmt.Printf("IP: %s\n", userIP)
			}
		}

		if onlineInfo != nil {
			if loginTime, ok := onlineInfo["authenticationTime"].(string); ok && loginTime != "" {
				fmt.Printf("Login Time: %s\n", loginTime)
			}
			if location, ok := onlineInfo["nodePhysicalLocation"].(string); ok && location != "" {
				fmt.Printf("Location: %s\n", location)
			}
		}
	} else {
		fmt.Println("Status information unavailable")
	}
}

// PrintAccountInfo prints account information
func PrintAccountInfo(accountInfo map[string]interface{}) {
	if accountInfo == nil {
		fmt.Println("Account information unavailable")
		return
	}

	fmt.Println("Account Information:")

	// Basic fields mapping
	basicFields := map[string]string{
		"name":          "Name",
		"service":       "Service",
		"allowMab":      "MAB Allowed",
		"nosenseEnable": "Nosense Enabled",
		"goLink":        "Portal URL",
	}

	// Display basic information
	for key, label := range basicFields {
		if value, exists := accountInfo[key]; exists && value != nil {
			switch v := value.(type) {
			case bool:
				if v {
					fmt.Printf("  %s: Yes\n", label)
				} else {
					fmt.Printf("  %s: No\n", label)
				}
			case string:
				if v != "" {
					fmt.Printf("  %s: %s\n", label, v)
				}
			default:
				fmt.Printf("  %s: %v\n", label, v)
			}
		}
	}

	// Display account details if available
	if accountInfoList, ok := accountInfo["accountInfo"].([]interface{}); ok {
		fmt.Println("  Details:")
		for _, detail := range accountInfoList {
			if detailMap, ok := detail.(map[string]interface{}); ok {
				title, titleOk := detailMap["title"].(string)
				content, contentOk := detailMap["content"].(string)
				if titleOk && contentOk && title != "" && content != "" {
					fmt.Printf("    %s: %s\n", title, content)
				}
			}
		}
	}

	// Display other fields (excluding already processed and unimportant ones)
	excludedFields := map[string]bool{
		"name":             true,
		"service":          true,
		"allowMab":         true,
		"nosenseEnable":    true,
		"goLink":           true,
		"accountInfo":      true,
		"portalSuccessUrl": true,
	}

	for key, value := range accountInfo {
		if !excludedFields[key] && value != nil && value != "" {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}
}

// PrintServicesList prints available services list
func PrintServicesList(servicesData interface{}) {
	if servicesData == nil {
		fmt.Println("No services available")
		return
	}

	fmt.Println("Available Services:")

	var services []interface{}

	// Extract services from different possible structures
	switch data := servicesData.(type) {
	case map[string]interface{}:
		// Try different possible field names
		if serviceList, ok := data["services"].([]interface{}); ok {
			services = serviceList
		} else if serviceList, ok := data["serviceList"].([]interface{}); ok {
			services = serviceList
		} else if serviceList, ok := data["data"].([]interface{}); ok {
			services = serviceList
		} else {
			// Look for any array field
			for _, value := range data {
				if serviceList, ok := value.([]interface{}); ok && len(serviceList) > 0 {
					services = serviceList
					break
				}
			}
		}
	case []interface{}:
		services = data
	}

	if len(services) == 0 {
		fmt.Println("  No services found in response")
		return
	}

	// Display services list
	for i, service := range services {
		switch s := service.(type) {
		case string:
			fmt.Printf("  %d. %s\n", i+1, s)
		case map[string]interface{}:
			serviceName := ""
			if name, ok := s["name"].(string); ok {
				serviceName = name
			} else if name, ok := s["serviceName"].(string); ok {
				serviceName = name
			} else if name, ok := s["service"].(string); ok {
				serviceName = name
			} else {
				serviceName = fmt.Sprintf("%v", s)
			}
			fmt.Printf("  %d. %s\n", i+1, serviceName)
		default:
			fmt.Printf("  %d. %v\n", i+1, service)
		}
	}

	// Display quick selection mappings
	fmt.Println("\nQuick selection (for non-Chinese terminals):")
	fmt.Println("  campus or 1 -> 校园网")
	fmt.Println("  unicom or 2 -> 中国联通")
	fmt.Println("  telecom or 3 -> 中国电信")
	fmt.Println("  mobile or 4 -> 中国移动")
}

// InteractiveServiceSelection handles interactive service selection
func InteractiveServiceSelection(servicesData interface{}) (string, error) {
	PrintServicesList(servicesData)

	fmt.Print("\nPlease select a service (number/name/alias): ")
	var choice string
	fmt.Scanln(&choice)
	choice = strings.TrimSpace(choice)

	// If it's a number selection
	if len(choice) > 0 && choice[0] >= '1' && choice[0] <= '9' {
		choiceNum := int(choice[0] - '0')

		// From predefined mapping
		if choiceNum >= 1 && choiceNum <= 4 {
			serviceNames := []string{"校园网", "中国联通", "中国电信", "中国移动"}
			return serviceNames[choiceNum-1], nil
		}

		// From actual services list
		var services []interface{}
		switch data := servicesData.(type) {
		case map[string]interface{}:
			for _, value := range data {
				if serviceList, ok := value.([]interface{}); ok && len(serviceList) > 0 {
					services = serviceList
					break
				}
			}
		case []interface{}:
			services = data
		}

		if choiceNum >= 1 && choiceNum <= len(services) {
			service := services[choiceNum-1]
			switch s := service.(type) {
			case string:
				return s, nil
			case map[string]interface{}:
				if name, ok := s["name"].(string); ok {
					return name, nil
				} else if name, ok := s["serviceName"].(string); ok {
					return name, nil
				} else if name, ok := s["service"].(string); ok {
					return name, nil
				}
				return fmt.Sprintf("%v", s), nil
			default:
				return fmt.Sprintf("%v", service), nil
			}
		}
	}

	// Use service mapping for aliases
	serviceMapping := map[string]string{
		"campus":  "校园网",
		"unicom":  "中国联通",
		"telecom": "中国电信",
		"mobile":  "中国移动",
		"1":       "校园网",
		"2":       "中国联通",
		"3":       "中国电信",
		"4":       "中国移动",
	}

	if mapped, exists := serviceMapping[strings.ToLower(choice)]; exists {
		return mapped, nil
	}

	// Return original choice if no mapping found
	if choice == "" {
		fmt.Println("Selection cancelled, using default service: 校园网")
		return "校园网", nil
	}

	return choice, nil
}

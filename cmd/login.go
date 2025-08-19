package cmd

import (
	"fmt"

	"ruijie-go/internal/client"
	"ruijie-go/internal/config"
	"ruijie-go/internal/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	loginUsername string
	loginPassword string
	loginService  string
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to network",
	Long: `Login to the Ruijie network authentication system.

Examples:
  ruijie-go login -u 1145141919810 -p mypassword
  ruijie-go login -s campus
  ruijie-go login  # Interactive login`,
	RunE: runLogin,
}

func init() {
	rootCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringVarP(&loginUsername, "username", "u", "", "Username for authentication")
	loginCmd.Flags().StringVarP(&loginPassword, "password", "p", "", "Password for authentication")
	loginCmd.Flags().StringVarP(&loginService, "service", "s", "", "Service name. Supports aliases: campus/1=校园网, unicom/2=中国联通, telecom/3=中国电信, mobile/4=中国移动")
}

func runLogin(cmd *cobra.Command, args []string) error {
	// Create configuration
	cfg := config.NewConfig()
	cfg.LoadFromViper()
	cfg.UpdateFromFlags(loginUsername, loginPassword, loginService, viper.GetString("proxy"), viper.GetBool("verbose"))

	// Handle service selection
	serviceName := cfg.Service
	if loginService == "" {
		// Check if -s flag was provided without value (interactive service selection)
		if cmd.Flags().Changed("service") && loginService == "" {
			fmt.Println("Fetching available services...")

			// Get credentials first if not provided
			if !cfg.ValidateCredentials() {
				if err := cfg.GetCredentialsInteractive(); err != nil {
					return fmt.Errorf("failed to get credentials: %w", err)
				}
			}

			// Create client and get services
			ruijieClient := client.NewRuijieClient(cfg.Proxies, cfg.Verbose)
			servicesData, err := ruijieClient.GetAvailableServices(cfg.Username, cfg.Password)
			if err != nil {
				return fmt.Errorf("failed to get available services: %w", err)
			}

			// Interactive service selection
			selectedService, err := utils.InteractiveServiceSelection(servicesData)
			if err != nil {
				return fmt.Errorf("service selection failed: %w", err)
			}
			serviceName = selectedService

			// Clear cookies to avoid duplicate login
			// Note: This is handled automatically in Go client
		} else if loginService != "" {
			// Resolve service name from aliases
			serviceName = cfg.ResolveServiceName(loginService)
		}
	} else {
		serviceName = cfg.ResolveServiceName(loginService)
	}

	// Get credentials if not provided
	if !cfg.ValidateCredentials() {
		if err := cfg.GetCredentialsInteractive(); err != nil {
			return fmt.Errorf("failed to get credentials: %w", err)
		}
	}

	// Create Ruijie client
	ruijieClient := client.NewRuijieClient(cfg.Proxies, cfg.Verbose)

	// Execute login
	if err := ruijieClient.Login(cfg.Username, cfg.Password, serviceName); err != nil {
		fmt.Printf("Error: %s\n", config.GetErrorMessage(err))
		return err
	}

	fmt.Printf("Login successful to service: %s\n", serviceName)
	return nil
}

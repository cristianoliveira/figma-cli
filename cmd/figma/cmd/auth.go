package cmd

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/auth"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(authCmd)
}

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication for Figma API",
	Long: `Authenticate with Figma API using Personal Access Tokens (PAT) or OAuth 2.0.

Use 'figma auth login' to start authentication flow.
Use 'figma auth logout' to clear stored credentials.
Use 'figma auth status' to check current authentication state.`,
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Figma API",
	Long: `Authenticate with Figma API using either Personal Access Token (PAT) or OAuth 2.0.

PAT authentication:
  figma auth login --token <your-personal-access-token>

OAuth 2.0 authentication (interactive):
  figma auth login --oauth

If no flag provided, the command will prompt for token.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		cfg, err := GetConfig(cmd)
		if err != nil {
			return err
		}
		logger, err := GetLogger(cmd)
		if err != nil {
			return err
		}
		_ = logger // TODO: use logger

		tokenFlag, _ := cmd.Flags().GetString("token")
		oauthFlag, _ := cmd.Flags().GetBool("oauth")

		mgr, err := auth.NewManager(cfg)
		if err != nil {
			return fmt.Errorf("failed to create token manager: %w", err)
		}

		if oauthFlag {
			// OAuth flow
			return fmt.Errorf("OAuth authentication not yet implemented")
		} else {
			// PAT flow
			var token string
			if tokenFlag != "" {
				token = tokenFlag
			} else {
				// Prompt user for token
				fmt.Print("Enter your Personal Access Token: ")
				_, err := fmt.Scanln(&token)
				if err != nil {
					return fmt.Errorf("failed to read token: %w", err)
				}
			}
			if token == "" {
				return fmt.Errorf("token cannot be empty")
			}
			// Store token as PAT
			authToken := &auth.Token{
				Type:         auth.TokenTypePAT,
				AccessToken:  token,
				RefreshToken: "",
				ExpiresAt:    nil,
			}
			if err := mgr.StoreToken(ctx, authToken); err != nil {
				return fmt.Errorf("failed to store token: %w", err)
			}
			fmt.Println("Authentication successful! Token stored securely.")
		}
		return nil
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear stored authentication credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		cfg, err := GetConfig(cmd)
		if err != nil {
			return err
		}
		mgr, err := auth.NewManager(cfg)
		if err != nil {
			return fmt.Errorf("failed to create token manager: %w", err)
		}
		if err := mgr.ClearToken(ctx); err != nil {
			return fmt.Errorf("failed to clear token: %w", err)
		}
		fmt.Println("Logged out successfully.")
		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		cfg, err := GetConfig(cmd)
		if err != nil {
			return err
		}
		mgr, err := auth.NewManager(cfg)
		if err != nil {
			return fmt.Errorf("failed to create token manager: %w", err)
		}
		token, err := mgr.GetToken(ctx)
		if err != nil {
			return fmt.Errorf("failed to retrieve token: %w", err)
		}
		if token == nil {
			fmt.Println("Not authenticated.")
			fmt.Println("Run 'figma auth login' to authenticate.")
			return nil
		}
		fmt.Printf("Authenticated via %s\n", token.Type)
		if token.ExpiresAt != nil {
			fmt.Printf("Token expires at: %s\n", token.ExpiresAt.Format("2006-01-02 15:04:05"))
		}
		// Mask token for security
		masked := ""
		if len(token.AccessToken) > 8 {
			masked = token.AccessToken[:4] + "..." + token.AccessToken[len(token.AccessToken)-4:]
		} else {
			masked = "***"
		}
		fmt.Printf("Token: %s\n", masked)
		return nil
	},
}

func init() {
	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(logoutCmd)
	authCmd.AddCommand(statusCmd)

	loginCmd.Flags().String("token", "", "Personal Access Token (PAT) for authentication")
	loginCmd.Flags().Bool("oauth", false, "Use OAuth 2.0 flow (interactive)")
}

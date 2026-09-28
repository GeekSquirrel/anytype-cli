package apikey

import (
	"github.com/spf13/cobra"

	apiKeyCreateCmd "github.com/anyproto/anytype-cli/cmd/auth/apikey/create"
	apiKeyGrantCmd "github.com/anyproto/anytype-cli/cmd/auth/apikey/grant"
	apiKeyListCmd "github.com/anyproto/anytype-cli/cmd/auth/apikey/list"
	apiKeyRevokeCmd "github.com/anyproto/anytype-cli/cmd/auth/apikey/revoke"
)

// NewApiKeyCmd creates the auth apikey command
func NewApiKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apikey <command>",
		Short: "Manage API keys for programmatic access",
		Long:  "Create, list, and revoke API keys that can be used for programmatic access to Anytype.",
	}

	// Add subcommands
	cmd.AddCommand(apiKeyCreateCmd.NewCreateCmd())
	cmd.AddCommand(apiKeyGrantCmd.NewGrantCmd())
	cmd.AddCommand(apiKeyListCmd.NewListCmd())
	cmd.AddCommand(apiKeyRevokeCmd.NewRevokeCmd())

	return cmd
}

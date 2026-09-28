package create

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/anyproto/anytype-heart/pkg/lib/pb/model"

	"github.com/anyproto/anytype-cli/cmd/cmdutil"
	"github.com/anyproto/anytype-cli/core"
	"github.com/anyproto/anytype-cli/core/output"
)

// The scope names heart itself prints in auth errors.
var scopeNames = map[string]model.AccountAuthLocalApiScope{
	"limited": model.AccountAuth_Limited,
	"jsonapi": model.AccountAuth_JsonAPI,
}

func NewCreateCmd() *cobra.Command {
	var scopeFlag string

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new API key",
		Long: `Create a new API key for programmatic access to Anytype.

Scope controls what the key may call:
  limited   key-based HTTP auth only (no /v2 JSON API)
  jsonapi   /v1 and /v2 JSON API access`,
		Args: cmdutil.ExactArgs(1, "cannot create API key: name argument required"),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			scope, ok := scopeNames[strings.ToLower(scopeFlag)]
			if !ok {
				return output.Error("invalid scope %q: must be one of limited, jsonapi", scopeFlag)
			}

			resp, err := core.CreateAPIKey(name, scope)
			if err != nil {
				return output.Error("Failed to create API key: %w", err)
			}

			output.Success("API key created successfully")
			output.Info("Name: %s", name)
			output.Info("Scope: %s", strings.ToLower(scopeFlag))
			output.Info("Key: %s", resp.AppKey)
			fmt.Fprintln(cmd.OutOrStdout())

			return nil
		},
	}

	cmd.Flags().StringVar(&scopeFlag, "scope", "limited", "Key scope: limited or jsonapi")

	return cmd
}

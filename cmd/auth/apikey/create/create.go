package create

import (
	"github.com/spf13/cobra"
	"github.com/anyproto/anytype-heart/pkg/lib/pb/model"

	"github.com/anyproto/anytype-cli/cmd/cmdutil"
	"github.com/anyproto/anytype-cli/core"
	"github.com/anyproto/anytype-cli/core/output"
)

func NewCreateCmd() *cobra.Command {
	var (
		scopeFlag string
		spaces    string
		allSpaces bool
		perm      string
	)

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new API key",
		Long: `Create a new API key for programmatic access to Anytype.

Scope controls what the key may call:
  limited   key-based HTTP auth only (no /v2 JSON API)
  jsonapi   /v1 and /v2 JSON API access

A grant (JsonAPI keys only) narrows the key to specific spaces:
  --spaces id1,id2   only these spaces
  --all-spaces       every space, including ones created later
  --perm read|readwrite  permission inside the granted spaces
Without a grant a JsonAPI key can access every space. Grants can be
changed later with 'anytype auth apikey grant' — same key, no re-issue.`,
		Args: cmdutil.ExactArgs(1, "cannot create API key: name argument required"),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			scope, err := core.ParseScope(scopeFlag)
			if err != nil {
				return output.Error("%w", err)
			}

			flags := core.GrantFlags{Spaces: spaces, AllSpaces: allSpaces, Perm: perm}
			if flags.GrantRequiresScope() && scope != model.AccountAuth_JsonAPI {
				return output.Error("a grant requires --scope jsonapi (grants do not exist on limited keys)")
			}
			grant, err := flags.BuildGrant()
			if err != nil {
				return output.Error("%w", err)
			}

			resp, err := core.CreateAPIKey(name, scope, grant)
			if err != nil {
				return output.Error("Failed to create API key: %w", err)
			}

			output.Success("API key created successfully")
			output.Info("Name: %s", name)
			output.Info("Scope: %s", scopeFlag)
			if grant != nil {
				output.Info("Grant: %s", core.DescribeGrant(grant))
			}
			output.Info("Key: %s", resp.AppKey)

			return nil
		},
	}

	cmd.Flags().StringVar(&scopeFlag, "scope", "limited", "Key scope: limited or jsonapi")
	cmd.Flags().StringVar(&spaces, "spaces", "", "Comma-separated space ids the key may access")
	cmd.Flags().BoolVar(&allSpaces, "all-spaces", false, "Grant every space (including ones created later)")
	cmd.Flags().StringVar(&perm, "perm", "read", "Permission inside the granted spaces: read or readwrite")

	cmd.MarkFlagsMutuallyExclusive("spaces", "all-spaces")

	return cmd
}

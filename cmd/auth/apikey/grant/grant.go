package grant

import (
	"github.com/spf13/cobra"

	"github.com/anyproto/anytype-heart/pkg/lib/pb/model"

	"github.com/anyproto/anytype-cli/cmd/cmdutil"
	"github.com/anyproto/anytype-cli/core"
	"github.com/anyproto/anytype-cli/core/output"
)

func NewGrantCmd() *cobra.Command {
	var (
		spaces    string
		allSpaces bool
		perm      string
		clear     bool
	)

	cmd := &cobra.Command{
		Use:   "grant <id>",
		Short: "Set the space grant of an existing API key",
		Long: `Replace the space grant of an existing JsonAPI key, in place.

The key string never changes, so clients (MCP config, desktop approvals)
keep working with no redistribution; heart drops its HTTP session cache,
so the new grant applies from the very next request. Repeat calls simply
replace the grant; --clear removes it (back to every-space access).

Grants only exist on JsonAPI keys: a grant on a Limited key is refused.
The space ids are the full ids (bafyre...) as returned by 'anytype space list'.`,
		Args: cmdutil.ExactArgs(1, "cannot update API key grant: id argument required (see 'anytype auth apikey list')"),
		RunE: func(cmd *cobra.Command, args []string) error {
			appHash := args[0]

			var grant *model.AccountAuthAppGrant
			if clear {
				if spaces != "" || allSpaces {
					return output.Error("--clear cannot be combined with --spaces or --all-spaces")
				}
			} else {
				flags := core.GrantFlags{Spaces: spaces, AllSpaces: allSpaces, Perm: perm}
				if !flags.GrantRequiresScope() {
					return output.Error("nothing to set: pass --spaces <id,id,...>, --all-spaces, or --clear")
				}
				g, err := flags.BuildGrant()
				if err != nil {
					return output.Error("%w", err)
				}
				grant = g
			}

			if err := core.UpdateAPIKeyGrant(appHash, grant); err != nil {
				return output.Error("Failed to update API key grant: %w", err)
			}

			if clear {
				output.Success("Grant cleared for key %s (every-space access)", appHash)
			} else {
				output.Success("Grant updated for key %s: %s", appHash, core.DescribeGrant(grant))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&spaces, "spaces", "", "Comma-separated space ids the key may access")
	cmd.Flags().BoolVar(&allSpaces, "all-spaces", false, "Grant every space (including ones created later)")
	cmd.Flags().StringVar(&perm, "perm", "read", "Permission inside the granted spaces: read or readwrite")
	cmd.Flags().BoolVar(&clear, "clear", false, "Remove the grant (every-space access)")

	cmd.MarkFlagsMutuallyExclusive("spaces", "all-spaces", "clear")

	return cmd
}

package list

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/anyproto/anytype-cli/core"
	"github.com/anyproto/anytype-cli/core/output"
)

func NewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all API keys",
		Long:  "List all API keys associated with your account",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := core.ListAPIKeys()
			if err != nil {
				return output.Error("Failed to list API keys: %w", err)
			}

			if len(resp.App) == 0 {
				output.Info("No API keys found.")
				return nil
			}

			// Sort by creation date (newest first)
			sort.Slice(resp.App, func(i, j int) bool {
				return resp.App[i].CreatedAt > resp.App[j].CreatedAt
			})

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tID\tKEY\tSCOPE\tGRANT\tCREATED")
			fmt.Fprintln(w, "----\t--\t---\t-----\t-----\t----------")

			for _, app := range resp.App {
				createdAt := time.Unix(app.CreatedAt, 0).Format("2006-01-02 15:04:05")
				shortKey := app.AppKey
				if len(shortKey) > 8 {
					shortKey = shortKey[:8] + "..."
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", app.AppName, app.AppHash, shortKey, app.Scope.String(), core.DescribeGrant(app.Grant), createdAt)
			}

			w.Flush()
			return nil
		},
	}

	return cmd
}

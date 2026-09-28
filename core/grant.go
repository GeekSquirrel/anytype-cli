package core

import (
	"fmt"
	"strings"

	"github.com/anyproto/anytype-heart/pkg/lib/pb/model"
)

// Scope and permission names as heart prints them in errors and whoami.
var scopeNames = map[string]model.AccountAuthLocalApiScope{
	"limited": model.AccountAuth_Limited,
	"jsonapi": model.AccountAuth_JsonAPI,
}

var permNames = map[string]model.AccountAuthAppGrantPerm{
	"read":      model.AccountAuthAppGrant_Read,
	"readwrite": model.AccountAuthAppGrant_ReadWrite,
}

// ParseScope maps a --scope flag value to the heart enum.
func ParseScope(v string) (model.AccountAuthLocalApiScope, error) {
	scope, ok := scopeNames[strings.ToLower(v)]
	if !ok {
		return 0, fmt.Errorf("invalid scope %q: must be one of limited, jsonapi", v)
	}
	return scope, nil
}

// GrantFlags is the shared flag set of `apikey create` and `apikey grant`:
// either an explicit space list or --all-spaces, plus a permission level.
type GrantFlags struct {
	Spaces    string
	AllSpaces bool
	Perm      string
}

// BuildGrant converts the flags into a heart grant. A nil result means "no
// grant", which for a JsonAPI key is the unrestricted (legacy) form.
func (f GrantFlags) BuildGrant() (*model.AccountAuthAppGrant, error) {
	if f.Spaces == "" && !f.AllSpaces {
		return nil, nil
	}
	if f.Spaces != "" && f.AllSpaces {
		return nil, fmt.Errorf("--spaces and --all-spaces are mutually exclusive")
	}
	perm, ok := permNames[strings.ToLower(f.Perm)]
	if !ok {
		return nil, fmt.Errorf("invalid perm %q: must be one of read, readwrite", f.Perm)
	}
	grant := &model.AccountAuthAppGrant{Perm: perm}
	if f.AllSpaces {
		grant.AllSpaces = true
		return grant, nil
	}
	var ids []string
	for _, s := range strings.Split(f.Spaces, ",") {
		if s = strings.TrimSpace(s); s != "" {
			ids = append(ids, s)
		}
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("--spaces is empty after parsing; list comma-separated space ids or use --all-spaces")
	}
	grant.SpaceIds = ids
	return grant, nil
}

// GrantRequiresScope reports whether the flags express a grant, which heart
// only accepts on JsonAPI keys.
func (f GrantFlags) GrantRequiresScope() bool {
	return f.Spaces != "" || f.AllSpaces
}

// DescribeGrant renders a grant compactly for list output.
func DescribeGrant(grant *model.AccountAuthAppGrant) string {
	if grant == nil {
		return "-"
	}
	perm := strings.ToLower(grant.Perm.String())
	if grant.AllSpaces {
		return "all spaces (" + perm + ")"
	}
	return fmt.Sprintf("%d spaces (%s)", len(grant.SpaceIds), perm)
}

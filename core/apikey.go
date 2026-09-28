package core

import (
	"context"
	"fmt"
	"github.com/anyproto/anytype-heart/pkg/lib/pb/model"

	"github.com/anyproto/anytype-heart/pb"
	"github.com/anyproto/anytype-heart/pb/service"
)

// CreateAPIKey creates a new API key for local app access.
// scope may be AccountAuth_Limited or AccountAuth_JsonAPI; heart rejects
// AccountAuth_Full here (Full is reserved for account-key sessions).
// grant (JsonAPI keys only) narrows the key to a set of spaces; nil means
// unrestricted access to every space.
func CreateAPIKey(name string, scope model.AccountAuthLocalApiScope, grant *model.AccountAuthAppGrant) (*pb.RpcAccountLocalLinkCreateAppResponse, error) {
	var resp *pb.RpcAccountLocalLinkCreateAppResponse

	err := GRPCCall(func(ctx context.Context, client service.ClientCommandsClient) error {
		var err error
		resp, err = client.AccountLocalLinkCreateApp(ctx, &pb.RpcAccountLocalLinkCreateAppRequest{
			App: &model.AccountAuthAppInfo{
				AppName: name,
				Scope:   scope,
				Grant:   grant,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to create API key: %w", err)
		}

		if resp.Error != nil && resp.Error.Code != pb.RpcAccountLocalLinkCreateAppResponseError_NULL {
			return fmt.Errorf("API error: %s", resp.Error.Description)
		}

		return nil
	})

	return resp, err
}

// ListAPIKeys lists all API keys
func ListAPIKeys() (*pb.RpcAccountLocalLinkListAppsResponse, error) {
	var resp *pb.RpcAccountLocalLinkListAppsResponse

	err := GRPCCall(func(ctx context.Context, client service.ClientCommandsClient) error {
		var err error
		resp, err = client.AccountLocalLinkListApps(ctx, &pb.RpcAccountLocalLinkListAppsRequest{})
		if err != nil {
			return fmt.Errorf("failed to list API keys: %w", err)
		}

		if resp.Error != nil && resp.Error.Code != pb.RpcAccountLocalLinkListAppsResponseError_NULL {
			return fmt.Errorf("API error: %s", resp.Error.Description)
		}

		return nil
	})

	return resp, err
}

// RevokeAPIKey revokes an API key by appId
func RevokeAPIKey(appId string) error {
	return GRPCCall(func(ctx context.Context, client service.ClientCommandsClient) error {
		resp, err := client.AccountLocalLinkRevokeApp(ctx, &pb.RpcAccountLocalLinkRevokeAppRequest{
			AppHash: appId,
		})
		if err != nil {
			return fmt.Errorf("failed to revoke API key: %w", err)
		}

		if resp.Error != nil && resp.Error.Code != pb.RpcAccountLocalLinkRevokeAppResponseError_NULL {
			return fmt.Errorf("API error: %s", resp.Error.Description)
		}

		return nil
	})
}

// UpdateAPIKeyGrant replaces the space grant of an existing key in place.
// The key string never changes and heart evicts its HTTP session cache, so
// the new grant applies from the next request. A nil grant clears the
// scoping (back to every-space access).
func UpdateAPIKeyGrant(appHash string, grant *model.AccountAuthAppGrant) error {
	return GRPCCall(func(ctx context.Context, client service.ClientCommandsClient) error {
		resp, err := client.AccountLocalLinkUpdateApp(ctx, &pb.RpcAccountLocalLinkUpdateAppRequest{
			AppHash: appHash,
			Grant:   grant,
		})
		if err != nil {
			return fmt.Errorf("failed to update API key grant: %w", err)
		}

		if resp.Error != nil && resp.Error.Code != pb.RpcAccountLocalLinkUpdateAppResponseError_NULL {
			return fmt.Errorf("API error: %s", resp.Error.Description)
		}

		return nil
	})
}

package merchantpermission

import "context"

type MerchantPermission interface {
	ValidateMerchant(ctx context.Context, apiKey string) (map[string]interface{}, error)
}

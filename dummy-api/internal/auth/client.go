package auth

import (
	"context"
	"fmt"

	identitypb "github.com/vidwadeseram/go-boilerplate/identity-api/gen/grpc/identity/pb"
	"google.golang.org/grpc"
)

// Claims represent authenticated identity information shared by identity-api.
type Claims struct {
	UserID string
	Email  string
}

// Client validates tokens by delegating to identity-api over gRPC.
type Client struct {
	identity identitypb.IdentityClient
}

// NewClient constructs a Client using an existing gRPC connection.
func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{identity: identitypb.NewIdentityClient(conn)}
}

// Validate asks identity-api to validate the provided token string.
func (c *Client) Validate(ctx context.Context, token string) (*Claims, error) {
	resp, err := c.identity.ValidateToken(ctx, &identitypb.ValidateTokenRequest{Token: token})
	if err != nil {
		return nil, err
	}
	if !resp.GetValid() {
		if resp.Reason != nil {
			return nil, fmt.Errorf("%s", *resp.Reason)
		}
		return nil, fmt.Errorf("token validation failed")
	}
	return &Claims{UserID: resp.GetUserId(), Email: resp.GetEmail()}, nil
}

package gobotbsky

import (
	context "context"
	"errors"
	"fmt"
	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

// Connect is a helper method for Authenticate
func (c *BskyAgent) Connect(ctx context.Context) error {
	return c.Authenticate(ctx)
}

// Authenticate to the Personal Data Server setup in NewAgent
func (c *BskyAgent) Authenticate(ctx context.Context) error {

	// If no auth, create a new session
	if c.client.Auth == nil {
		c.logger.Debug("No auth. Creating new session.")
		err := c.createSession(ctx)
		if err != nil {
			return fmt.Errorf("error creating session. %w", err)
		}
		return nil
	}

	// Return if access token hasn't expired.
	aExp := TokenExpiration(c.client.Auth.AccessJwt)
	if aExp.After(time.Now()) {
		c.logger.Debug("Access token still valid.")
		return nil
	}

	// Refresh if refresh token has not expired, since access token is expired.
	rExp := TokenExpiration(c.client.Auth.RefreshJwt)
	if rExp.After(time.Now()) {
		c.logger.Debug("Access token expired. Refreshing.")
		err := c.refreshSession(ctx)
		if err != nil {
			return fmt.Errorf("error refreshing session. %w", err)
		}
		return nil
	}

	err := c.createSession(ctx)
	if err != nil {
		return fmt.Errorf("error creating session. %w", err)
	}
	c.logger.Debug(fmt.Sprintf("All tokens expired. Creating new session. Access Expired: %s, Refresh Expired: %s", aExp, rExp))
	return nil
}

func (c *BskyAgent) refreshSession(ctx context.Context) error {
	session, err := atproto.ServerRefreshSession(ctx, c.client)
	if err != nil {
		return fmt.Errorf("error refreshing session. %w", err)
	}
	c.updateClientAuth(session.AccessJwt, session.RefreshJwt, session.Handle, session.Did)
	c.lastRefresh = time.Now()
	return nil
}

func (c *BskyAgent) createSession(ctx context.Context) error {
	session, err := atproto.ServerCreateSession(ctx, c.client, &atproto.ServerCreateSession_Input{
		Identifier: c.handle, Password: c.apikey,
	})
	if err != nil {
		return fmt.Errorf("error creating new session. %w", err)
	}
	c.updateClientAuth(session.AccessJwt, session.RefreshJwt, session.Handle, session.Did)
	c.lastCreate = time.Now()
	return nil
}

func (c *BskyAgent) updateClientAuth(accessJwt, refreshJwt, handle, did string) {
	c.client.Auth = &xrpc.AuthInfo{
		AccessJwt:  accessJwt,
		RefreshJwt: refreshJwt,
		Handle:     handle,
		Did:        did,
	}
}

func TokenExpiration(token string) *jwt.NumericDate {
	t, _, err := jwt.NewParser().ParseUnverified(token, jwt.MapClaims{})
	if err != nil && !errors.Is(err, jwt.ErrTokenUnverifiable) {
		return &jwt.NumericDate{}
	}

	exp, err := t.Claims.GetExpirationTime()
	if err != nil {
		return &jwt.NumericDate{}
	}
	return exp
}

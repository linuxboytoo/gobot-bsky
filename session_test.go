package gobotbsky

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBskyAgent_Authenticate(t *testing.T) {
	tests := []struct {
		accessTokenExpiration  time.Time
		refreshTokenExpiration time.Time
		created                bool
		refreshed              bool
	}{
		{
			timeExpired(),
			timeNotExpired(),
			false,
			true,
		},
		{
			timeNotExpired(),
			timeNotExpired(),
			false,
			false,
		},
		{
			timeExpired(),
			timeExpired(),
			true,
			false,
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, test := range tests {
		pds := NewMockPDS()

		pds.SetAccessTokenExpiration(test.accessTokenExpiration)
		pds.SetRefreshTokenExpiration(test.refreshTokenExpiration)

		if err := pds.Start(); err != nil {
			t.Fatalf("Error starting Mock PDC. %v", err)
		}

		agent := NewAgent(ctx, pds.URL(), "test", "testkey")
		agent.logger = &SimpleLogger{}

		assert.Nil(t, agent.client.Auth)
		if err := agent.Connect(ctx); err != nil {
			t.Fatalf("Error on initial connect. %v", err)
		}
		assert.NotNil(t, agent.client.Auth)

		lastRefresh := agent.lastRefresh
		lastCreate := agent.lastCreate
		assert.NoError(t, agent.Authenticate(ctx))
		assert.Equal(t, test.refreshed, lastRefresh != agent.lastRefresh && lastCreate == agent.lastCreate)
		assert.Equal(t, test.created, lastRefresh == agent.lastRefresh && lastCreate != agent.lastCreate)
	}

}

func timeExpired() time.Time {
	return time.Now().Add(-1 * time.Hour)
}

func timeNotExpired() time.Time {
	return time.Now().Add(1 * time.Hour)
}

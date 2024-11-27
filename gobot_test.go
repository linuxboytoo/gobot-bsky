package gobotbsky

import (
	"context"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBskyAgent_PostToFeed(t *testing.T) {
	ctx := context.Background()
	pds := NewMockPDS()
	if err := pds.Start(); err != nil {
		t.Fatalf("Error starting Mock PDS. %v", err)
	}

	tests := []struct {
		accessTokenExpiration  time.Time
		refreshTokenExpiration time.Time
		created                bool
		refreshed              bool
	}{
		{timeNotExpired(), timeNotExpired(), false, false},
	}

	for _, test := range tests {
		a := NewAgent(ctx, pds.URL(), "testhandle", "testkey")

		pds.SetAccessTokenExpiration(test.accessTokenExpiration)
		pds.SetRefreshTokenExpiration(test.refreshTokenExpiration)

		_, _, err := a.PostToFeed(ctx, bsky.FeedPost{})
		if err != nil {
			t.Fatalf("Error Posting to Feed. %v", err)
		}

		assert.Equal(t, pds.authCount, 1)
	}
}

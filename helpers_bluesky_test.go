package gobotbsky

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"net/http/httptest"
	"time"
)

type MockPDS struct {
	server             *httptest.Server
	signingKey         *ecdsa.PrivateKey
	accessTokenExpire  time.Time
	refreshTokenExpire time.Time
}

func NewMockPDS() *MockPDS {
	signingKey, err := generateSigningKey()
	if err != nil {
		fmt.Printf("Error generating signing key. %v\n", err)
	}
	return &MockPDS{
		signingKey:         signingKey,
		accessTokenExpire:  time.Now().Add(1 * time.Minute),
		refreshTokenExpire: time.Now().Add(5 * time.Minute),
	}
}

func (m *MockPDS) SetAccessTokenExpiration(exp time.Time) {
	m.accessTokenExpire = exp
}

func (m *MockPDS) SetRefreshTokenExpiration(exp time.Time) {
	m.refreshTokenExpire = exp
}

func (m *MockPDS) Start() error {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		o, err := generateSession(m.accessTokenExpire, m.refreshTokenExpire, m.signingKey)
		if err != nil {
			fmt.Printf("error generating session. %v\n", err)
			w.WriteHeader(500)
		}
		j, err := json.Marshal(o)
		if err != nil {
			w.WriteHeader(500)
		}
		_, err = w.Write(j)
		if err != nil {
			w.WriteHeader(500)
		}
	}))
	m.server = server
	return nil
}
func (m *MockPDS) Stop() {
	m.server.Close()
}
func (m *MockPDS) URL() string {
	return m.server.URL
}

func generateSession(accessTokenExp time.Time, refreshTokenExp time.Time, signingKey *ecdsa.PrivateKey) (atproto.ServerCreateSession_Output, error) {
	accessToken, err := generateSignedToken(accessTokenExp, signingKey)
	if err != nil {
		return atproto.ServerCreateSession_Output{}, fmt.Errorf("error generating access token. %w", err)
	}
	refreshToken, err := generateSignedToken(refreshTokenExp, signingKey)
	if err != nil {
		return atproto.ServerCreateSession_Output{}, fmt.Errorf("error generating access token. %w", err)
	}
	return atproto.ServerCreateSession_Output{
		AccessJwt:  accessToken,
		RefreshJwt: refreshToken,
	}, nil
}
func generateSigningKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}
func generateSignedToken(expiration time.Time, key *ecdsa.PrivateKey) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{"exp": expiration.Unix()})

	s, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("error getting signed string. %w", err)
	}
	return s, nil
}

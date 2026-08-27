package test

import (
	"context"
	"strings"
	"testing"
	"time"

	jwtpkg "fullstack-app/server/pkg/jwt"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() {
		_ = rdb.Close()
		mr.Close()
	})
	return mr, rdb
}

func TestJWTGenerateAndParseTokenPair(t *testing.T) {
	_, rdb := newTestRedis(t)
	manager := jwtpkg.NewManager(rdb, 2, 24)
	loginAt := time.Now().Add(-10 * time.Minute).Unix()

	pair, err := manager.GenerateTokenPair(42, loginAt)
	if err != nil {
		t.Fatalf("GenerateTokenPair() error = %v", err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" || pair.AccessToken == pair.RefreshToken {
		t.Fatalf("unexpected token pair: %+v", pair)
	}
	if len(pair.AccessToken) != 43 || len(pair.RefreshToken) != 43 {
		t.Fatalf("token length = (%d, %d), want 43-character opaque tokens", len(pair.AccessToken), len(pair.RefreshToken))
	}
	if strings.Contains(pair.AccessToken, "42") || strings.Contains(pair.RefreshToken, "42") {
		t.Fatalf("tokens should not expose user ID: %+v", pair)
	}
	if delta := pair.ExpiresAt - time.Now().Add(2*time.Hour).Unix(); delta < -1 || delta > 1 {
		t.Fatalf("ExpiresAt delta = %d seconds, want approximately zero", delta)
	}

	accessClaims, err := manager.ParseAccessToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("ParseAccessToken() error = %v", err)
	}
	if accessClaims.UserID != 42 || accessClaims.LoginAt != loginAt || accessClaims.TokenType != "access" || accessClaims.ExpiresAt != pair.ExpiresAt {
		t.Fatalf("unexpected access claims: %+v", accessClaims)
	}

	refreshClaims, err := manager.ParseRefreshToken(pair.RefreshToken)
	if err != nil {
		t.Fatalf("ParseRefreshToken() error = %v", err)
	}
	if refreshClaims.UserID != 42 || refreshClaims.LoginAt != loginAt || refreshClaims.TokenType != "refresh" {
		t.Fatalf("unexpected refresh claims: %+v", refreshClaims)
	}
	if refreshClaims.ExpiresAt <= accessClaims.ExpiresAt {
		t.Fatalf("refresh expiration %d should be after access expiration %d", refreshClaims.ExpiresAt, accessClaims.ExpiresAt)
	}

	ctx := context.Background()
	if got := rdb.Exists(ctx, "session:access:"+pair.AccessToken).Val(); got != 1 {
		t.Fatalf("access session exists = %d, want 1", got)
	}
	if got := rdb.Exists(ctx, "session:refresh:"+pair.RefreshToken).Val(); got != 1 {
		t.Fatalf("refresh session exists = %d, want 1", got)
	}
}

func TestJWTParseRejectsInvalidExpiredAndMalformedTokens(t *testing.T) {
	_, rdb := newTestRedis(t)
	manager := jwtpkg.NewManager(rdb, -1, 1)
	pair, err := manager.GenerateTokenPair(9, 123)
	if err != nil {
		t.Fatalf("GenerateTokenPair() error = %v", err)
	}

	if _, err := manager.ParseAccessToken(pair.AccessToken); err == nil || err.Error() != "token expired" {
		t.Fatalf("ParseAccessToken() error = %v, want token expired", err)
	}
	if _, err := manager.ParseAccessToken("missing"); err == nil || err.Error() != "invalid token" {
		t.Fatalf("missing access token error = %v, want invalid token", err)
	}
	if _, err := manager.ParseRefreshToken(pair.AccessToken); err == nil || err.Error() != "invalid token" {
		t.Fatalf("wrong token type error = %v, want invalid token", err)
	}

	if err := rdb.Set(context.Background(), "session:access:malformed", "not-json", time.Hour).Err(); err != nil {
		t.Fatalf("seed malformed session: %v", err)
	}
	if _, err := manager.ParseAccessToken("malformed"); err == nil || err.Error() != "invalid token" {
		t.Fatalf("malformed access token error = %v, want invalid token", err)
	}
}

func TestJWTRevoke(t *testing.T) {
	_, rdb := newTestRedis(t)
	manager := jwtpkg.NewManager(rdb, 1, 2)
	pair, err := manager.GenerateTokenPair(11, time.Now().Unix())
	if err != nil {
		t.Fatalf("GenerateTokenPair() error = %v", err)
	}

	if err := manager.Revoke(""); err != nil {
		t.Fatalf("Revoke(empty) error = %v, want nil", err)
	}
	if err := manager.Revoke(pair.AccessToken); err != nil {
		t.Fatalf("Revoke() error = %v", err)
	}
	if got := rdb.Exists(context.Background(), "session:access:"+pair.AccessToken).Val(); got != 0 {
		t.Fatalf("revoked session exists = %d, want 0", got)
	}
	if _, err := manager.ParseAccessToken(pair.AccessToken); err == nil || err.Error() != "invalid token" {
		t.Fatalf("revoked token error = %v, want invalid token", err)
	}
}

func TestJWTMaybeRenewAccessToken(t *testing.T) {
	_, rdb := newTestRedis(t)
	manager := jwtpkg.NewManager(rdb, 1, 2)
	now := time.Now()

	cases := []struct {
		name      string
		claims    *jwtpkg.Claims
		wantRenew bool
	}{
		{name: "nil claims", claims: nil, wantRenew: false},
		{name: "refresh token", claims: &jwtpkg.Claims{UserID: 1, TokenType: "refresh", ExpiresAt: now.Add(time.Minute).Unix()}, wantRenew: false},
		{name: "not near expiration", claims: &jwtpkg.Claims{UserID: 2, LoginAt: 100, TokenType: "access", ExpiresAt: now.Add(2 * time.Hour).Unix()}, wantRenew: false},
		{name: "near expiration", claims: &jwtpkg.Claims{UserID: 3, LoginAt: 200, TokenType: "access", ExpiresAt: now.Add(10 * time.Minute).Unix()}, wantRenew: true},
		{name: "already expired", claims: &jwtpkg.Claims{UserID: 4, TokenType: "access", ExpiresAt: now.Add(-time.Minute).Unix()}, wantRenew: false},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			newToken, err := manager.MaybeRenewAccessToken(tt.claims)
			if err != nil {
				t.Fatalf("MaybeRenewAccessToken() error = %v", err)
			}
			if (newToken != "") != tt.wantRenew {
				t.Fatalf("new token present = %t, want %t", newToken != "", tt.wantRenew)
			}
			if !tt.wantRenew {
				return
			}
			newClaims, err := manager.ParseAccessToken(newToken)
			if err != nil {
				t.Fatalf("ParseAccessToken(new token) error = %v", err)
			}
			if newClaims.UserID != tt.claims.UserID || newClaims.LoginAt != tt.claims.LoginAt || newClaims.TokenType != "access" {
				t.Fatalf("unexpected renewed claims: %+v", newClaims)
			}
		})
	}
}

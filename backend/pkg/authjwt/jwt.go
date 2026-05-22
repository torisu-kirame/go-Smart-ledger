package authjwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

const (
	ClaimTypeAccess  = "access"
	ClaimTypeRefresh = "refresh"
)

type Claims struct {
	UserID   string `json:"uid"`
	Username string `json:"usr"`
	Type     string `json:"typ"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

type Config struct {
	AccessSecret  string
	RefreshSecret string
	AccessExpire  int64 // seconds
	RefreshExpire int64 // seconds
}

func Issue(cfg Config, userID, username string) (*TokenPair, error) {
	now := time.Now()
	accessExp := now.Add(time.Duration(cfg.AccessExpire) * time.Second)
	refreshExp := now.Add(time.Duration(cfg.RefreshExpire) * time.Second)

	access, err := sign(cfg.AccessSecret, Claims{
		UserID:   userID,
		Username: username,
		Type:     ClaimTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExp),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   userID,
		},
	})
	if err != nil {
		return nil, err
	}
	refresh, err := sign(cfg.RefreshSecret, Claims{
		UserID:   userID,
		Username: username,
		Type:     ClaimTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExp),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   userID,
		},
	})
	if err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    cfg.AccessExpire,
	}, nil
}

func RefreshAccess(cfg Config, refreshToken string) (*TokenPair, error) {
	claims, err := Parse(cfg.RefreshSecret, refreshToken, ClaimTypeRefresh)
	if err != nil {
		return nil, err
	}
	return Issue(cfg, claims.UserID, claims.Username)
}

func Parse(secret, tokenStr, expectType string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid || claims.Type != expectType {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func sign(secret string, claims Claims) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

package apiserver

import (
	"fmt"
	"github.com/cap79/GopherSocial/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"strconv"
	"time"
)

var signingMethod = jwt.SigningMethodHS256

type JwtManager struct {
	cfg *config.Config
}

func NewJwtManager(cfg *config.Config) *JwtManager {
	return &JwtManager{cfg: cfg}
}

type TokenPair struct {
	AccessToken  *jwt.Token
	RefreshToken *jwt.Token
}

type CustomClaims struct {
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func (j *JwtManager) Parse(token string) (*jwt.Token, error) {
	parser := jwt.NewParser()
	jwtToken, err := parser.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if t.Method != signingMethod {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(j.cfg.JwtSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}
	return jwtToken, nil
}

func (j *JwtManager) IsAccessToken(token *jwt.Token) bool {
	jwtClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}
	if tokenType, ok := jwtClaims["token_type"]; ok {
		return tokenType == "access"
	}
	return false
}

func (j *JwtManager) GenerateTokenPair(userId int) (*TokenPair, error) {
	var err error
	jwtAccessToken := jwt.NewWithClaims(signingMethod, CustomClaims{
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.cfg.ApiServerHost + ":" + j.cfg.ApiServerAddr,
			Subject:   strconv.Itoa(userId),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)),
		},
	})
	secret := []byte(j.cfg.JwtSecret)

	jwtAccessToken.Raw, err = jwtAccessToken.SignedString(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token %w", err)
	}

	jwtRefreshToken := jwt.NewWithClaims(signingMethod, CustomClaims{
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.cfg.ApiServerHost + ":" + j.cfg.ApiServerAddr,
			Subject:   strconv.Itoa(userId),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 30)),
		},
	})
	jwtRefreshToken.Raw, err = jwtAccessToken.SignedString(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token %w", err)
	}

	return &TokenPair{
		AccessToken:  jwtAccessToken,
		RefreshToken: jwtRefreshToken,
	}, nil
}

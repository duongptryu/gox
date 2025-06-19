package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/duongptryu/gox/syserr"
)

// JWTService implements JWT token operations
type JWTService struct {
	secretKey          []byte
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey string, accessTokenExpiry, refreshTokenExpiry time.Duration) *JWTService {
	return &JWTService{
		secretKey:          []byte(secretKey),
		accessTokenExpiry:  accessTokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
	}
}

// Claims represents JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	UserType string `json:"user_type"`
	Type     string `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// GenerateTokenPair generates access and refresh tokens
func (s *JWTService) GenerateTokenPair(ctx context.Context, userID string, userType string) (accessToken, refreshToken string, expiresIn int64, err error) {
	// Generate access token
	accessClaims := Claims{
		UserID:   userID,
		UserType: userType,
		Type:     "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}

	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessTokenObj.SignedString(s.secretKey)
	if err != nil {
		return "", "", 0, syserr.Wrap(err, syserr.InternalCode, "failed to generate access token")
	}

	// Generate refresh token
	refreshClaims := Claims{
		UserID:   userID,
		UserType: userType,
		Type:     "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.refreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}

	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshTokenObj.SignedString(s.secretKey)
	if err != nil {
		return "", "", 0, syserr.Wrap(err, syserr.InternalCode, "failed to generate refresh token")
	}

	return accessToken, refreshToken, int64(s.accessTokenExpiry.Seconds()), nil
}

// ValidateToken validates a JWT token and returns claims
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, syserr.Wrap(err, syserr.UnauthorizedCode, "invalid token")
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, syserr.New(syserr.UnauthorizedCode, "invalid token claims")
}

// ValidateAccessToken validates specifically an access token
func (s *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != "access" {
		return nil, syserr.New(syserr.UnauthorizedCode, "token is not an access token")
	}

	return claims, nil
}

// ValidateRefreshToken validates specifically a refresh token
func (s *JWTService) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != "refresh" {
		return nil, syserr.New(syserr.UnauthorizedCode, "token is not a refresh token")
	}

	return claims, nil
}

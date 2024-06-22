package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

// Load environment variables when the package is initialized
func init() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	accessTokenKey = []byte(os.Getenv("ACCESS_TOKEN_KEY"))
	refreshTokenKey = []byte(os.Getenv("REFRESH_TOKEN_KEY"))
}

var (
	accessTokenKey  []byte
	refreshTokenKey []byte
)

type Claims struct {
	Content string `json:"content"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(content string) (string, error) {
	claims := &Claims{
		Content: content,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(accessTokenKey)
}

func GenerateRefreshToken(content string) (string, error) {
	claims := &Claims{
		Content: content,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 168)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(refreshTokenKey)
}

func ValidateAccessToken(token string) (bool, error) {
	_, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return accessTokenKey, nil
	})

	if err != nil {
		return false, err
	}
	return true, nil
}

func ValidateRefreshToken(token string) (bool, error) {
	_, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return refreshTokenKey, nil
	})

	if err != nil {
		return false, err
	}
	return true, nil
}

func DecodeAccessToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return accessTokenKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

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
	resetTokenKey = []byte(os.Getenv("RESET_TOKEN_KEY"))
}

var (
	accessTokenKey  []byte
	refreshTokenKey []byte
	resetTokenKey   []byte
)

type Claims struct {
	Content uint `json:"content"`
	jwt.RegisteredClaims
}

type ForgotClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(content uint) (string, error) {
	claims := &Claims{
		Content: content,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 1)), // Token expires in 1 hour
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(accessTokenKey)
}

func GenerateRefreshToken(content uint) (string, error) {
	claims := &Claims{
		Content: content,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 336)), // Token expires in 14 days
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(refreshTokenKey)
}

func GenerateResetToken(email string) (string, error) {
	claims := &ForgotClaims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 1)), // Token expires in 1 hour
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(resetTokenKey)
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

func VerifyResetToken(tokenStr string) (*ForgotClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &ForgotClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return resetTokenKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*ForgotClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
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

func DecodeRefreshToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return refreshTokenKey, nil
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

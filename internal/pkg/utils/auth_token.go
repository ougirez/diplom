package utils

import (
	"errors"
	"fmt"
	"github.com/ougirez/diplom/internal/pkg/constants"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/spf13/viper"
)

type AuthTokenWrapper struct {
	UserID int64  `json:"user_id"`
	Secret string `json:"secret"`
	jwt.StandardClaims
}

func GenerateAuthToken(atw *AuthTokenWrapper) (string, error) {
	if atw.ExpiresAt == 0 {
		t := time.Second * time.Duration(viper.GetInt64(constants.ViperJWTTTLKey))
		atw.ExpiresAt = time.Now().Add(t).Unix()
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, atw)
	authToken, err := jwtToken.SignedString([]byte(viper.GetString(constants.ViperJWTSecretKey)))
	if err != nil {
		return "", fmt.Errorf("%w: %v", constants.ErrSignToken, err)
	}

	return authToken, nil
}

func ParseAuthToken(authToken string) (*AuthTokenWrapper, error) {
	t, err := jwt.ParseWithClaims(
		authToken,
		&AuthTokenWrapper{},
		keyFunc([]byte(viper.GetString(constants.ViperJWTSecretKey))),
	)

	if ve, ok := err.(*jwt.ValidationError); ok {
		// check if Expiration error was set
		if ve.Errors&jwt.ValidationErrorExpired == jwt.ValidationErrorExpired {
			return nil, constants.ErrAuthTokenExpired
		} else {
			return nil, constants.ErrAuthTokenInvalid
		}
	} else if err != nil {
		return nil, fmt.Errorf("%w: %v", constants.ErrParseAuthToken, err)
	}

	atw, ok := t.Claims.(*AuthTokenWrapper)
	if !ok {
		return nil, fmt.Errorf("%w: %v", constants.ErrAuthTokenInvalid, errors.New("failed casting claims"))
	}

	return atw, nil
}

// KeyFunc returns key function for validating a token
func keyFunc(key []byte) func(token *jwt.Token) (interface{}, error) {
	return func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, constants.ErrUnexpectedSigningMethod
		}
		return key, nil
	}
}

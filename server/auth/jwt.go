package auth

import (
	"errors"
	"fmt"
	"github.com/pascaldekloe/jwt"
	"github.com/warneki/spaced/server/config"
	"time"
)

func GenerateJWT(username string, client string) jwt.Claims {
	var claims jwt.Claims
	claims.Subject = username
	claims.Issuer = config.JwtIssuer
	claims.Issued = jwt.NewNumericTime(time.Now())
	claims.Expires = jwt.NewNumericTime(time.Now().AddDate(0, 0, 1))
	claims.Set = map[string]interface{}{
		"client": client,
	}
	return claims
}

func SignAndSerializeJWT(claims jwt.Claims) (string, error) {
	token, err := claims.HMACSign("HS256", []byte(config.Key))
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return string(token), nil
}


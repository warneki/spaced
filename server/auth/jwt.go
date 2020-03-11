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

func VerifyJwt(token string) (jwt.Claims, error) {
	claims, err := jwt.HMACCheck([]byte(token), []byte(config.Key))
	if err != nil {
		return jwt.Claims{}, err
	}
	notExpired := claims.Valid(time.Now())
	correctIssuer := claims.Issuer == config.JwtIssuer
	correctClient := claims.Set["client"] == "web"

	if notExpired && correctIssuer && correctClient {
		return *claims, nil
	}
	errMsg := fmt.Sprintf("Expired: %t; Bad issuer :%t; Bad client :%t", !notExpired, !correctIssuer, !correctClient)
	fmt.Println(errMsg + " for claims " + string(claims.Raw))
	return *claims, errors.New(errMsg)
}

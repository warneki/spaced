package auth

import (
	"errors"
	"fmt"
	"github.com/pascaldekloe/jwt"
	"github.com/warneki/spaced/server/config"
	"strconv"
	"time"
)

func GenerateJWT(username string, client string) (jwt.Claims, string) {
	var claims jwt.Claims
	claims.Subject = username
	claims.Issuer = config.JwtIssuer
	now := time.Now().Round(time.Second)
	claims.Issued = jwt.NewNumericTime(now)
	claims.Expires = jwt.NewNumericTime(now.AddDate(0, 0, 1))

	// clientName identifies the client from which user connects plus issued and expire dates
	clientName := client + "_" + strconv.FormatInt(claims.Issued.Time().Unix(),
		10) + "_" + strconv.FormatInt(claims.Expires.Time().Unix(), 10)
	claims.Set = map[string]interface{}{
		"client": clientName,
	}
	return claims, clientName
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

	if notExpired && correctIssuer {
		return *claims, nil
	}
	errMsg := fmt.Sprintf("Expired: %t; Bad issuer :%t", !notExpired, !correctIssuer)
	fmt.Println(errMsg + " for claims " + string(claims.Raw))
	return *claims, errors.New(errMsg)
}

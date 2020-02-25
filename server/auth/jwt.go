package auth

import (
	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/SermoDigital/jose/jwt"
	"github.com/warneki/spaced/server/config"
	"time"
)

func GenerateJWT(username string, aud []string) jwt.JWT {
	expires := time.Now().Add(time.Duration(24) * time.Hour)
	claims := jws.Claims{
		"exp": expires,
		"iat": time.Now(),
		"sub": username,
		"aud": aud,
		"iss": config.JWTIssuer,
	}

	jwt := jws.NewJWT(claims, crypto.SigningMethodRS256)
	return jwt
}

func SignAndSerializeJWT(j jwt.JWT) string {
	rsaPrivate, _ := crypto.ParseRSAPrivateKeyFromPEM([]byte(config.PrKey))
	b, _ := j.Serialize(rsaPrivate)

	return string(b)
}

package internal

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/hgyowan/go-pkg-library/envs"
)

var JwtKey = []byte(envs.JwtSecret)

type JWTCustomClaims struct {
	UserID   uint
	LangCode string
	jwt.StandardClaims
}

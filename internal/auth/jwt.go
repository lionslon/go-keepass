package auth

import (
	"fmt"
	"github.com/lionslon/go-keepass/internal/server/config"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

type Authorizator struct {
	cfg *config.Config
}

var jwtAuth Authorizator

func Initialize(cfg *config.Config) {
	jwtAuth = Authorizator{
		cfg: cfg,
	}
}

func CreateToken(id string) (string, error) {

	if id == `` {
		return ``, fmt.Errorf("invalid id")
	}

	// Заполняем payload, для наших потребностей хватит jwt.StandardClaims
	expirationTime := time.Now().Add(jwtAuth.cfg.JWTDuration)
	claims := jwt.StandardClaims{
		Id:        id,
		ExpiresAt: expirationTime.Unix(),
	}

	// Непосредственно вычисляем токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtAuth.cfg.JWTKey)
	if err != nil {
		return ``, fmt.Errorf("cannot sign token claims: %w", err)
	}

	//Добавим Bearer, вернем
	return strings.Join([]string{"Bearer", tokenString}, ` `), nil
}

func verifyToken(token string) (string, error) {

	var claims jwt.StandardClaims

	_, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return jwtAuth.cfg.JWTKey, nil
	})
	if err != nil {
		return ``, fmt.Errorf("invalid jwt: %s", err)
	}

	return claims.Id, nil
}

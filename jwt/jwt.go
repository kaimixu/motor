package jwt

import (
	"errors"
	"fmt"

	"github.com/dgrijalva/jwt-go"
)

type JWT struct {
	secret string
}

type MotorClaims struct {
	jwt.StandardClaims
	Data map[string]interface{}
}

func NewJWT(secret string) *JWT {
	return &JWT{secret}
}

// 生成token
func (j *JWT) GenToken(claims *MotorClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secret))
}

// 验证token
func (j *JWT) ParseToken(tokenStr string) (*MotorClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &MotorClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(j.secret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*MotorClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的token")
}

// 判断token是否过期
func (j *JWT) IsExpires(err error) bool {
	if ve, ok := err.(*jwt.ValidationError); ok {
		if ve.Errors&jwt.ValidationErrorExpired != 0 {
			return true
		}
	}

	return false
}

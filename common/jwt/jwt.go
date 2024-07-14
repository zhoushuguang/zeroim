package jwt

import (
	"github.com/dgrijalva/jwt-go"
)

type Reg struct {
	secret string
}

func NewReg(secret string) *Reg {
	return &Reg{
		secret: secret,
	}
}

func (r *Reg) getValidationKey(*jwt.Token) (interface{}, error) {
	return []byte(r.secret), nil
}

func (r *Reg) VerifyToken(userId string, input string) error {
	token, err := jwt.Parse(input, r.getValidationKey)
	if err != nil {
		return err
	}
	if jwt.SigningMethodHS256.Alg() != token.Header["alg"] {
		return jwt.ErrSignatureInvalid
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		user, ok := claims["usr"]
		if !ok {
			return jwt.ErrSignatureInvalid
		}
		if userStr, ok := user.(string); !ok || userStr != userId {
			return jwt.ErrSignatureInvalid
		}
	}

	return nil
}

package tokens

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Token OAUTH/OIDC Token
type Token struct {
	Alg    string                 `json:"alg,omitempty" yaml:"alg,omitempty"`
	Kid    string                 `json:"kid,omitempty" yaml:"kid,omitempty"`
	Iss    string                 `json:"iss,omitempty" yaml:"iss,omitempty"`
	Exp    int64                  `json:"exp,omitempty" yaml:"exp,omitempty"`
	Aud    string                 `json:"aud,omitempty" yaml:"aud,omitempty"`
	Claims map[string]interface{} `json:"claims,omitempty" yaml:"claims,omitempty"`
}

// TokenFromBase64 ...
func TokenFromBase64(tokenString string) (*Token, error) {

	token := &Token{}
	_, err := jwt.Parse(tokenString, func(jwtToken *jwt.Token) (interface{}, error) {

		claims, ok := jwtToken.Claims.(jwt.MapClaims)
		if !ok {
			return nil, fmt.Errorf("Token claims have unexpected format")
		}

		for k, v := range claims {

			if k == "iss" {
				token.Iss, _ = v.(string)
			}

			if k == "exp" {
				floatValue := v.(float64)
				token.Exp = int64(floatValue)
			}

			if k == "aud" {
				token.Aud, _ = v.(string)
			}

		}

		token.Claims = claims
		token.Kid = jwtToken.Header["kid"].(string)
		token.Alg = jwtToken.Header["alg"].(string)

		if token.Exp == 0 {
			return nil, fmt.Errorf("Expiration(exp) not found")
		}

		return nil, nil

	})

	if err != nil {
		if err.Error() == "Token is expired" {
			return token, nil
		}
		if err.Error() == "Token used before issued" {
			return token, nil
		}

		if err.Error() == "key is of invalid type" {
			return token, nil
		}

		return nil, err
	}

	return token, nil

}

// Valid Returns true if entity is valid
func (t *Token) Valid() bool {
	if t.Exp > 0 && t.Exp > time.Now().Unix() {
		return true
	}
	return false
}

package token

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Main ...
func Main() {

	// h := hmac.New(sha256.New, []byte("H0mv9WDGUE9XrygQ8uBAyzPzksGibVsdu6xznj6UGGIMD4iavLkGdVBuPQ"))
	// hmacSampleSecret := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"foo": "bar",
		"nbf": time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte("secureSecretText"))

	fmt.Println(tokenString, err)

}

func main() {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	claims := &jwt.StandardClaims{
		ExpiresAt: 15000,
		Issuer:    "test",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)

	tokenString, err := token.SignedString(key)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(tokenString)
}

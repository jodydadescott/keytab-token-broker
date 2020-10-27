/*
Copyright © 2020 Jody Scott <jody@thescottsweb.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package token

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/jodydadescott/tokens2secrets/internal/publickey"
)

func Test1(t *testing.T) {
	err := runTest1()
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func runTest1() error {

	now := time.Now().Unix()

	// We generate a couple private keys and store their public key counterparts in a new Public Key Cache.
	// Then we create a new Token cache and provide it with the just created PublicKey cache. This allows
	// us to provide the matching public key for testing without the need to query external web servers.

	privateKeyA, publicKeyA, err := generateKeypair("https://issuer-a", "x", now+3600)
	if err != nil {
		return err
	}

	privateKeyB, publicKeyB, err := generateKeypair("https://issuer-b", "x", now+3600)
	if err != nil {
		return err
	}

	junkKey, _, err := generateKeypair("https://issuer-a", "x", now+3600)
	if err != nil {
		return err
	}

	testKeyCache := publickey.Dummy()

	testKeyCache.PutKey(publicKeyA)
	testKeyCache.PutKey(publicKeyB)

	config := &Config{}
	tokenCache, err := config.Build(testKeyCache)
	if err != nil {
		return err
	}

	validTokenSignedByIssuerA, err := newToken("https://issuer-a", "x", now+600, privateKeyA)
	if err != nil {
		return err
	}

	validTokenSignedByIssuerB, err := newToken("https://issuer-b", "x", now+600, privateKeyB)
	if err != nil {
		return err
	}

	expiredTokenSignedByIssuerA, err := newToken("https://issuer-a", "x", now-600, privateKeyA)
	if err != nil {
		return err
	}

	expiredTokenSignedByIssuerB, err := newToken("https://issuer-b", "x", now-600, privateKeyB)
	if err != nil {
		return err
	}

	invalidTokenClaimingIssuerA, err := newToken("https://issuer-a", "x", now+600, junkKey)
	if err != nil {
		return err
	}

	tokenFromMissingIssuer, err := newToken("https://does-not-exist", "x", now+600, junkKey)
	if err != nil {
		return err
	}

	// Err NOT expected
	_, err = tokenCache.ParseToken(validTokenSignedByIssuerA)
	if err != nil {
		return err
	}

	// Err NOT expected
	_, err = tokenCache.ParseToken(validTokenSignedByIssuerB)
	if err != nil {
		return err
	}

	// Err expected
	_, err = tokenCache.ParseToken(expiredTokenSignedByIssuerA)
	if err == nil {
		return fmt.Errorf("Expected ErrExpired, got nil")
	}
	if err != ErrExpired {
		return fmt.Errorf(fmt.Sprintf("Expected ErrExpired, got %s", err.Error()))
	}

	// Err expected
	_, err = tokenCache.ParseToken(expiredTokenSignedByIssuerB)
	if err == nil {
		return fmt.Errorf("Expected ErrExpired, got nil")
	}
	if err != ErrExpired {
		return fmt.Errorf(fmt.Sprintf("Expected ErrExpired, got %s", err.Error()))
	}

	// Err expected
	_, err = tokenCache.ParseToken(invalidTokenClaimingIssuerA)
	if err == nil {
		return fmt.Errorf("Expected ErrSignatureInvalid, got nil")
	}
	if err != ErrSignatureInvalid {
		return fmt.Errorf(fmt.Sprintf("Expected ErrSignatureInvalid, got %s", err.Error()))
	}

	_, err = tokenCache.ParseToken(tokenFromMissingIssuer)
	if err == nil {
		return fmt.Errorf("Expected ErrNotFound, got nil")
	}
	if err != ErrSignatureInvalid {
		return fmt.Errorf(fmt.Sprintf("Expected ErrNotFound, got %s", err.Error()))
	}

	// 	time.Sleep(6 * time.Second)

	return nil
}

func newToken(iss, kid string, exp int64, key *ecdsa.PrivateKey) (string, error) {

	claims := &jwt.StandardClaims{
		ExpiresAt: exp,
		Issuer:    iss,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = kid
	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func generateKeypair(iss, kid string, exp int64) (*ecdsa.PrivateKey, *publickey.PublicKey, error) {

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	if err != nil {
		return nil, nil, err
	}

	publicKey := &publickey.PublicKey{
		EcdsaPublicKey: privateKey.Public().(*ecdsa.PublicKey),
		Iss:            iss,
		Kid:            kid,
		Kty:            "EC",
		Exp:            exp,
	}

	return privateKey, publicKey, nil
}

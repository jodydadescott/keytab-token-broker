package tokenstore

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"kbridge/internal/cachemap"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.uber.org/zap"
)

const (
	defaultTokenCleanup    int = 30
	defaultPublicKeyCleaup int = 3000
	maxIdleConnections     int = 20
	defaultRequestTimeout  int = 5
)

// Config ...
type Config struct {
	TokenCleanup    int
	PublicKeyCleaup int
	RequestTimeout  int
}

// NewConfig ...
func NewConfig() *Config {
	return &Config{}
}

// TokenStore ...
type TokenStore struct {
	tokenCacheMap     *cachemap.CacheMap
	requestTimeout    int
	publicKeyCacheMap *cachemap.CacheMap
	httpClient        *http.Client
}

// PublicKey ...
type PublicKey struct {
	EcdsaPublicKey *ecdsa.PublicKey `json:"-"`
	Created        int64            `json:"created,omitempty"`
	Kty            string           `json:"kty,omitempty"`
}

// Valid ...
func (t *PublicKey) Valid() bool {
	// This needs a true implementation
	return true
}

// JSON ...
func (t *PublicKey) JSON() string {
	e, err := json.Marshal(t)
	if err != nil {
		panic(err.Error())
	}
	return string(e)
}

// AssertTypeValid ...
func (t *PublicKey) AssertTypeValid() error {
	return nil
}

type openIDConfiguration []struct {
	Issuer                                    string   `json:"issuer,omitempty"`
	AuthorizationEndpoint                     string   `json:"authorization_endpoint,omitempty"`
	TokenEndpoint                             string   `json:"token_endpoint,omitempty"`
	UserinfoEndpoint                          string   `json:"userinfo_endpoint,omitempty"`
	RegistrationEndpoint                      string   `json:"registration_endpoint,omitempty"`
	JwksURI                                   string   `json:"jwks_uri,omitempty"`
	ResponseTypesSupported                    []string `json:"response_types_supported,omitempty,omitempty"`
	ResponseModesSupported                    []string `json:"response_modes_supported,omitempty"`
	GrantTypesSupported                       []string `json:"grant_types_supported,omitempty"`
	SubjectTypesSupported                     []string `json:"subject_types_supported,omitempty"`
	IDTokenSigningAlgValuesSupported          []string `json:"id_token_signing_alg_values_supported,omitempty"`
	ScopesSupported                           []string `json:"scopes_supported,omitempty"`
	TokenEndpointAuthMethodsSupported         []string `json:"token_endpoint_auth_methods_supported,omitempty"`
	ClaimsSupported                           []string `json:"claims_supported,omitempty"`
	CodeChallengeMethodsSupported             []string `json:"code_challenge_methods_supported,omitempty"`
	IntrospectionEndpoint                     string   `json:"introspection_endpoint,omitempty"`
	IntrospectionEndpointAuthMethodsSupported []string `json:"introspection_endpoint_auth_methods_supported,omitempty"`
	RevocationEndpoint                        string   `json:"revocation_endpoint,omitempty"`
	RevocationEndpointAuthMethodsSupported    []string `json:"revocation_endpoint_auth_methods_supported,omitempty"`
	EndSessionEndpoint                        string   `json:"end_session_endpoint,omitempty"`
	RequestParameterSupported                 bool     `json:"request_parameter_supported,omitempty"`
	RequestObjectSigningAlgValuesSupported    []string `json:"request_object_signing_alg_values_supported,omitempty"`
}

func openIDConfigurationFromJSON(b []byte) (*openIDConfiguration, error) {
	var t openIDConfiguration
	err := json.Unmarshal(b, &t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

type jwk struct {
	Kty string `json:"kty,omitempty"`
	Alg string `json:"alg,omitempty"`
	Kid string `json:"kid,omitempty"`
	Use string `json:"use,omitempty"`
	E   string `json:"e,omitempty"`
	N   string `json:"n,omitempty"`
	Crv string `json:"crv,omitempty"`
	X   string `json:"x,omitempty"`
	Y   string `json:"y,omitempty"`
}

func (t *jwk) json() string {
	e, err := json.Marshal(t)
	if err != nil {
		panic(err.Error())
	}
	return string(e)
}

type jwks struct {
	Keys []jwk `json:"keys"`
}

// JSON ...
func (t *jwks) json() string {
	e, err := json.Marshal(t)
	if err != nil {
		panic(err.Error())
	}
	return string(e)
}

// Build ...
func (t *Config) Build() *TokenStore {

	tokenCleanup := defaultTokenCleanup
	publicKeyCleaup := defaultPublicKeyCleaup
	requestTimeout := defaultRequestTimeout

	if t.TokenCleanup > 0 {
		tokenCleanup = t.TokenCleanup
	}

	if t.PublicKeyCleaup > 0 {
		publicKeyCleaup = t.PublicKeyCleaup
	}

	if t.RequestTimeout > 0 {
		requestTimeout = t.RequestTimeout
	}

	return &TokenStore{
		tokenCacheMap:     cachemap.NewCacheMap(tokenCleanup),
		publicKeyCacheMap: cachemap.NewCacheMap(publicKeyCleaup),
		httpClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: maxIdleConnections,
			},
			Timeout: time.Duration(requestTimeout) * time.Second,
		},
	}

}

// GetToken ...
func (t *TokenStore) GetToken(token string) (*Token, error) {

	if token == "" {
		panic("String 'token' is required")
	}

	shortTokenString := token[1:8] + "..."

	if e, exist := t.tokenCacheMap.Get(token); exist {

		xtoken := e.(*Token)

		zap.L().Debug(fmt.Sprintf("Token=%s found in cache", shortTokenString))

		if !xtoken.Valid() {
			zap.L().Debug(fmt.Sprintf("GetToken(%s)->[valid]", shortTokenString))
			return xtoken, nil
		}

		zap.L().Debug(fmt.Sprintf("GetToken(%s)->[expired]", shortTokenString))
		return nil, ErrExpired
	}

	zap.L().Debug(fmt.Sprintf("Token=%s not found in cache", shortTokenString))

	xtoken := &Token{}

	_, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {

		// SigningMethodRSA

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return nil, fmt.Errorf("Token claims have unexpected format")
		}

		for k, v := range claims {

			if k == "iss" {
				xtoken.Iss, _ = v.(string)
			}

			if k == "exp" {
				floatValue := v.(float64)
				xtoken.Exp = int64(floatValue)
			}

			if k == "aud" {
				xtoken.Aud, _ = v.(string)
			}

			switch v.(type) {

			case map[string]interface{}:
				for subK, subV := range v.(map[string]interface{}) {
					if subK == "keytab" {
						if s, ok := subV.(string); ok {
							xtoken.Keytabs = append(xtoken.Keytabs, s)
						}
					}
				}

			default:
				// do nothing, just return attributes
			}

			if k == "keytab" {
				if s, ok := v.(string); ok {
					xtoken.Keytabs = append(xtoken.Keytabs, s)
				}
			}

		}

		xtoken.Kid = token.Header["kid"].(string)
		xtoken.Alg = token.Header["alg"].(string)

		if xtoken.Iss == "" {
			return nil, fmt.Errorf("Issuer(iss) not found in claims")
		}

		if !strings.HasPrefix(xtoken.Iss, "https://") {
			return nil, fmt.Errorf("Issuer is not valid")
		}

		if xtoken.Exp == 0 {
			return nil, fmt.Errorf("Expiration(exp) not found")
		}

		publicKey, err := t.getKey(xtoken.Iss, xtoken.Kid)
		if err != nil {
			return nil, err
		}

		if publicKey.Kty == "" {
			return nil, fmt.Errorf("kty is empty. should be EC or RSA")
		}

		switch token.Method.(type) {
		case *jwt.SigningMethodECDSA:
			if publicKey.Kty != "EC" {
				return nil, fmt.Errorf("Expected value for kty is EC not %s", publicKey.Kty)
			}
			break
		case *jwt.SigningMethodRSA:
			if publicKey.Kty != "RSA" {
				return nil, fmt.Errorf("Expected value for kty is RSA not %s", publicKey.Kty)
			}
			break
		default:
			return nil, fmt.Errorf("Signing type %s unsupported", publicKey.Kty)
		}

		return publicKey.EcdsaPublicKey, nil

	})

	if err != nil {
		if err.Error() == "Token is expired" {
			// Consume and log
		} else if err.Error() == "Token used before issued" {
			// Consume and log
		} else {
			zap.L().Debug(fmt.Sprintf("GetToken(%s)->[err=%s]", shortTokenString, err))
			return nil, err
		}
	}

	if !xtoken.Valid() {
		zap.L().Debug(fmt.Sprintf("GetToken(%s)->[err=%s]", shortTokenString, ErrExpired))
		return nil, ErrExpired
	}

	t.tokenCacheMap.Put(token, xtoken)

	zap.L().Debug(fmt.Sprintf("GetToken(%s)->[valid and added to cache]", shortTokenString))
	return xtoken, nil
}

func (t *TokenStore) getOpenIDConfiguration(fqdn string) (*openIDConfiguration, error) {

	resp, err := t.httpClient.Get(fqdn)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return openIDConfigurationFromJSON(b)
}

func (t *TokenStore) getJWKs(fqdn string) (*jwks, error) {

	resp, err := t.httpClient.Get(fqdn)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result jwks
	err = json.Unmarshal(b, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (t *TokenStore) getKey(iss, kid string) (*PublicKey, error) {

	if iss == "" {
		panic("iss must not be empty")
	}

	if kid == "" {
		panic("kid must not be empty")
	}

	issKid := iss + ":" + kid

	if e, exist := t.publicKeyCacheMap.Get(issKid); exist {

		publicKey := e.(*PublicKey)

		return publicKey, nil
	}

	openIDConfiguration, err := t.getOpenIDConfiguration(iss)
	if err != nil {
		return nil, err
	}

	// This is ugly. Could result in many errors logged when one will suffice
	for _, config := range *openIDConfiguration {
		if strings.HasPrefix(config.JwksURI, "https://") {
			jwks, err := t.getJWKs(config.JwksURI)
			if err == nil {
				for _, jwk := range jwks.Keys {
					if jwk.Kid == kid {
						key, err := newKey(&jwk)
						if err == nil {
							t.publicKeyCacheMap.Put(issKid, key)
							zap.L().Info(fmt.Sprintf("key for iss %s and kid %s created and added to cache", iss, kid))
							return key, nil
						}
						zap.L().Error(err.Error())
					}
				}
			} else {
				zap.L().Error(err.Error())
			}
		} else {
			zap.L().Debug(fmt.Sprintf("JWKS URL %s malformed", config.JwksURI))
		}

	}

	return nil, ErrNotFound
}

func newKey(jwk *jwk) (*PublicKey, error) {

	if jwk.Kty == "" {
		return nil, fmt.Errorf("Kty is empty")
	}

	switch jwk.Kty {
	case "EC":
		return newKeyEC(jwk)
	case "RSA":
		panic("Not yet supported")
	}

	return nil, fmt.Errorf("jwk kty type %s not supported", jwk.Kty)
}

func newKeyEC(jwk *jwk) (*PublicKey, error) {

	var curve elliptic.Curve

	switch jwk.Alg {

	case "ES224":
		curve = elliptic.P224()
	case "ES256":
		curve = elliptic.P256()
	case "ES384":
		curve = elliptic.P384()
	case "ES521":
		curve = elliptic.P521()

	default:
		return nil, fmt.Errorf("Curve %s not supported", jwk.Alg)
	}

	byteX, err := base64.RawURLEncoding.DecodeString(jwk.X)
	if err != nil {
		return nil, err
	}

	byteY, err := base64.RawURLEncoding.DecodeString(jwk.Y)
	if err != nil {
		return nil, err
	}

	return &PublicKey{
		EcdsaPublicKey: &ecdsa.PublicKey{
			Curve: curve,
			X:     new(big.Int).SetBytes(byteX),
			Y:     new(big.Int).SetBytes(byteY),
		},
		Created: time.Now().Unix(),
		Kty:     jwk.Kty,
	}, nil

}

// Shutdown ...
func (t *TokenStore) Shutdown() {
	zap.L().Debug("Shutting down")
	t.publicKeyCacheMap.Shutdown()
	t.tokenCacheMap.Shutdown()
}

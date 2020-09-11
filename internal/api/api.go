package api

import (
	"errors"
	"fmt"
	"kbridge/internal/keytabstore"
	"kbridge/internal/noncestore"
	"kbridge/internal/tokenstore"

	"github.com/jinzhu/copier"
	"go.uber.org/zap"
)

// ErrExpired ...
var ErrExpired error = errors.New("Expired")

// API ...
type API struct {
	tokenStore  *tokenstore.TokenStore
	keytabStore *keytabstore.KeytabStore
	nonceStore  *noncestore.NonceStore
}

// NewAPI ...
func NewAPI(tokenStore *tokenstore.TokenStore, keytabStore *keytabstore.KeytabStore, nonceStore *noncestore.NonceStore) *API {
	return &API{
		tokenStore:  tokenStore,
		keytabStore: keytabStore,
		nonceStore:  nonceStore,
	}
}

// NewNonce ...
func (t *API) NewNonce(token string) (*noncestore.Nonce, error) {

	// Need to validate iss

	if token == "" {
		panic("String 'token' is empty")
	}

	shortToken := token[1:8] + "..."

	_, err := t.tokenStore.GetToken(token)

	if err != nil {
		zap.L().Debug(fmt.Sprintf("NewNonce(%s)->[err=%s]", shortToken, err))
		return nil, err
	}

	nonce := t.nonceStore.NewNonce()
	shortNonce := nonce.Value[1:8] + "..."

	// Make a copy of the entity to hand out so that encapsulation is preserved
	clone := &noncestore.Nonce{}
	err = copier.Copy(&clone, &nonce)
	if err != nil {
		panic(err)
	}
	zap.L().Debug(fmt.Sprintf("NewNonce(%s)->[%s]", shortToken, shortNonce))
	return clone, nil
}

// GetKeytab ...
func (t *API) GetKeytab(token, principal string) (*keytabstore.Keytab, error) {

	// Need to validate iss

	if token == "" {
		panic("String 'token' is empty")
	}

	if principal == "" {
		panic("String 'principal' is empty")
	}

	shortToken := token[1:8] + "..."

	xtoken, err := t.tokenStore.GetToken(token)
	if err != nil {
		return nil, err
	}

	if xtoken.Aud == "" {
		err = fmt.Errorf("Audience is empty")
		zap.L().Debug(fmt.Sprintf("GetKeytab(token=%s,principal=%s)->[err=%s]", shortToken, principal, err))
		return nil, err
	}

	_, err = t.nonceStore.GetNonce(xtoken.Aud)
	if err != nil {
		zap.L().Debug(fmt.Sprintf("GetKeytab(token=%s,principal=%s)->[err=%s]", shortToken, principal, err))
		return nil, err
	}

	principalMatch := false
	for _, _principal := range xtoken.Keytabs {
		if principal == _principal {
			principalMatch = true
			break
		}
	}

	if !principalMatch {
		err = fmt.Errorf("Requested principal does not match authorized principals")
		zap.L().Debug(fmt.Sprintf("GetKeytab(token=%s,principal=%s)->[err=%s]", shortToken, principal, err))
		return nil, err
	}

	keytab, err := t.keytabStore.GetKeytab(principal)

	if err != nil {
		zap.L().Debug(fmt.Sprintf("GetKeytab(token=%s,principal=%s)->[err=%s]", shortToken, principal, err))
		return nil, err
	}

	clone := &keytabstore.Keytab{}
	err = copier.Copy(&clone, &keytab)
	if err != nil {
		panic(err)
	}

	zap.L().Debug(fmt.Sprintf("GetKeytab(token=%s,principal=%s)->[keytab.Principal=%s]", shortToken, principal, keytab.Principal))
	return clone, nil

}

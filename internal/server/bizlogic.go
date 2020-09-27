package server

import (
	"context"
	"fmt"

	"github.com/jodydadescott/kerberos-bridge/internal/keytabs"
	"github.com/jodydadescott/kerberos-bridge/internal/nonces"
	"go.uber.org/zap"
)

func (t *Server) newNonce(ctx context.Context, token string) (*nonces.Nonce, error) {

	if token == "" {
		zap.L().Debug("Token is empty")
		return nil, ErrAuthFail
	}

	shortToken := token[1:8] + "..."

	xtoken, err := t.tokenCache.GetToken(token)
	if err != nil {
		zap.L().Debug(fmt.Sprintf("NewNonce(%s)->[err=%s]", shortToken, err))
		return nil, ErrAuthFail
	}

	// Validate that token is allowed to pull nonce
	decision, err := t.policy.renderDecision(ctx, xtoken)
	if err != nil {
		zap.L().Debug(fmt.Sprintf("NewNonce(%s)->[err=%s]", shortToken, err))
		return nil, ErrAuthFail
	}

	if !decision.GetNonce {
		err = fmt.Errorf("Authorization denied")
		zap.L().Debug(fmt.Sprintf("NewNonce(%s)->[err=%s]", shortToken, err))
		return nil, ErrAuthFail
	}

	nonce := t.nonceCache.NewNonce()
	shortNonce := nonce.Value[1:8] + "..."

	zap.L().Debug(fmt.Sprintf("NewNonce(%s)->[%s]", shortToken, shortNonce))
	return nonce, nil
}

func (t *Server) getKeytab(ctx context.Context, token, principal string) (*keytabs.Keytab, error) {

	shortToken := ""

	if token == "" || principal == "" {
		var err error

		if token == "" && principal == "" {
			err = fmt.Errorf("Token and Principal are empty")
		} else if token == "" {
			err = fmt.Errorf("Token is empty")
		} else {
			err = fmt.Errorf("Principal is empty")
		}

		shortToken = token[1:8] + ".."
		zap.L().Debug(fmt.Sprintf("GetKeytab(token=%s,principal=%s)->[err=%s]", shortToken, principal, err))
		return nil, ErrAuthFail
	}

	shortToken = token[1:8] + ".."

	xtoken, err := t.tokenCache.GetToken(token)
	if err != nil {
		zap.L().Debug(fmt.Sprintf("GetKeytab(token=%s,principal=%s)->[err=%s]", shortToken, principal, err))
		return nil, ErrAuthFail
	}

	if xtoken.Aud == "" {
		err = fmt.Errorf("Audience is empty")
		zap.L().Debug(fmt.Sprintf("GetKeytab(token=%s,principal=%s)->[err=%s]", shortToken, principal, err))
		return nil, ErrAuthFail
	}

	nonce := t.nonceCache.GetNonce(xtoken.Aud)
	if nonce == nil {
		zap.L().Debug(fmt.Sprintf("GetKeytab(token=%s,principal=%s)->[err=%s]", shortToken, principal, err))
		return nil, ErrAuthFail
	}

	decision, err := t.policy.renderDecision(ctx, xtoken)
	if err != nil {
		zap.L().Debug(fmt.Sprintf("GetKeytab(token=%s,principal=%s)->[err=%s]", shortToken, principal, err))
		return nil, ErrAuthFail
	}

	if !decision.hasPrincipal(principal) {
		err = fmt.Errorf("Authorization denied")
		zap.L().Debug(fmt.Sprintf("GetKeytab(token=%s,principal=%s)->[err=%s]", shortToken, principal, err))
		return nil, ErrAuthFail
	}

	keytab := t.keytabCache.GetKeytab(principal)

	if keytab == nil {
		zap.L().Debug(fmt.Sprintf("GetKeytab(token=%s,principal=%s)->[err=%s]", shortToken, principal, err))
		return nil, ErrNotFound
	}

	zap.L().Debug(fmt.Sprintf("GetKeytab(token=%s,principal=%s)->[valid keytab]", shortToken, principal))
	return keytab, nil

}

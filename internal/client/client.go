package client

// import (
// 	"fmt"
// 	"io/ioutil"
// 	"net/http"
// 	"time"

// 	"github.com/jodydadescott/kerberos-bridge/internal/model"
// 	"go.uber.org/zap"
// )

// const (
// 	maxIdleConnections    int = 4
// 	defaultRequestTimeout int = 5
// 	requestTimeout        int = 60
// )

// // Config ...
// type Config struct {
// 	TokenHost          string
// 	KerberosBridgeHost string
// 	Principal          string
// }

// // Client ...
// type Client struct {
// 	httpClient         *http.Client
// 	tokenHost          string
// 	kerberosBridgeHost string
// 	principal          string
// 	token              string
// }

// // NewClient Returns a new client
// func NewClient(config *Config) (*Client, error) {

// 	if config.TokenHost == "" {
// 		return nil, fmt.Errorf("TokenHost is empty")
// 	}

// 	if config.KerberosBridgeHost == "" {
// 		return nil, fmt.Errorf("kerberosBridgeHost is empty")
// 	}

// 	if config.Principal == "" {
// 		return nil, fmt.Errorf("Principal is empty")
// 	}

// 	return &Client{
// 		httpClient: &http.Client{
// 			Transport: &http.Transport{
// 				MaxIdleConnsPerHost: maxIdleConnections,
// 			},
// 			Timeout: time.Duration(requestTimeout) * time.Second,
// 		},
// 		kerberosBridgeHost: config.KerberosBridgeHost,
// 		tokenHost:          config.TokenHost,
// 		principal:          config.Principal,
// 	}, nil
// }

// // Kinit ...
// func (t *Client) Kinit() error {

// 	zap.L().Debug("Getting initial token")
// 	token, err := t.getToken("init")
// 	if err != nil {
// 		return err
// 	}

// 	zap.L().Debug("Getting nonce")
// 	nonce, err := t.getNonce(token)
// 	if err != nil {
// 		return err
// 	}

// 	zap.L().Debug("Getting audience token")
// 	token, err = t.getToken(nonce.Value)
// 	if err != nil {
// 		return err
// 	}

// 	zap.L().Debug("Getting nonce")
// 	// getnonce

// 	return nil
// }

// func (t *Client) getToken(audience string) (*model.Token, error) {

// 	req, err := http.NewRequest("GET", t.tokenHost+"?type=OAUTH&audience="+audience, nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	req.Header.Set("X-Aporeto-Metadata", "secrets")

// 	resp, err := t.httpClient.Do(req)

// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	b, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if resp.StatusCode != http.StatusOK {
// 		return nil, fmt.Errorf(string(b))
// 	}

// 	token, err := model.TokenFromBase64(string(b))
// 	if err != nil {
// 		return nil, err
// 	}
// 	return token, nil
// }

// func (t *Client) getNonce(token string) (*model.Nonce, error) {

// 	resp, err := t.httpClient.Get(t.kerberosBridgeHost + "?bearertoken=" + token)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	b, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if resp.StatusCode == http.StatusOK {
// 		nonce, err := model.NonceFromJSON(b)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return nonce, nil
// 	}

// 	return nil, fmt.Errorf(string(b))
// }

// func (t *Client) getKeytab(token string) (*model.Keytab, error) {

// 	resp, err := t.httpClient.Get(t.kerberosBridgeHost + "?bearertoken=" + token + "&principal=" + t.principal)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	b, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if resp.StatusCode == http.StatusOK {
// 		nonce, err := model.NonceFromJSON(b)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return nonce, nil
// 	}

// 	return nil, fmt.Errorf(string(b))
// }

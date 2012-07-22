package httpcodec;

import(
	"fmt"
	"github.com/garyburd/go-oauth/oauth"
	"net/http"
	"net/rpc"
	"net/url"
)

type OAuthConfig struct {
	Consumer oauth.Credentials
	Token oauth.Credentials
}

func SetDefaultOAuthConfig(defaults OAuthConfig, config *OAuthConfig) {
	if config.Token.Token == "" {
		config.Token.Token = defaults.Token.Token
		config.Token.Secret = defaults.Token.Secret
		config.Consumer.Token = defaults.Consumer.Token
		config.Consumer.Secret = defaults.Consumer.Secret
	}
}

func NewOAuth1Encoder(config OAuthConfig) HeaderEncoder {
	var client oauth.Client = oauth.Client{Credentials: config.Consumer}
	var token oauth.Credentials = config.Token
	return func(r *rpc.Request, v interface{}, req *http.Request) error {
		fmt.Printf("blah\n")
		var signUrl url.URL = *req.URL
		var params = signUrl.Query()
	  // OAuth library appends the full set of signed params to the signing url.
		signUrl.RawQuery = ""
		client.SignParam(&token, req.Method, signUrl.String(), params)

		req.URL.RawQuery = params.Encode()
		return nil
	}
}
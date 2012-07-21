package httpcodec;

import(
	gooauth "github.com/csmcanarney/gooauth"
	"net/http"
	"net/rpc"
	"net/url"
)

type OAuthConfig struct {
	Token gooauth.Token
}

func SetDefaultOAuthConfig(defaults OAuthConfig, config *OAuthConfig) {
	if config.Token.ConsumerKey == "" {
		config.Token.ConsumerKey = defaults.Token.ConsumerKey
		config.Token.ConsumerSecret = defaults.Token.ConsumerSecret
		config.Token.Key = defaults.Token.Key
		config.Token.Secret = defaults.Token.Secret
		config.Token.SigMethod = defaults.Token.SigMethod
		// PrivateKey unused at the time of writing.
	}
}

func OAuthHeaderEncoder(config OAuthConfig) HeaderEncoder {
	var localConfig OAuthConfig = config
	return func(r *rpc.Request, v interface{}, req *http.Request) error {
		var signUrl url.URL = *req.URL
		var params = signUrl.Query()
	  // gooauth library appends the full set of signed params to the signing url.
		signUrl.RawQuery = ""
		if err := localConfig.Token.Sign(params, req.Method, signUrl.String()); err != nil {
			return err
		}

		req.URL.RawQuery = params.Encode()
		return nil
	}
}
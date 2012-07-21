package main;

import(
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"text/template"
)

type Encoding string
const(
	URLEncoded Encoding = "url"
	JSON = "json"
	XML = "xml"
)

type Field struct {
	DataType string `json:"type"`
	Fields map[string]Field `json:"fields"`
}

type Struct struct {
	Fields map[string]Field `json:"fields"`
}

type Method struct {
	HttpMethod string `json:"httpMethod"`
	Endpoint string `json:"endpoint"`
	Name string `json:"name"`
	RequestCodec Encoding `json:"requestCodec"`
	ResponseCodec Encoding `json:"responseCodec"`
	OAuthVersion int `json:"isOAuth"` // 0 means no oauth
}

func (m Method) IsOAuth1() bool {
	return m.OAuthVersion == 1
}

func (m Method) IsUrlEncoded() bool {
	return m.RequestCodec == URLEncoded
}

func (m Method) BodyEncoder() (string, error) {
	switch m.RequestCodec {
	case URLEncoded:
		return "nil", nil
	case JSON:
		return "httpcodec.JSONBodyEncoder", nil
	case XML:
		return "httpcodec.XMLBodyEncoder", nil
	}
	return "", errors.New(
		fmt.Sprintf("Don't know of an encoder for request codec %s", m.RequestCodec))
}

func (m Method) DecoderCtor() (string, error) {
	switch m.ResponseCodec {
	case JSON:
		return "httpcodec.JSONDecoder", nil
	case XML:
		return "httpcodec.XMLDecoder", nil
	}
	return "", errors.New(
		fmt.Sprintf("Don't know of a decoder for response codec %s", m.ResponseCodec))
}

type ServiceOAuthConfig struct {
	V1 ServiceOAuthV1Config `json:"v1"`
}

type ServiceOAuthV1Config struct {
	ConsumerKey string `json:"consumerKey"`
	ConsumerSecret string `json:"consumerSecret"`
	Key string `json:"token"`
	Secret string `json:"tokenSecret"`
}

type Service struct {
	Methods []Method `json:"methods"`
	Name string `json:"name"`
	OAuth ServiceOAuthConfig `json:"oauth"`
}

func (s Service) HasOAuth1Config() bool {
	return s.OAuth.V1.ConsumerKey != ""
}

type Client struct {
	Services []Service `json:"services"`
	PackageName string `json:"package"`
}

func (c Client) HasOAuth1Service() bool {
	for _, s := range c.Services {
		if s.HasOAuth1Config() {
			return true
		}
	}

	return false
}

func readConfig(r io.Reader, c *Client) bool {
	configDecoder :=  json.NewDecoder(r)
	if err := configDecoder.Decode(c); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot parse config: %s\n", err.Error())
		return false
	}
	return true
}

func validate(c Client) bool {
	for _, s := range(c.Services) {
		for _, m := range(s.Methods) {
			if (m.RequestCodec == JSON || m.RequestCodec == XML) && m.HttpMethod != "POST" {
				fmt.Fprintf(os.Stderr, "For non-urlencoded requests, method must be POST\n")
				return false
			}
		}
	}
	return true
}

func main() {
	config := Client{}
	if !readConfig(os.Stdin, &config) {
		return
	}

	if tmpl, err := template.New("out").Parse(codeTemplate); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal template parsing error: %s\n", err.Error())
	} else {
		if !validate(config) {
			return
		}
		if err = tmpl.Execute(os.Stdout, config); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing code: %s\n", err.Error())
		}
	}
}

const codeTemplate string = `package {{.PackageName}};

import(
{{if .HasOAuth1Service}}  gooauth "github.com/csmcanarney/gooauth"{{end}}
  "httpcodec"
  "net/http"
  "net/rpc"
)

{{range .Services}}
{{ $serviceName := printf "%sService" .Name }}
{{ $defaultOAuth := printf "default%sOAuthConfig" $serviceName }}
{{ $oauthConfig := .OAuth.V1 }}

// Definition of the {{.Name}} service.
type {{$serviceName}} struct {
  client *rpc.Client
}
{{if .HasOAuth1Config}}
var {{$defaultOAuth}} httpcodec.OAuthConfig = httpcodec.OAuthConfig{
{{with $oauthConfig}}
Token: gooauth.Token{
  ConsumerKey: "{{.ConsumerKey}}",
  ConsumerSecret: "{{.ConsumerSecret}}",
  Key: "{{.Key}}",
  Secret: "{{.Secret}}",
  SigMethod: gooauth.SM_HMAC,
},
{{end}}
}
{{end}}

func New{{$serviceName}}() *{{$serviceName}} {
  return New{{$serviceName}}WithClientConfig(new(http.Client), httpcodec.Config{})
}

func New{{$serviceName}}WithConfig(config httpcodec.Config) *{{$serviceName}} {
  return New{{$serviceName}}WithClientConfig(new(http.Client), config)
}

func New{{$serviceName}}WithClient(client *http.Client) *{{$serviceName}} {
  return New{{$serviceName}}WithClientConfig(client, httpcodec.Config{})
}

func New{{$serviceName}}WithClientConfig(client *http.Client, config httpcodec.Config) *{{$serviceName}} {
{{if .HasOAuth1Config}}  httpcodec.SetDefaultOAuthConfig({{$defaultOAuth}}, &config.OAuth){{end}}
  codec := httpcodec.NewHTTPCodec("{{.Name}}", client, config)
{{range .Methods}}
  // Method {{.Name}}
  codec.Register("{{.Name}}", &httpcodec.MethodCodec{
    HeaderEncoders: []httpcodec.HeaderEncoder{
      httpcodec.NewEndpointRequestEncoder("{{.HttpMethod}}", "{{.Endpoint}}"),
{{if .IsUrlEncoded}}      httpcodec.NewURLEncoder(),{{end}}
{{if .IsOAuth1}}      httpcodec.NewOAuth1Encoder(config.OAuth),{{end}}
    },
    BodyEncoder: {{.BodyEncoder}},
    HeaderDecoder: httpcodec.StandardHeaderDecoder,
    Decoder: {{.DecoderCtor}},
  })

{{end}}
  codec.Start()
  return &{{.Name}}Service{client: rpc.NewClientWithCodec(codec)}
}

{{range .Methods}}
func (s *{{$serviceName}}) {{.Name}}(args interface{}, reply interface{}) error {
  return s.client.Call("{{.Name}}", args, reply)
}
{{end}}

{{end}}
`

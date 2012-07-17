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
	Method string `json:"method"`
	Endpoint string `json:"endpoint"`
	Name string `json:"name"`
	RequestCodec Encoding `json:"requestCodec"`
	ResponseCodec Encoding `json:"responseCodec"`
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

type Service struct {
	Methods []Method `json:"methods"`
	Name string `json:"name"`
}

type Client struct {
	Services []Service `json:"services"`
	PackageName string `json:"packageName"`
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
			if (m.RequestCodec == JSON || m.RequestCodec == XML) && m.Method != "POST" {
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
  "httpcodec"
)

{{range .Services}}
{{ $serviceName := printf "%sService" .Name }}

// Definition of the {{.Name}} service.
type {{$serviceName}} struct {
  client *rpc.Client
}

func New{{$serviceName}}() *{{$serviceName}} {
  return New{{$serviceName}}WithClient(new(http.Client))
}

func New{{$serviceName}}WithClient(client *http.Client) *{{$serviceName}} {
  codec := httpcodec.NewHTTPCodec("{{.Name}}", client)
{{range .Methods}}
  // Method {{.Name}}
  codec.Register("{{.Name}}", httpcodec.MethodCodec{
    HeaderEncoders: [
      httpcodec.NewEndpointRequestEncoder("{{.Method}}", "{{.Endpoint}}"),
{{if .IsUrlEncoded}}      httpcodec.NewURLEncoder(),{{end}}
    ],
    BodyEncoder: {{.BodyEncoder}},
    HeaderDecoder: httpcodec.StandardHeaderDecoder,
    Decoder: {{.DecoderCtor}}
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

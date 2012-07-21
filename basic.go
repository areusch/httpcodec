package httpcodec;

import(
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/rpc"
	"net/url"
)

func NewEndpointRequestEncoder(method string, rawurl string) HeaderEncoder {
	if endpoint, err := url.Parse(rawurl); err == nil {
		return func(req *rpc.Request, v interface{}, httpReq *http.Request) error {
			httpReq.Method = method
			httpReq.URL = endpoint
			httpReq.Proto = "HTTP/1.1"
			httpReq.ProtoMajor = 1
			httpReq.ProtoMinor = 1
			httpReq.Header = make(http.Header)
			httpReq.Host = endpoint.Host
			return nil
		}
	} else {
		panic(err)
	}
	return nil
}

func NewURLEncoder() HeaderEncoder {
	return func(req *rpc.Request, v interface{}, httpReq *http.Request) (err error) {
		var query url.Values
		if query, err = url.ParseQuery(httpReq.URL.RawQuery); err != nil {
			return
		}
		URLEncoder{query}.Encode(v)
		httpReq.URL.RawQuery = query.Encode()
		return nil
	}
}

func JSONBodyEncoder(w io.Writer) Encoder {
	return json.NewEncoder(w)
}

func XMLBodyEncoder(w io.Writer) Encoder {
	return xml.NewEncoder(w)
}

// Returns an error if response code != 200
func StandardHeaderDecoder(resp *http.Response, rpcResp *rpc.Response) error {
	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Server returned HTTP response: %s", resp.Status))
	}
	return nil
}

func JSONDecoder(r io.Reader) Decoder {
	return json.NewDecoder(r)
}

func XMLDecoder(r io.Reader) Decoder {
	return xml.NewDecoder(r)
}

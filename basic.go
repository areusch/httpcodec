package httpcodec;

import(
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/rpc"
	"net/url"
)

func NewEndpointRequestEncoder(method string, endpoint *url.URL) HeaderEncoder {
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

func JSONDecoder(r io.Reader) Decoder {
	return json.NewDecoder(r)
}

func XMLDecoder(r io.Reader) Decoder {
	return xml.NewDecoder(r)
}

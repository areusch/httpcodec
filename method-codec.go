package httpcodec;

import(
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/rpc"
)


type Decoder interface {
	Decode(v interface{}) error
}

type DecoderCtor func(io.Reader) Decoder

type Encoder interface {
	Encode(v interface{}) error
}

type EncoderCtor func(io.Writer) Encoder

type HeaderEncoder func(*rpc.Request, interface{}, *http.Request) error
type StatusDecoder func(*http.Response, *rpc.Response) error

type MethodCodec struct {
	// Invoked first to create the http request.
	HeaderEncoders []HeaderEncoder

	// If non-nil, invoked to encode the body of the request.
	BodyEncoder EncoderCtor

	StatusDecoder StatusDecoder
	Decoder DecoderCtor
}

func (m *MethodCodec) CreateRequest(req *rpc.Request, v interface{}) (httpReq *http.Request, err error) {
	httpReq = new(http.Request)
	for _, e := range(m.HeaderEncoders) {
		if err = e(req, v, httpReq); err != nil {
			return
		}
	}

	if m.BodyEncoder != nil {
		bodyBuffer := new(bytes.Buffer)
		encoder := m.BodyEncoder(bodyBuffer)
		if err = encoder.Encode(v); err != nil {
			httpReq = nil
			return
		}

		httpReq.ContentLength = int64(bodyBuffer.Len())
		httpReq.Body = ioutil.NopCloser(bodyBuffer)
	}

	return httpReq, nil
}

func (m *MethodCodec) ReadStatus(httpReply *http.Response, resp *rpc.Response) error {
	return m.StatusDecoder(httpReply, resp)
}

func (m *MethodCodec) ReadReturnValue(httpReply *http.Response, resp interface{}) (err error) {
	decoder := m.Decoder(httpReply.Body)
	err = decoder.Decode(resp)
	return
}

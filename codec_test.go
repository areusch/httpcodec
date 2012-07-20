package httpcodec;

import(
	"io"
	"net/http"
	"net/rpc"
	"testing"
)

type TestBodyCoder uint64

func MakeTestBodyEncoder(w io.Writer) *TestBodyCoder {
	return new(TestBodyCoder)
}

func MakeTestBodyDecoder(r io.Reader) *TestBodyCoder {
	return new(TestBodyCoder)
}

func (t *TestBodyCoder) Encode(v interface{}) error {
	(*t)++
	return nil
}

func (t *TestBodyCoder) Decode(v interface{}) error {
	(*t)++
	return nil
}

func (t TestBodyCoder) GetNumCalls() uint64 {
	return uint64(t)
}

func Test_oneRequest(t *testing.T) {
	var encoder, decoder *TestBodyCoder
	numCoderCalls := 0
	v := new(int)
	methodCodec := MethodCodec{
	HeaderEncoders: []HeaderEncoder{
			func(r *rpc.Request, v interface{}, h *http.Request) error {
				if r != nil {
					t.Fatalf("Expected nil RPC request, got %v", r)
				}
				numCoderCalls++
				return nil
		}},
	BodyEncoder: func(w io.Writer) Encoder {
			encoder = MakeTestBodyEncoder(w)
			return encoder
		},
	HeaderDecoder: func(r *http.Response, resp *rpc.Response) error {
			if numCoderCalls == 0 {
				t.Fatalf("Expected a call to header encoders, got none")
			}
			numCoderCalls++
			return nil
		},
	Decoder: func(r io.Reader) Decoder {
			decoder = MakeTestBodyDecoder(r)
			return decoder
	}}

	if req, err := methodCodec.CreateRequest(nil, v); err != nil {
		t.Fatalf("Expected no error, received %v", err)
	} else if req == nil {
		t.Fatalf("Received nil req and err")
	}

	if numCoderCalls != 1 {
		t.Fatalf("Expected 1 coder call, got %d", numCoderCalls)
	}

	if encoder == nil {
		t.Fatalf("Expected encoder instantiated, never was")
	}

	if encoder.GetNumCalls() != 1 {
		t.Fatalf("Expected 1 call to encode, saw %d", encoder.GetNumCalls())
	}

	if err := methodCodec.ReadResponseHeader(nil, nil); err != nil {
		t.Fatalf("Expected no error, received %v", err)
	}

	if numCoderCalls != 2 {
		t.Fatalf("Expected 2 coder calls, got %d", numCoderCalls)
	}

	resp := http.Response{}
	if err := methodCodec.ReadReturnValue(&resp, v); err != nil {
		t.Fatalf("Expected no error, received %v", err)
	}

	if decoder == nil {
		t.Fatalf("Expected decoder instantiated, never was")
	}

	if decoder.GetNumCalls() != 1 {
		t.Fatalf("Expected 1 call to decode, saw %d", decoder.GetNumCalls())
	}
}

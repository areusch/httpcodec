package httpcodec;

import(
	"fmt"
	"net/http"
	"net/rpc"
)

type ActiveCall struct {
	codec *MethodCodec
	seq uint64
	request *http.Request
	reply *http.Response
	err error
}

type HTTPCodec struct {
	Service string
	Open bool
	client *http.Client
	methods map[string]*MethodCodec
	requestQueue chan *ActiveCall
	replyQueue chan *ActiveCall
	bodyQueue chan *ActiveCall
}

func NewHTTPCodec(Service string, client *http.Client) *HTTPCodec {
	return &HTTPCodec{Service: Service, Open: true, client: client, methods: make(map[string]*MethodCodec)}
}

func (c *HTTPCodec) Register(method string, codec *MethodCodec) {
	if c.IsStarted() {
		panic("Register called after HTTPCodec started!")
	}

	c.methods[method] = codec
}

func (c HTTPCodec) IsStarted() bool {
	return c.requestQueue != nil
}

func (c *HTTPCodec) Start() {
	if c.IsStarted() {
		panic("Started HTTPCodec twice!")
	}

	c.requestQueue = make(chan *ActiveCall, 1)
	c.replyQueue = make(chan *ActiveCall, 1)
	c.bodyQueue = make(chan *ActiveCall, 1)
	go c.requestSender()
}

func (c *HTTPCodec) requestSender() {
	for req, ok := <- c.requestQueue; ok; req = <- c.requestQueue {
		req.reply, req.err = c.client.Do(req.request)
		c.replyQueue <- req
	}
}

func (c *HTTPCodec) RunRequest(codec *MethodCodec, req *rpc.Request, v interface{}) (err error) {
	var httpReq *http.Request
	if httpReq, err = codec.CreateRequest(req, v); err != nil {
		return
	}

	c.requestQueue <- &ActiveCall{codec: codec, request: httpReq}
	return
}

type ErrNoSuchMethod struct {
	Service string
	Method string
}

func (e ErrNoSuchMethod) Error() string {
	return fmt.Sprintf("No such RPC method: %s.%s", e.Service, e.Method)
}

func (c *HTTPCodec) WriteRequest(req *rpc.Request, v interface{}) (err error) {
	if codec, ok := c.methods[req.ServiceMethod]; !ok {
		err = ErrNoSuchMethod{Service: c.Service, Method: req.ServiceMethod}
	} else {
		err = c.RunRequest(codec, req, v)
	}
	return
}

func (c *HTTPCodec) ReadResponseHeader(resp *rpc.Response) (err error) {
	req := <-c.replyQueue
	resp.Seq = req.seq
	if req.err != nil {
		resp.Error = req.err.Error()
	} else {
		if e2 := req.codec.ReadResponseHeader(req.reply, resp); e2 != nil {
			resp.Error = e2.Error()
		}
	}
	c.bodyQueue <- req
	return
}

func (c *HTTPCodec) ReadResponseBody(v interface{}) (err error) {
	req := <-c.bodyQueue
	err = req.codec.ReadReturnValue(req.reply, v)
	req.reply.Body.Close()
	return
}

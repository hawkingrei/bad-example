package liverpc

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"strings"
	"time"

	"go-common/library/conf/env"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/net/trace"

	"github.com/gogo/protobuf/proto"
	"github.com/json-iterator/go"
	"github.com/pkg/errors"
)

// ClientConn connect represent a real client connection to a rpc server
type ClientConn struct {
	addr        string
	network     string
	rwc         io.ReadWriteCloser
	Timeout     time.Duration
	DialTimeout time.Duration
}

type fullReqMsg struct {
	Header *Header     `json:"header"`
	HTTP   *HTTP       `json:"http"`
	Body   interface{} `json:"body"`
}

// Dial dial a rpc server
func Dial(ctx context.Context, network, addr string, timeout time.Duration, connTimeout time.Duration) (*ClientConn, error) {
	c := &ClientConn{
		addr:        addr,
		network:     network,
		Timeout:     timeout,
		DialTimeout: connTimeout,
	}
	conn, err := net.DialTimeout(c.network, c.addr, c.DialTimeout)
	if err != nil {
		return nil, err
	}
	c.rwc = conn
	return c, err
}

// Close close the caller connection.
func (c *ClientConn) Close() error {
	if c.rwc != nil {
		return c.rwc.Close()
	}
	return nil
}

func (c *ClientConn) writeRequest(ctx context.Context, req *protoReq) (err error) {
	var (
		headerBuf = make([]byte, _headerLen)
		header    = req.Header
		body      = req.Body
	)
	binary.BigEndian.PutUint32(headerBuf[0:4], header.magic)
	binary.BigEndian.PutUint32(headerBuf[4:8], header.timestamp)
	binary.BigEndian.PutUint32(headerBuf[8:12], header.checkSum)
	binary.BigEndian.PutUint32(headerBuf[12:16], header.version)
	binary.BigEndian.PutUint32(headerBuf[16:20], header.reserved)
	binary.BigEndian.PutUint32(headerBuf[20:24], header.seq)
	binary.BigEndian.PutUint32(headerBuf[24:28], uint32(len(body)))
	copy(headerBuf[28:60], header.cmd)
	if _, err = c.rwc.Write(headerBuf); err != nil {
		err = errors.Wrap(err, "write req header error")
		return
	}
	if log.V(2) {
		log.Info("liverpc body: %s", string(body))
	}
	if _, err = c.rwc.Write(body); err != nil {
		err = errors.Wrap(err, "write req body error")
		return
	}
	return
}

func (c *ClientConn) readResponse(ctx context.Context, resp *protoResp) (err error) {
	var (
		headerBuf = make([]byte, _headerLen)
		length    int
	)
	if _, err = c.rwc.Read(headerBuf); err != nil {
		err = errors.Wrap(err, "read resp header error")
		return
	}
	resp.Header.magic = binary.BigEndian.Uint32(headerBuf[0:4])
	resp.Header.timestamp = binary.BigEndian.Uint32(headerBuf[4:8])
	resp.Header.checkSum = binary.BigEndian.Uint32(headerBuf[8:12])
	resp.Header.version = binary.BigEndian.Uint32(headerBuf[12:16])
	resp.Header.reserved = binary.BigEndian.Uint32(headerBuf[16:20])
	resp.Header.seq = binary.BigEndian.Uint32(headerBuf[20:24])
	resp.Header.length = binary.BigEndian.Uint32(headerBuf[24:28])
	resp.Header.cmd = headerBuf[28:60]
	resp.Body = make([]byte, resp.Header.length)
	if length, err = io.ReadFull(c.rwc, resp.Body); err != nil {
		err = errors.Wrap(err, "read resp body error")
		return
	}
	if uint32(length) != resp.Header.length {
		err = errors.New("bad resp body data")
		return
	}
	return
}

func (c *ClientConn) composeReqPackHeader(reqPack *protoReq, version int, serviceMethod string) {
	reqPack.Header.magic = _magic
	reqPack.Header.checkSum = 0
	reqPack.Header.seq = 1
	reqPack.Header.timestamp = uint32(time.Now().Unix())
	reqPack.Header.reserved = 0
	reqPack.Header.version = uint32(version)
	// command: {message_type}controller.method
	reqPack.Header.cmd = make([]byte, 32)
	reqPack.Header.cmd[0] = _cmdReqType
	// serviceMethod: Room.room_init
	copy(reqPack.Header.cmd[1:], []byte(serviceMethod))
}

func (c *ClientConn) setupDeadline(ctx context.Context) error {
	t, _ := ctx.Value(KeyTimeout).(time.Duration)
	if t == 0 {
		t = c.Timeout
	}

	conn := c.rwc.(net.Conn)
	if conn != nil {
		err := conn.SetDeadline(time.Now().Add(t))
		if err != nil {
			conn.Close()
			return err
		}
	}
	return nil
}

// CallRaw call the service method, waits for it to complete, and returns reply its error status.
// this is can be use without protobuf
// client: {service}
// serviceMethod: {version}|{controller.method}
// httpURL: /room/v1/Room/room_init
// httpURL: /{service}/{version}/{controller}/{method}
func (c *ClientConn) CallRaw(ctx context.Context, version int, serviceMethod string, in *Args) (out *Reply, err error) {
	var (
		reqPack  protoReq
		respPack protoResp
		code     = "0"
		now      = time.Now()
	)
	if err = c.setupDeadline(ctx); err != nil {
		return
	}
	// it is ok for request http field to be nil

	if in.Header == nil {
		hdr, _ := ctx.Value(KeyHeader).(*Header)
		if hdr == nil {
			hdr = createHeader(ctx)
		}
		in.Header = hdr
	}
	if in.HTTP == nil {
		htp, _ := ctx.Value(KeyHTTP).(*HTTP)
		in.HTTP = htp
	}
	if in.Body == nil {
		in.Body = map[string]interface{}{}
	}

	c.composeReqPackHeader(&reqPack, version, serviceMethod)

	defer func() {
		stats.Timing(serviceMethod, int64(time.Since(now)/time.Millisecond))
		if code != "" {
			stats.Incr(serviceMethod, code)
		}
	}()
	var reqBytes []byte
	if reqBytes, err = json.Marshal(in); err != nil {
		err = errors.Wrap(err, "CallRaw json marshal error")
		code = "marshalErr"
		return
	}
	reqPack.Body = reqBytes

	if err = c.writeRequest(ctx, &reqPack); err != nil {
		code = "writeErr"
		return
	}
	if err = c.readResponse(ctx, &respPack); err != nil {
		code = "readErr"
		return
	}

	out = &Reply{}
	if err = json.Unmarshal(respPack.Body, out); err != nil {
		err = errors.Wrap(err, "proto unmarshal error: "+string(respPack.Body))
		code = "unmarshalErr"
		return
	}
	return
}

// Call call the service method, waits for it to complete, and returns its error status.
// this is used with protobuf generated msg
// client: {service}
// serviceMethod: {version}|{controller.method}
// httpURL: /room/v1/Room/room_init
// httpURL: /{service}/{version}/{controller}/{method}
func (c *ClientConn) Call(ctx context.Context, version int, serviceMethod string, in, out proto.Message) (err error) {
	var (
		reqPack  protoReq
		respPack protoResp
		code     = "0"
		now      = time.Now()
	)
	if err = c.setupDeadline(ctx); err != nil {
		return
	}
	fullMsg := &fullReqMsg{}

	hdr, _ := ctx.Value(KeyHeader).(*Header)
	if hdr != nil {
		fullMsg.Header = hdr
	} else {
		fullMsg.Header = createHeader(ctx)
	}

	htp, _ := ctx.Value(KeyHTTP).(*HTTP)
	if htp != nil {
		fullMsg.HTTP = htp
	}

	fullMsg.Body = in

	// it is ok for request http field to be nil

	c.composeReqPackHeader(&reqPack, version, serviceMethod)
	defer func() {
		stats.Timing(serviceMethod, int64(time.Since(now)/time.Millisecond))
		if code != "" {
			stats.Incr(serviceMethod, code)
		}
	}()
	var reqBody []byte
	if reqBody, err = json.Marshal(fullMsg); err != nil {
		err = errors.Wrap(err, "Call json marshal error")
		code = "marshalErr"
		return
	}
	reqPack.Body = reqBody

	if err = c.writeRequest(ctx, &reqPack); err != nil {
		code = "writeErr"
		return
	}
	if err = c.readResponse(ctx, &respPack); err != nil {
		code = "readErr"
		return
	}

	if err = jsoniter.Unmarshal(respPack.Body, out); err != nil {
		err = errors.Wrap(err, "proto unmarshal error: "+string(respPack.Body))
		code = "unmarshalErr"
		return
	}
	return
}

func createHeader(ctx context.Context) *Header {
	header := &Header{}
	header.UserIp = metadata.String(ctx, metadata.RemoteIP)
	header.Caller = strings.Replace(env.AppID, ".", "-", -1)
	if header.Caller == "" {
		header.Caller = "unknown"
	}
	tracer, ok := metadata.Value(ctx, metadata.Trace).(trace.Trace)
	if ok {
		trace.Inject(tracer, nil, header)
	}
	mid, _ := metadata.Value(ctx, "mid").(int64)
	header.Uid = mid
	//header.Platform = ctx.Request.FormValue("platform")
	return header
}

package libnet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

const maxBodySize = 1 << 12

/*
总长度
----|
header头长度
--|
header头
-|-|--|--|----|------...
版本号|状态码|消息类型|命令|seq|pb body体

header头长度=1字节版本号+1字节状态码+2字节消息类型+2字节命令+4字节seq
总长度=header头+header头长度+pb body体长度

----|--|-|-|--|--|----|body
总长度|header头长度|版本号|状态码|消息类型|命令|seq｜body
总长度=2+1+1+2+2+4+len(body)
header头长度=1+1+2+2+4
*/

const (
	packSize      = 4
	headerSize    = 2
	verSize       = 1
	statusSize    = 1
	serviceIdSize = 2
	cmdSize       = 2
	seqSize       = 4
	rawHeaderSize = verSize + statusSize + serviceIdSize + cmdSize + seqSize
	maxPackSize   = maxBodySize + rawHeaderSize + headerSize + packSize
	// offset
	headerOffset    = 0
	verOffset       = headerOffset + headerSize
	statusOffset    = verOffset + verSize
	serviceIdOffset = statusOffset + statusSize
	cmdOffset       = serviceIdOffset + serviceIdSize
	seqOffset       = cmdOffset + cmdSize
	bodyOffset      = seqOffset + seqSize
)

var (
	ErrRawPackLen   = errors.New("default server codec pack length error")
	ErrRawHeaderLen = errors.New("default server codec header length error")
)

type Header struct {
	Version   uint8
	Status    uint8
	ServiceId uint16
	Cmd       uint16
	Seq       uint32
}

type Message struct {
	Header
	Body []byte
}

func (m *Message) Fromat() string {
	return fmt.Sprintf("Version:%d, Status:%d, ServiceId:%d, Cmd:%d, Seq:%d, Body:%s",
		m.Version, m.Status, m.ServiceId, m.Cmd, m.Seq, string(m.Body))
}

type Protocol interface {
	NewCodec(conn net.Conn) Codec
}

type Codec interface {
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
	Receive() (*Message, error)
	Send(Message) error
	Close() error
}

type IMProtocol struct{}

func NewIMProtocol() Protocol {
	return &IMProtocol{}
}

func (p *IMProtocol) NewCodec(conn net.Conn) Codec {
	return &imCodec{conn: conn}
}

type imCodec struct {
	conn net.Conn
}

func (c *imCodec) readPackSize() (uint32, error) {
	return c.readUint32BE()
}

func (c *imCodec) readUint32BE() (uint32, error) {
	b := make([]byte, packSize)
	_, err := io.ReadFull(c.conn, b)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(b), nil
}

func (c *imCodec) readPacket(msgSize uint32) ([]byte, error) {
	b := make([]byte, msgSize)
	_, err := io.ReadFull(c.conn, b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (c *imCodec) Receive() (*Message, error) {
	packLen, err := c.readPackSize()
	if err != nil {
		return nil, err
	}

	if packLen > maxPackSize {
		return nil, ErrRawPackLen
	}

	buf, err := c.readPacket(packLen)
	if err != nil {
		return nil, err
	}

	var msg Message
	headerLen := binary.BigEndian.Uint16(buf[headerOffset:verOffset])
	msg.Version = buf[verOffset]
	msg.Status = buf[statusOffset]
	msg.ServiceId = binary.BigEndian.Uint16(buf[serviceIdOffset:cmdOffset])
	msg.Cmd = binary.BigEndian.Uint16(buf[cmdOffset:seqOffset])
	msg.Seq = binary.BigEndian.Uint32(buf[seqOffset:bodyOffset])
	//logx.Infof("msg.Seq:%+v", msg.Seq)

	if headerLen != rawHeaderSize {
		return nil, ErrRawHeaderLen
	}

	if packLen > uint32(headerLen) {
		msg.Body = buf[bodyOffset:packLen]
	}

	logx.Infof("receive msg:%+v", msg)
	return &msg, nil
}

func (c *imCodec) Send(msg Message) error {
	packLen := headerSize + rawHeaderSize + len(msg.Body)
	packLenBuf := make([]byte, packSize)
	binary.BigEndian.PutUint32(packLenBuf[:packSize], uint32(packLen))

	buf := make([]byte, packLen)
	// header
	binary.BigEndian.PutUint16(buf[headerOffset:], uint16(rawHeaderSize))
	buf[verOffset] = msg.Version
	buf[statusOffset] = msg.Status
	binary.BigEndian.PutUint16(buf[serviceIdOffset:], msg.ServiceId)
	binary.BigEndian.PutUint16(buf[cmdOffset:], msg.Cmd)
	binary.BigEndian.PutUint32(buf[seqOffset:], msg.Seq)

	// body
	copy(buf[headerSize+rawHeaderSize:], msg.Body)
	allBuf := append(packLenBuf, buf...)
	n, err := c.conn.Write(allBuf)
	if err != nil {
		return err
	}
	if n != len(allBuf) {
		return fmt.Errorf("n:%d, len(buf):%d", n, len(buf))
	}
	return nil
}

func (c *imCodec) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *imCodec) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

func (c *imCodec) Close() error {
	return c.conn.Close()
}

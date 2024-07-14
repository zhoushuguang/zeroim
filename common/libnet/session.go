package libnet

import (
	"errors"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zhoushuguang/zeroim/common/session"

	"github.com/zeromicro/go-zero/core/logx"
)

func init() {
	globalSessionId = uint64(rand.New(rand.NewSource(time.Now().Unix())).Uint32())
}

var (
	SessionClosedError  = errors.New("session Closed")
	SessionBlockedError = errors.New("session Blocked")

	globalSessionId uint64
)

type Session struct {
	id         uint64
	token      string
	codec      Codec
	manager    *Manager
	sendChan   chan Message
	closeFlag  int32
	closeChan  chan int
	closeMutex sync.Mutex
}

func NewSession(manager *Manager, codec Codec, sendChanSize int) *Session {
	s := &Session{
		codec:     codec,
		manager:   manager,
		closeChan: make(chan int),
		id:        atomic.AddUint64(&globalSessionId, 1),
	}
	if sendChanSize > 0 {
		s.sendChan = make(chan Message, sendChanSize)
		go s.sendLoop()
	}

	return s
}

func (s *Session) Name() string {
	return s.manager.Name
}

func (s *Session) ID() uint64 {
	return s.id
}

func (s *Session) Token() string {
	return s.token
}

func (s *Session) Session() session.Session {
	return session.NewSession(s.manager.Name, s.token, s.id)
}

func (s *Session) SetToken(token string) {
	s.token = token
}

func (s *Session) sendLoop() {
	defer s.Close()
	for {
		select {
		case msg := <-s.sendChan:
			err := s.codec.Send(msg)
			if err != nil {
				logx.Errorf("[sendLoop] s.codec.Send msg: %v error: %v", msg, err)
				return
			}
		case <-s.closeChan:
			return
		}
	}
}

func (s *Session) Receive() (*Message, error) {
	return s.codec.Receive()
}

func (s *Session) Send(msg Message) error {
	if s.IsClosed() {
		return SessionClosedError
	}
	if s.sendChan == nil {
		return s.codec.Send(msg)
	}
	select {
	case s.sendChan <- msg:
		return nil
	default:
		return SessionBlockedError
	}
}

func (s *Session) IsClosed() bool {
	return atomic.LoadInt32(&s.closeFlag) == 1
}

func (s *Session) Close() error {
	if atomic.CompareAndSwapInt32(&s.closeFlag, 0, 1) {
		err := s.codec.Close()
		close(s.closeChan)
		if s.manager != nil {
			s.manager.removeSession(s)
		}
		return err
	}

	return SessionClosedError
}

func (s *Session) SetReadDeadline(time time.Time) error {
	return s.codec.SetReadDeadline(time)
}

func (s *Session) SetWriteDeadline(time time.Time) error {
	return s.codec.SetWriteDeadline(time)
}

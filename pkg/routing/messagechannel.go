package routing

import (
	"sync"

	"github.com/livekit/protocol/livekit"
	"google.golang.org/protobuf/proto"
)

const DefaultMessageChannelSize = 200

type MessageChannel struct {
	connectionID livekit.ConnectionID
	msgChan      chan proto.Message
	onClose      func()
	isClosed     bool
	lock         sync.RWMutex
}

func NewDefaultMessageChannel(connectionID livekit.ConnectionID) *MessageChannel {
	return NewMessageChannel(connectionID, DefaultMessageChannelSize)
}

func NewMessageChannel(connectionID livekit.ConnectionID, size int) *MessageChannel {
	return &MessageChannel{
		connectionID: connectionID,
		// allow some buffer to avoid blocked writes
		msgChan: make(chan proto.Message, size),
	}
}

func (m *MessageChannel) OnClose(f func()) {
	m.onClose = f
}

func (m *MessageChannel) IsClosed() bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.isClosed
}

func (m *MessageChannel) WriteMessage(msg proto.Message) error {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if m.isClosed {
		return ErrChannelClosed
	}

	select {
	case m.msgChan <- msg:
		// published
		return nil
	default:
		// channel is full
		return ErrChannelFull
	}
}

func (m *MessageChannel) ReadChan() <-chan proto.Message {
	return m.msgChan
}

func (m *MessageChannel) Close() {
	m.lock.Lock()
	if m.isClosed {
		m.lock.Unlock()
		return
	}
	m.isClosed = true
	close(m.msgChan)
	m.lock.Unlock()

	if m.onClose != nil {
		m.onClose()
	}
}

func (m *MessageChannel) ConnectionID() livekit.ConnectionID {
	return m.connectionID
}

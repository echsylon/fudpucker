package entity

import (
	"echsylon/fudpucker/entity/unit"
	"echsylon/fudpucker/entity/utils"
)

type MessageType byte

func (t MessageType) String() string {
	switch t {
	case MessageTypeCommandHail:
		return "CommandHail"
	case MessageTypeCommandPatch:
		return "CommandPatch"
	case MessageTypeEventSync:
		return "EventSync"
	case MessageTypeEventPeer:
		return "EventPeer"
	case MessageTypeEventFarewell:
		return "EventFarewell"
	default:
		return "unknown"
	}
}

const (
	MessageTypeUnknown MessageType = iota
	MessageTypeCommandHail
	MessageTypeCommandPatch
	MessageTypeEventSync
	MessageTypeEventPeer
	MessageTypeEventFarewell
)

const (
	SignatureLength  = 0
	MaxMessageLength = 1 * unit.MiB
)

type Message interface {
	GetId() Id
	GetSender() Id
	GetType() MessageType
	GetData() []byte
	GetSignature() []byte
}

type message struct {
	id          Id
	sender      Id
	messageType MessageType
	data        []byte
}

func NewMessage(id Id, sender Id, messageType MessageType, bytes []byte) Message {
	return &message{
		id:          id,
		sender:      sender,
		messageType: messageType,
		data:        bytes,
	}
}

func (t MessageType) Bytes() []byte { return utils.ByteToBytes(byte(t)) }

func (m *message) GetId() Id            { return m.id }
func (m *message) GetSender() Id        { return m.sender }
func (m *message) GetType() MessageType { return m.messageType }
func (m *message) GetData() []byte      { return m.data[:len(m.data)-SignatureLength] }
func (m *message) GetSignature() []byte { return m.data[len(m.data)-SignatureLength:] }

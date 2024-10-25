package message

import (
	"bytes"
	"echsylon/fudpucker/entity"
	"echsylon/fudpucker/entity/unit"
	"echsylon/fudpucker/entity/utils"
	"errors"

	"github.com/echsylon/go-log"
)

func NewReceiveMessageHandler(
	saluteOnHail func(string, entity.Message) error,
	updateState func(string, entity.Message) error,
	updateDevice func(string, entity.Message) error,
	updatePeer func(string, entity.Message) error,
	deletePeer func(string, entity.Message) error,
	isMessageInQuarantine func(entity.Id) bool,
	putMessageInQuarantine func(entity.Id, entity.Id),
) func(string, []byte) error {
	return func(sender string, data []byte) (err error) {
		reader := bytes.NewBuffer(data)
		idLen := len(entity.ZeroId)

		bytes := reader.Next(idLen)
		messageId, err := entity.NewBytesId(bytes)
		if err != nil {
			log.Error("Failed reading incomming message")
			return
		}

		typeValue, err := reader.ReadByte()
		messageType := entity.MessageType(typeValue)
		if err != nil {
			log.Error("Failed reading incomming message %s", messageId)
			return
		}

		if isMessageInQuarantine(messageId) {
			log.Notice("Already seen incomming message %s (type=%s), ignoring", messageId, messageType)
			return nil
		}

		bytes = reader.Next(idLen)
		senderId, err := entity.NewBytesId(bytes)
		if err != nil {
			log.Error("Failed reading incomming message %s (type=%s)", messageId, messageType)
			return
		}

		log.Information("Successfully read message %s (type=%s)", messageId, messageType)

		bytes = reader.Next(unit.MaxInt)
		message := entity.NewMessage(messageId, senderId, messageType, bytes)

		putMessageInQuarantine(messageId, senderId)

		switch message.GetType() {
		case entity.MessageTypeCommandHail:
			err = saluteOnHail(sender, message)

		case entity.MessageTypeCommandPatch:
			err = updateState(sender, message)

		case entity.MessageTypeEventSync:
			err = updateDevice(sender, message)

		case entity.MessageTypeEventPeer:
			err = updatePeer(sender, message)

		case entity.MessageTypeEventFarewell:
			err = deletePeer(sender, message)

		default:
			err = errors.New("unexpected message type")
		}

		if err != nil {
			log.Error("Failed handling message %s (type=%s)", messageId, messageType)
		}

		return err
	}
}

func NewSendMessageHandler(
	sendMessage func(string, []byte) error,
	putMessageInQuarantine func(entity.Id, entity.Id),
) func(entity.Peer, entity.Message) error {

	return func(peer entity.Peer, message entity.Message) error {
		writer := bytes.NewBuffer([]byte{})
		writer.Write(message.GetId().Bytes())
		writer.Write(message.GetType().Bytes())
		writer.Write(message.GetSender().Bytes())
		writer.Write(message.GetData())
		writer.Write(message.GetSignature())

		err := sendMessage(peer.GetAddress(), writer.Bytes())
		if err != nil {
			log.Error("Failed sending message %s (type=%s)", message.GetId(), message.GetType())
		} else {
			putMessageInQuarantine(message.GetId(), peer.GetId())
			log.Information("Successfully sent message %s (type=%s)", message.GetId(), message.GetType())
		}

		return err
	}
}

func NewHailMessageProvider(
	getHostId func() (entity.Id, error),
) func() (entity.Message, error) {

	return func() (entity.Message, error) {
		if hostId, err := getHostId(); err != nil {
			return nil, err
		} else if msgId, err := entity.NewRandomId(); err != nil {
			return nil, err
		} else {
			return entity.NewMessage(msgId, hostId, entity.MessageTypeCommandHail, nil), nil
		}
	}
}

func NewPatchMessageProvider(
	getHostId func() (entity.Id, error),
) func(entity.Id, entity.DeviceState) (entity.Message, error) {

	return func(deviceId entity.Id, newState entity.DeviceState) (entity.Message, error) {
		if hostId, err := getHostId(); err != nil {
			return nil, err
		} else if msgId, err := entity.NewRandomId(); err != nil {
			return nil, err
		} else {
			writer := bytes.NewBuffer([]byte{})
			writer.Write(deviceId.Bytes())
			writer.Write(utils.ByteToBytes(byte(newState)))
			return entity.NewMessage(msgId, hostId, entity.MessageTypeCommandPatch, writer.Bytes()), nil
		}
	}
}

func NewPatchMessageReader() func(entity.Message) (entity.Id, entity.DeviceState, error) {
	return func(message entity.Message) (entity.Id, entity.DeviceState, error) {
		reader := bytes.NewBuffer(message.GetData())
		idLen := len(entity.ZeroId)
		idBytes := reader.Next(idLen)
		if deviceId, err := entity.NewBytesId(idBytes); err != nil {
			return entity.ZeroId, entity.DeviceStateOff, err
		} else if value, err := reader.ReadByte(); err != nil {
			return entity.ZeroId, entity.DeviceStateOff, err
		} else {
			return deviceId, entity.DeviceState(value), nil
		}
	}
}

func NewSyncMessageProvider(
	getHostId func() (entity.Id, error),
) func(entity.Device) (entity.Message, error) {

	return func(device entity.Device) (entity.Message, error) {
		if hostId, err := getHostId(); err != nil {
			return nil, err
		} else if msgId, err := entity.NewRandomId(); err != nil {
			return nil, err
		} else {
			writer := bytes.NewBuffer([]byte{})
			writer.Write(device.GetId().Bytes())
			writer.Write(device.GetOwner().Bytes())
			writer.Write(utils.ByteToBytes(byte(device.GetType())))
			writer.Write(utils.ByteToBytes(byte(device.GetState())))
			writer.Write(utils.Int64ToBytes(int64(device.GetVersion())))
			return entity.NewMessage(msgId, hostId, entity.MessageTypeEventSync, writer.Bytes()), nil
		}
	}
}

func NewSyncMessageReader() func(entity.Message) (entity.Device, error) {
	return func(message entity.Message) (entity.Device, error) {
		reader := bytes.NewBuffer(message.GetData())
		idLen := len(entity.ZeroId)
		if deviceId, err := entity.NewBytesId(reader.Next(idLen)); err != nil {
			return nil, err
		} else if ownerId, err := entity.NewBytesId(reader.Next(idLen)); err != nil {
			return nil, err
		} else if typeValue, err := reader.ReadByte(); err != nil {
			return nil, err
		} else if stateValue, err := reader.ReadByte(); err != nil {
			return nil, err
		} else {
			versionBytes := reader.Next(unit.MaxInt)
			deviceType := entity.DeviceType(typeValue)
			deviceState := entity.DeviceState(stateValue)
			version := utils.BytesToInt64(versionBytes)
			device := entity.NewDevice(deviceId, ownerId, deviceType, deviceState, int(version))
			return device, nil
		}
	}
}

func NewPeerMessageProvider(
	getHostId func() (entity.Id, error),
) func(entity.Peer) (entity.Message, error) {

	return func(peer entity.Peer) (entity.Message, error) {
		if hostId, err := getHostId(); err != nil {
			return nil, err
		} else if msgId, err := entity.NewRandomId(); err != nil {
			return nil, err
		} else {
			writer := bytes.NewBuffer([]byte{})
			writer.Write(peer.GetId().Bytes())
			writer.Write([]byte(peer.GetAddress()))
			return entity.NewMessage(msgId, hostId, entity.MessageTypeEventPeer, writer.Bytes()), nil
		}
	}
}

func NewPeerMessageReader() func(entity.Message) (entity.Peer, error) {
	return func(message entity.Message) (entity.Peer, error) {
		reader := bytes.NewBuffer(message.GetData())
		idLen := len(entity.ZeroId)
		idBytes := reader.Next(idLen)
		if deviceId, err := entity.NewBytesId(idBytes); err != nil {
			return nil, err
		} else {
			addressBytes := reader.Next(unit.MaxInt)
			address := string(addressBytes)
			peer := entity.NewPeer(deviceId, address)
			return peer, nil
		}
	}
}

func NewFarewellMessageProvider(
	getHostId func() (entity.Id, error),
) func() (entity.Message, error) {

	return func() (entity.Message, error) {
		if hostId, err := getHostId(); err != nil {
			return nil, err
		} else if msgId, err := entity.NewRandomId(); err != nil {
			return nil, err
		} else {
			return entity.NewMessage(msgId, hostId, entity.MessageTypeEventFarewell, nil), nil
		}
	}
}

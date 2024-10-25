package message

import (
	"echsylon/fudpucker/entity"
	"echsylon/fudpucker/entity/unit"

	"github.com/echsylon/go-log"
)

func NewSaluteOnHailUseCase(
	getDeviceIds func() ([]entity.Id, error),
	getDevice func(entity.Id) (entity.Device, error),
	getPeers func() []entity.Peer,
	rememberPeer func(entity.Peer),
	createSyncMessage func(entity.Device) (entity.Message, error),
	createPeerMessage func(entity.Peer) (entity.Message, error),
	sendMessage func(entity.Peer, entity.Message) error,
) func(string, entity.Message) error {

	return func(senderAddress string, message entity.Message) error {
		peer := entity.NewPeer(message.GetSender(), senderAddress)
		rememberPeer(peer)

		deviceIds, err := getDeviceIds()
		if err != nil {
			deviceIds = []entity.Id{}
		}

		for _, id := range deviceIds {
			if device, err := getDevice(id); err != nil {
				log.Warning("Failed getting device to sync, skipping")
			} else if message, err := createSyncMessage(device); err != nil {
				log.Warning("Failed creating sync message, skipping")
			} else if err := sendMessage(peer, message); err != nil {
				log.Warning("Failed sending sync message, trying next")
			}
		}

		// Don't sync peers for demo
		//peers := getPeers()
		//for _, peer := range peers {
		//	if peer.GetId() == message.GetSender() {
		//		continue
		//	} else if message, err := createPeerMessage(peer); err != nil {
		//		continue
		//	} else if err := sendMessage(senderAddress, message); err != nil {
		//		continue
		//	}
		//}

		return nil
	}
}

func NewSaveStateUseCase(
	readMessage func(entity.Message) (entity.Id, entity.DeviceState, error),
	checkIfOwner func(entity.Id) (bool, error),
	getDevice func(entity.Id) (entity.Device, error),
	saveData func(entity.Id, entity.DeviceState, int) error,
	createSyncMessage func(entity.Device) (entity.Message, error),
	getRandomPeers func(entity.Id, int) ([]entity.Peer, error),
	sendMessage func(entity.Peer, entity.Message) error,
) func(string, entity.Message) error {
	return func(senderAddress string, message entity.Message) error {
		var messageToPropagate entity.Message = message
		var deviceId, newState, err = readMessage(message)
		if err != nil {
			return err
		}

		isOwner, err := checkIfOwner(deviceId)
		if err == nil && isOwner {
			if device, err := getDevice(deviceId); err != nil {
				return err
			} else if err := saveData(deviceId, newState, device.GetVersion()+1); err != nil {
				return err
			} else if updatedDevice, err := getDevice(deviceId); err != nil {
				return err
			} else if messageToPropagate, err = createSyncMessage(updatedDevice); err != nil {
				return err
			}
		}

		peers, err := getRandomPeers(entity.ZeroId, unit.MinInt)
		if err != nil {
			return err
		}

		for _, peer := range peers {
			sendMessage(peer, messageToPropagate)
		}

		return nil
	}
}

func NewSaveDeviceUseCase(
	checkIfOwner func(entity.Id) (bool, error),
	checkVersion func(entity.Id, int) (bool, error),
	readCandidate func(entity.Message) (entity.Device, error),
	saveCandidate func(entity.Device) error,
	getDevice func(entity.Id) (entity.Device, error),
	getRandomPeers func(entity.Id, int) ([]entity.Peer, error),
	createSyncMessage func(entity.Device) (entity.Message, error),
	sendMessage func(entity.Peer, entity.Message) error,
) func(string, entity.Message) error {

	return func(senderAddress string, message entity.Message) error {
		candidate, err := readCandidate(message)
		if err != nil {
			log.Error("Failed to read device state from message")
			return err
		}

		deviceId := candidate.GetId()
		isOwner, err := checkIfOwner(deviceId)
		if err == nil && isOwner {
			return nil // Only we can push this device's state.
		}

		messageToPropagate := message
		isNewer, err := checkVersion(deviceId, candidate.GetVersion())
		if err == nil && !isNewer {
			// Our device state is newer than the one provided in the incoming
			// message. Get our data and propagate that one instead. If we fail
			// to construct a new message, we rather bail out than propagate
			// outdated information.
			if device, err := getDevice(deviceId); err != nil {
				return nil
			} else if msg, err := createSyncMessage(device); err != nil {
				return nil
			} else {
				messageToPropagate = msg
			}
		} else if err := saveCandidate(candidate); err != nil {
			log.Warning("Failed to save new device state, ignoring")
			return err
		}

		peers, err := getRandomPeers(messageToPropagate.GetId(), unit.MinInt)
		if err != nil {
			log.Warning("Failed to select peer pool, ignoring")
			return err
		}

		for _, peer := range peers {
			sendMessage(peer, messageToPropagate)
		}

		return nil
	}
}

func NewSavePeerUseCase(
	readPeer func(entity.Message) (entity.Peer, error),
	savePeer func(entity.Peer),
) func(string, entity.Message) error {

	return func(senderAddress string, message entity.Message) error {
		// Don't propagate this message as our peers could then
		// be fooled into thinking our IP address is associated
		// with the peer being described in this message.
		sender := entity.NewPeer(message.GetSender(), senderAddress)
		savePeer(sender)

		if peer, err := readPeer(message); err != nil {
			return err
		} else {
			savePeer(peer)
			return nil
		}
	}
}

func NewDeletePeerUseCase(
	deletePeer func(entity.Id),
) func(string, entity.Message) error {
	return func(senderAddress string, message entity.Message) error {
		// Don't propagate this message as our peers could then
		// be fooled into thinking our IP address is associated
		// with the peer being described in this message.
		deletePeer(message.GetSender())
		return nil
	}
}

func NewSendHailCommandUseCase(
	getRandomPeers func(entity.Id, int) ([]entity.Peer, error),
	createHailMessage func() (entity.Message, error),
	getDeviceIds func() ([]entity.Id, error),
	getDevice func(entity.Id) (entity.Device, error),
	createSyncMessage func(entity.Device) (entity.Message, error),
	sendMessage func(entity.Peer, entity.Message) error,
) func() error {

	return func() error {
		message, err := createHailMessage()
		if err != nil {
			return err
		}

		peers, err := getRandomPeers(message.GetId(), unit.MinInt)
		if err != nil {
			return err
		}

		deviceIds, err := getDeviceIds()
		if err != nil {
			deviceIds = []entity.Id{}
		}

		messages := make([]entity.Message, 0)
		for _, id := range deviceIds {
			if device, err := getDevice(id); err == nil {
				if message, err := createSyncMessage(device); err == nil {
					messages = append(messages, message)
				}
			}
		}

		for _, peer := range peers {
			sendMessage(peer, message)
			for _, message := range messages {
				sendMessage(peer, message)
			}
		}

		return nil
	}
}

func NewSendFarewellEventUseCase(
	getRandomPeers func(entity.Id, int) ([]entity.Peer, error),
	createFarewellMessage func() (entity.Message, error),
	sendMessage func(entity.Peer, entity.Message) error,
) func() error {

	return func() error {
		message, err := createFarewellMessage()
		if err != nil {
			return err
		}

		peers, err := getRandomPeers(message.GetId(), unit.MaxInt)
		if err != nil {
			return err
		}

		for _, peer := range peers {
			sendMessage(peer, message)
		}

		return nil
	}
}

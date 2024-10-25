package data

import (
	"echsylon/fudpucker/entity"
	"echsylon/fudpucker/entity/unit"
	"errors"
)

func NewCreateDeviceUseCase(
	getHostId func() (entity.Id, error),
	saveDevice func(entity.Device) error,
) func(entity.DeviceType, entity.DeviceState) (entity.Id, error) {

	return func(deviceType entity.DeviceType, deviceState entity.DeviceState) (entity.Id, error) {
		if hostId, err := getHostId(); err != nil {
			return entity.ZeroId, err
		} else if deviceId, err := entity.NewRandomId(); err != nil {
			return entity.ZeroId, err
		} else if device := entity.NewDevice(deviceId, hostId, deviceType, deviceState, 0); device == nil {
			return entity.ZeroId, errors.New("error create device")
		} else if err := saveDevice(device); err != nil {
			return entity.ZeroId, err
		} else {
			return deviceId, nil
		}
	}
}

func NewCheckIfOwnerUseCase(
	getHostId func() (entity.Id, error),
	getOwnerAttribute func(entity.Id) (entity.Id, error),
) func(entity.Id) (bool, error) {

	return func(deviceId entity.Id) (bool, error) {
		if host, err := getHostId(); err != nil {
			return false, err
		} else if owner, err := getOwnerAttribute(deviceId); err != nil {
			return false, err
		} else {
			return owner == host, nil
		}
	}
}

func NewCheckIfNewerUseCase(
	getVersionAttribute func(entity.Id) (int, error),
) func(entity.Id, int) (bool, error) {

	return func(deviceId entity.Id, version int) (bool, error) {
		if currentVersion, err := getVersionAttribute(deviceId); err != nil {
			return false, err
		} else {
			return version > currentVersion, nil
		}
	}
}

func NewGetHostInfoUseCase(
	getHostId func() (entity.Id, error),
	getLocalAddress func() (string, error),
	getBroadcastAddress func() (string, error),
) func() (entity.Host, error) {

	return func() (entity.Host, error) {
		if id, err := getHostId(); err != nil {
			return nil, err
		} else if address, err := getLocalAddress(); err != nil {
			return nil, err
		} else if broadcast, err := getBroadcastAddress(); err != nil {
			return nil, err
		} else {
			return entity.NewHost(id, address, broadcast), err
		}
	}
}

func NewPatchStateUseCase(
	checkIfOwner func(entity.Id) (bool, error),
	getDevice func(entity.Id) (entity.Device, error),
	saveData func(entity.Id, entity.DeviceState, int) error,
	createSyncMessage func(entity.Device) (entity.Message, error),
	createPatchMessage func(entity.Id, entity.DeviceState) (entity.Message, error),
	getRandomPeers func(entity.Id, int) ([]entity.Peer, error),
	sendMessage func(entity.Peer, entity.Message) error,
) func(entity.Id, entity.DeviceState) error {

	return func(deviceId entity.Id, newState entity.DeviceState) error {
		var message entity.Message = nil
		var isOwner, err = checkIfOwner(deviceId)
		if err != nil || !isOwner {
			// Not our device. Create a message requesting the owner to update it.
			if message, err = createPatchMessage(deviceId, newState); err != nil {
				return err
			}
		} else {
			// This is our device. Update it and create a sync message.
			if device, err := getDevice(deviceId); err != nil {
				return err
			} else if err := saveData(deviceId, newState, device.GetVersion()+1); err != nil {
				return err
			} else if updatedDevice, err := getDevice(deviceId); err != nil {
				return err
			} else if message, err = createSyncMessage(updatedDevice); err != nil {
				return err
			}
		}

		// Propagate the message (whichever it is) to a random set of peers.
		peers, err := getRandomPeers(entity.ZeroId, unit.MinInt)
		if err != nil {
			return err
		}

		for _, peer := range peers {
			sendMessage(peer, message)
		}

		return nil
	}
}

const defaultRandomPeerCount = 5

func NewRandomSafePeersForMessageUseCase(
	getHostId func() (entity.Id, error),
	getBroadcastAddress func() (string, error),
	getAllPeers func() []entity.Peer,
	isPeerInQuarantine func(entity.Id, entity.Id) bool,
) func(entity.Id, int) ([]entity.Peer, error) {

	return func(messageId entity.Id, peerCount int) ([]entity.Peer, error) {
		// If the network supports multicast UDP protocol, and we can
		// successfully get the broadcast address, then return it alone.
		if address, err := getBroadcastAddress(); err != nil && address != "" {
			if hostId, err := getHostId(); err != nil {
				return []entity.Peer{entity.NewPeer(hostId, address)}, nil
			}
		}

		// We're (ab)using the map implementation in Go here. The runtime
		// will take active measures to distort the insertion order when
		// iterating over the keys, hence they will be returned in random
		// order - which is exactly what we need. We are, arguably, doing
		// exactly what the maintainer didn't want the developers to do:
		// rely on the specific implementation details. Nice!
		peers := getAllPeers()
		safePeers := make(map[entity.Peer]any)
		for _, peer := range peers {
			if !isPeerInQuarantine(messageId, peer.GetId()) {
				safePeers[peer] = struct{}{}
			}
		}

		safePeerCount := len(safePeers)
		resultCount := peerCount

		if peerCount <= 0 {
			resultCount = defaultRandomPeerCount
		}

		if safePeerCount < resultCount {
			resultCount = safePeerCount
		}

		result := make([]entity.Peer, resultCount)
		index := 0

		for peer := range safePeers {
			if index == resultCount {
				break
			} else {
				result[index] = peer
				index++
			}
		}

		return result, nil
	}
}

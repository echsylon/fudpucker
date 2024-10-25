package request

import (
	"echsylon/fudpucker/entity"
	"encoding/json"
	"strconv"
)

func NewGetHostInfoRequestHandler(
	getHostInfo func() (entity.Host, error),
) func(map[string][]string, []byte) ([]byte, int) {

	return func(args map[string][]string, content []byte) ([]byte /*result*/, int /*status*/) {
		if host, err := getHostInfo(); err != nil {
			return nil, 500
		} else if json, err := hostToJson(host); err != nil {
			return nil, 500
		} else {
			return json, 200
		}
	}
}

func NewGetApiRequestHandler(
	getApiDocumentation func() (map[string]string, error),
) func(map[string][]string, []byte) ([]byte, int) {

	return func(args map[string][]string, content []byte) ([]byte /*result*/, int /*status*/) {
		if doc, err := getApiDocumentation(); err != nil {
			return nil, 500
		} else if json, err := docToJson(doc); err != nil {
			return nil, 500
		} else {
			return json, 200
		}
	}
}

func NewGetDeviceIdsRequestHandler(
	getDeviceIds func() ([]entity.Id, error),
) func(map[string][]string, []byte) ([]byte, int) {

	return func(args map[string][]string, content []byte) ([]byte /*result*/, int /*status*/) {
		if ids, err := getDeviceIds(); err != nil {
			return nil, 500
		} else if json, err := idsToJson(ids); err != nil {
			return nil, 500
		} else {
			return json, 200
		}
	}
}

func NewGetDeviceRequestHandler(
	getDevice func(entity.Id) (entity.Device, error),
) func(map[string][]string, []byte) ([]byte, int) {

	return func(args map[string][]string, content []byte) ([]byte /*result*/, int /*status*/) {
		if ids, ok := args["id"]; !ok {
			return nil, 400
		} else if device, err := getDevice(entity.NewStringId(ids[0])); err != nil {
			return nil, 404
		} else if json, err := deviceToJson(device); err != nil {
			return nil, 500
		} else {
			return json, 200
		}
	}
}

func NewPatchStateRequestHandler(
	updateState func(entity.Id, entity.DeviceState) error,
) func(map[string][]string, []byte) ([]byte, int) {

	return func(args map[string][]string, content []byte) ([]byte /*result*/, int /*status*/) {
		var deviceId entity.Id
		if ids, ok := args["id"]; !ok {
			return nil, 400
		} else {
			deviceId = entity.NewStringId(ids[0])
		}

		var deviceState entity.DeviceState
		if values, ok := args["state"]; !ok || len(values) == 0 {
			deviceState = entity.DeviceStateOff
		} else if value, err := strconv.Atoi(values[0]); err != nil {
			return nil, 400
		} else {
			deviceState = entity.DeviceState(value)
		}

		if err := updateState(deviceId, deviceState); err != nil {
			return nil, 500
		} else {
			return nil, 200
		}
	}
}

func NewCreateDeviceRequestHandler(
	createDevice func(entity.DeviceType, entity.DeviceState) (entity.Id, error),
) func(map[string][]string, []byte) ([]byte, int) {

	return func(args map[string][]string, content []byte) ([]byte /*result*/, int /*status*/) {
		var deviceType entity.DeviceType
		var deviceState entity.DeviceState

		if values, ok := args["type"]; !ok || len(values) == 0 {
			return nil, 400
		} else if value, err := strconv.Atoi(values[0]); err != nil {
			return nil, 400
		} else {
			deviceType = entity.DeviceType(value)
		}

		if values, ok := args["state"]; !ok || len(values) == 0 {
			deviceState = entity.DeviceStateOff
		} else if value, err := strconv.Atoi(values[0]); err != nil {
			return nil, 400
		} else {
			deviceState = entity.DeviceState(value)
		}

		if id, err := createDevice(deviceType, deviceState); err != nil {
			return nil, 500
		} else if json, err := idToJson(id); err != nil {
			return nil, 500
		} else {
			return json, 200
		}
	}
}

func NewDeleteDeviceRequestHandler(
	deleteDevice func(entity.Id) (bool, error),
) func(map[string][]string, []byte) ([]byte, int) {

	return func(args map[string][]string, content []byte) ([]byte /*result*/, int /*status*/) {
		if values, ok := args["id"]; !ok || len(values) == 0 {
			return nil, 400
		} else if deleted, err := deleteDevice(entity.NewStringId(values[0])); err != nil {
			return nil, 404 // not found
		} else if !deleted {
			return nil, 403 // forbidden
		} else {
			return nil, 200
		}
	}
}

func NewGetPeersRequestHandler(
	getPeers func() []entity.Peer,
) func(map[string][]string, []byte) ([]byte, int) {

	return func(args map[string][]string, content []byte) ([]byte /*result*/, int /*status*/) {
		peers := getPeers()
		if json, err := peersToJson(peers); err != nil {
			return nil, 500
		} else {
			return json, 200
		}
	}
}

func NewAddPeerRequestHandler(
	addPeer func(entity.Peer),
) func(map[string][]string, []byte) ([]byte, int) {

	return func(args map[string][]string, content []byte) ([]byte /*result*/, int /*status*/) {
		if ids, ok := args["id"]; !ok || len(ids) == 0 {
			return nil, 400
		} else if addresses, ok := args["address"]; !ok || len(addresses) == 0 {
			return nil, 400
		} else if peer := entity.NewPeer(entity.NewStringId(ids[0]), addresses[0]); peer == nil {
			return nil, 500
		} else {
			addPeer(peer)
			return nil, 200
		}
	}
}

func NewJoinNetworkRequestHandler(
	joinNetwork func() error,
) func(map[string][]string, []byte) ([]byte, int) {

	return func(args map[string][]string, content []byte) ([]byte /*result*/, int /*status*/) {
		if err := joinNetwork(); err != nil {
			return nil, 500
		} else {
			return nil, 202
		}
	}
}

func NewLeaveNetworkRequestHandler(
	leaveNetwork func() error,
) func(map[string][]string, []byte) ([]byte, int) {

	return func(args map[string][]string, content []byte) ([]byte /*result*/, int /*status*/) {
		if err := leaveNetwork(); err != nil {
			return nil, 500
		} else {
			return nil, 202
		}
	}
}

func NewShutdownRequestHandler(
	shutdown func() error,
) func(map[string][]string, []byte) ([]byte, int) {

	return func(args map[string][]string, content []byte) ([]byte /*result*/, int /*status*/) {
		if err := shutdown(); err != nil {
			return nil, 500
		} else {
			return nil, 200
		}
	}
}

// Helper functions
func idToJson(id entity.Id) ([]byte, error) {
	data := make(map[string]string)
	data["id"] = id.String()
	return json.Marshal(data)
}

func idsToJson(ids []entity.Id) ([]byte, error) {
	data := make([]string, len(ids))
	for index, id := range ids {
		data[index] = id.String()
	}
	return json.Marshal(data)
}

func docToJson(doc map[string]string) ([]byte, error) {
	return json.Marshal(doc)
}

func hostToJson(host entity.Host) ([]byte, error) {
	data := make(map[string]any)
	data["id"] = host.GetId().String()
	data["address"] = host.GetLocalAddress()
	//data["broadcastAddress"] = host.GetBroadcastAddress()
	return json.Marshal(data)
}

func deviceToJson(device entity.Device) ([]byte, error) {
	data := make(map[string]any)
	data["id"] = device.GetId().String()
	data["owner"] = device.GetOwner().String()
	data["type"] = device.GetType()
	data["state"] = device.GetState()
	data["version"] = device.GetVersion()
	return json.Marshal(data)
}

func peersToJson(peers []entity.Peer) ([]byte, error) {
	data := make(map[string]string)
	for _, peer := range peers {
		data[peer.GetId().String()] = peer.GetAddress()
	}
	return json.Marshal(data)
}

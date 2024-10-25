package data

import (
	"echsylon/fudpucker/entity"
	"echsylon/fudpucker/entity/utils"
	"errors"
)

func NewGetDeviceIdsDataAdapter(
	getData func(entity.Id, entity.Id) (map[entity.Id][]byte, error),
) func() ([]entity.Id, error) {

	return func() ([]entity.Id, error) {
		if data, err := getData(entity.ZeroId, entity.ZeroId); err != nil {
			return nil, err
		} else {
			result := make([]entity.Id, len(data))
			index := 0
			for id := range data {
				result[index] = id
				index++
			}
			return result, nil
		}
	}
}

func NewGetDeviceOwnerAttributeAdapter(
	getData func(entity.Id, entity.Id) (map[entity.Id][]byte, error),
) func(entity.Id) (entity.Id, error) {

	return func(deviceId entity.Id) (entity.Id, error) {
		ownerAttr := entity.NewStringId("owner")
		data, err := getData(deviceId, ownerAttr)
		if err != nil {
			return entity.ZeroId, err
		} else {
			return entity.NewBytesId(data[ownerAttr])
		}
	}
}

func NewGetStateVersionAttributeAdapter(
	getData func(entity.Id, entity.Id) (map[entity.Id][]byte, error),
) func(entity.Id) (int, error) {

	return func(deviceId entity.Id) (int, error) {
		versionAttr := entity.NewStringId("version")
		if data, err := getData(deviceId, versionAttr); err != nil {
			return 0, err
		} else {
			value := utils.BytesToInt64(data[versionAttr])
			return int(value), nil
		}
	}
}

func NewGetDeviceDataAdapter(
	getData func(entity.Id, entity.Id) (map[entity.Id][]byte, error),
) func(entity.Id) (entity.Device, error) {

	return func(deviceId entity.Id) (entity.Device, error) {
		if data, err := getData(deviceId, entity.ZeroId); err != nil {
			return nil, err
		} else {
			typeAttr := entity.NewStringId("type")
			ownerAttr := entity.NewStringId("owner")
			stateAttr := entity.NewStringId("state")
			versionAttr := entity.NewStringId("version")

			var ownerId entity.Id = entity.ZeroId
			var deviceType entity.DeviceType = entity.DeviceTypeUnknown
			var deviceState entity.DeviceState = entity.DeviceStateOff
			var stateVersion int = 0

			for attr := range data {
				if attr == ownerAttr {
					if ownerId, err = entity.NewBytesId(data[attr]); err != nil {
						ownerId = entity.ZeroId
					}
				} else if attr == typeAttr {
					value := utils.BytesToByte(data[attr])
					deviceType = entity.DeviceType(value)
				} else if attr == stateAttr {
					value := utils.BytesToByte(data[attr])
					deviceState = entity.DeviceState(value)
				} else if attr == versionAttr {
					value := utils.BytesToInt64(data[attr])
					stateVersion = int(value)
				}
			}

			if deviceType == entity.DeviceTypeUnknown {
				return nil, errors.New("data format error")
			} else {
				return entity.NewDevice(deviceId, ownerId, deviceType, deviceState, stateVersion), nil
			}
		}
	}
}

func NewCreateDeviceDataAdapter(
	saveData func(entity.Id, map[entity.Id][]byte) error,
) func(entity.Device) error {

	return func(device entity.Device) error {
		data := make(map[entity.Id][]byte)
		data[entity.NewStringId("type")] = utils.ByteToBytes(byte(device.GetType()))
		data[entity.NewStringId("state")] = utils.ByteToBytes(byte(device.GetState()))
		data[entity.NewStringId("version")] = utils.Int64ToBytes(int64(device.GetVersion()))
		data[entity.NewStringId("owner")] = device.GetOwner().Bytes()

		return saveData(device.GetId(), data)
	}
}

func NewPatchStateDataAdapter(
	saveData func(entity.Id, map[entity.Id][]byte) error,
) func(entity.Id, entity.DeviceState, int) error {

	return func(deviceId entity.Id, newState entity.DeviceState, newVersion int) error {
		data := make(map[entity.Id][]byte)
		data[entity.NewStringId("state")] = utils.ByteToBytes(byte(newState))
		data[entity.NewStringId("version")] = utils.Int64ToBytes(int64(newVersion))
		return saveData(deviceId, data)
	}
}

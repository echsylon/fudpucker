package request

import (
	"context"
	"echsylon/fudpucker/entity"
)

func NewGetApiDocUseCase() func() (map[string]string, error) {

	return func() (map[string]string, error) {
		data := make(map[string]string)
		data["DELETE /device/{id}"] = "Delete a device previously created by you."
		data["DELETE /network"] = "Leave the network, stop syncing state."
		data["GET /"] = "This resource"
		data["GET /device"] = "Get all devices your peer currently knows about."
		data["GET /device/{id}"] = "Get the last synched state for the given device."
		data["GET /info"] = "Display your peer info."
		data["GET /peer"] = "Get all peers you currently see."
		data["POST /device"] = "Create a new device, params: \"type\"=1 (light), \"state\"=[0|1] (off/on)"
		data["POST /network"] = "Join the network, start syncing state."
		data["POST /peer"] = "Manually add a new peer (needed in networks not supporting multicast)."
		data["POST /shutdown"] = "Shut down and exit the application."
		return data, nil
	}

}

func NewShutdownUseCase(
	clearPeerCache func(),
	clearMessageCache func(),
	shutDownServices context.CancelFunc,
) func() error {
	return func() error {
		clearPeerCache()
		clearMessageCache()
		shutDownServices()
		return nil
	}
}

func NewDeleteDeviceUseCase(
	checkIfOwner func(entity.Id) (bool, error),
	deleteDevice func(entity.Id) error,
) func(entity.Id) (bool, error) {
	return func(deviceId entity.Id) (bool, error) {
		if isOwner, err := checkIfOwner(deviceId); err != nil {
			return false, err
		} else if isOwner {
			return false, nil // We can only delete our devices
		} else if err := deleteDevice(deviceId); err != nil {
			return false, err
		} else {
			return true, nil
		}
	}
}

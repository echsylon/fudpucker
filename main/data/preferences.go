package data

import (
	"echsylon/fudpucker/entity"
	"echsylon/fudpucker/entity/utils"
	"fmt"
	"net"

	"github.com/denisbrodbeck/machineid"
	"github.com/google/uuid"
)

type Preferences interface {
	GetHostId() (entity.Id, error)
	GetBroadcastAddress() (string, error)
	GetLocalAddress() (string, error)
}

type preferences struct {
	cachedBroadcastAddress string
	cachedLocalAddress     string
	cachedMachineId        entity.Id
	httpPort               int
	udpPort                int
}

func NewPreferences(requestPort int, messagePort int) Preferences {
	return &preferences{
		httpPort: requestPort,
		udpPort:  messagePort,
	}
}

func (r *preferences) GetHostId() (entity.Id, error) {
	if r.cachedMachineId != entity.ZeroId {
		return r.cachedMachineId, nil
	} else if text, err := machineid.ID(); err != nil {
		return entity.ZeroId, err
	} else if err := uuid.Validate(text); err != nil {
		return entity.ZeroId, err
	} else {
		// Don't expose the local machine id in it's raw form,
		// do at least a minimal SHA1 hash on it.
		space := uuid.NameSpaceURL
		bytes := utils.StringToBytes(text)
		source := uuid.NewSHA1(space, bytes)
		r.cachedMachineId = entity.Id(source)
		return r.cachedMachineId, nil
	}
}

func (r *preferences) GetBroadcastAddress() (string, error) {
	if r.cachedBroadcastAddress == "" {
		if broadcast, local, err := findLocalUdpAddresses(); err != nil {
			return "", err
		} else {
			r.cachedBroadcastAddress = fmt.Sprintf("%s:%d", broadcast, r.udpPort)
			r.cachedLocalAddress = fmt.Sprintf("%s:%d", local, r.udpPort)
		}
	}
	return r.cachedBroadcastAddress, nil
}

func (r *preferences) GetLocalAddress() (string, error) {
	if r.cachedLocalAddress == "" {
		broadcast, local, err := findLocalUdpAddresses()
		if err != nil {
			return "", err
		} else {
			r.cachedBroadcastAddress = fmt.Sprintf("%s:%d", broadcast, r.udpPort)
			r.cachedLocalAddress = fmt.Sprintf("%s:%d", local, r.udpPort)
		}
	}
	return r.cachedLocalAddress, nil
}

func findLocalUdpAddresses() (broadcast string, local string, err error) {
	var interfaceAddresses []net.Addr
	if interfaceAddresses, err = net.InterfaceAddrs(); err != nil {
		return
	}

	for _, address := range interfaceAddresses {
		if ip, ok := address.(*net.IPNet); !ok {
			continue
		} else if ip4 := ip.IP.To4(); ip4 == nil {
			continue
		} else if ip4.IsLoopback() {
			continue
		} else if ip4.IsMulticast() && broadcast == "" {
			broadcast = ip4.String()
		} else if ip4.IsPrivate() && local == "" {
			local = ip4.String()
		}

		if broadcast != "" && local != "" {
			break
		}
	}

	return
}

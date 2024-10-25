package entity

type DeviceType byte
type DeviceState byte

const (
	DeviceTypeUnknown DeviceType = 0xf
	DeviceTypeLight              = iota
)

const (
	DeviceStateOff DeviceState = iota
	DeviceStateOn
)

type Device interface {
	GetId() Id
	GetOwner() Id
	GetType() DeviceType
	GetState() DeviceState
	GetVersion() int
}

type device struct {
	id           Id
	owner        Id
	deviceType   DeviceType
	deviceState  DeviceState
	stateVersion int
}

func NewDevice(id Id, owner Id, deviceType DeviceType, deviceState DeviceState, version int) Device {
	return &device{
		id:           id,
		deviceType:   deviceType,
		deviceState:  deviceState,
		stateVersion: version,
		owner:        owner,
	}
}

func (d *device) GetId() Id             { return d.id }
func (d *device) GetOwner() Id          { return d.owner }
func (d *device) GetType() DeviceType   { return d.deviceType }
func (d *device) GetState() DeviceState { return d.deviceState }
func (d *device) GetVersion() int       { return d.stateVersion }

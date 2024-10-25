package entity

type Event interface {
	GetId() Id
	GetDeviceId() Id
	GetState() byte
}

type event struct {
	id       Id
	deviceId Id
	state    byte
}

func NewOnEvent(id Id, device Id) Event {
	return &event{
		id:       id,
		deviceId: device,
		state:    byte(1),
	}
}

func NewOffEvent(id Id, device Id) Event {
	return &event{
		id:       id,
		deviceId: device,
		state:    byte(0),
	}
}

func (e *event) GetId() Id       { return e.id }
func (e *event) GetDeviceId() Id { return e.deviceId }
func (e *event) GetState() byte  { return e.state }

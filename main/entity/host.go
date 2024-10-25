package entity

type Host interface {
	GetId() Id
	GetLocalAddress() string
	GetBroadcastAddress() string
}

type host struct {
	id        Id
	address   string
	broadcast string
}

func NewHost(id Id, address string, broadcast string) Host {
	return &host{
		id:        id,
		address:   address,
		broadcast: broadcast,
	}
}

func (h *host) GetId() Id                   { return h.id }
func (h *host) GetLocalAddress() string     { return h.address }
func (h *host) GetBroadcastAddress() string { return h.broadcast }

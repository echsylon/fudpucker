package entity

type Peer interface {
	GetId() Id
	GetAddress() string
}

type peer struct {
	id      Id
	address string
}

func NewPeer(id Id, address string) Peer {
	return &peer{
		id:      id,
		address: address,
	}
}

func (p *peer) GetId() Id          { return p.id }
func (p *peer) GetAddress() string { return p.address }

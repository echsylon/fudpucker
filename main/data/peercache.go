package data

import (
	"echsylon/fudpucker/entity"
	"errors"
)

type PeerCache interface {
	AddPeer(entity.Peer)
	GetPeer(entity.Id) (entity.Peer, error)
	RemovePeer(entity.Id)
	GetAllPeers() []entity.Peer
	GetRandomPeers(int) []entity.Peer
	Reset()
}

type peerCache struct {
	peers map[entity.Id]entity.Peer
}

const (
	recommendedPeerCount = 5
)

func NewPeerCache() PeerCache {
	return &peerCache{peers: make(map[entity.Id]entity.Peer)}
}

func (r *peerCache) AddPeer(peer entity.Peer) {
	r.peers[peer.GetId()] = peer
}

func (r *peerCache) GetPeer(id entity.Id) (entity.Peer, error) {
	if peer, ok := r.peers[id]; !ok {
		return nil, errors.New("no such peer error")
	} else {
		return peer, nil
	}
}

func (r *peerCache) RemovePeer(id entity.Id) {
	delete(r.peers, id)
}

func (r *peerCache) GetAllPeers() []entity.Peer {
	result := make([]entity.Peer, len(r.peers))
	index := 0
	for _, peer := range r.peers {
		result[index] = peer
		index++
	}
	return result
}

func (r *peerCache) GetRandomPeers(count int) []entity.Peer {
	peerCount := len(r.peers)
	resultCount := count

	if count <= 0 {
		resultCount = recommendedPeerCount
	}

	if peerCount < resultCount {
		count = resultCount
	}

	result := make([]entity.Peer, resultCount)
	index := 0

	// Map key order is guaranteed to be random when iterating.
	for _, peer := range r.peers {
		if index == resultCount {
			break
		} else {
			result[index] = peer
			index++
		}
	}

	return result
}

func (r *peerCache) Reset() {
	clear(r.peers)
}

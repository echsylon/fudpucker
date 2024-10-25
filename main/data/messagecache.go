package data

import (
	"echsylon/fudpucker/entity"
	"time"
)

type MessageCache interface {
	ContainsMessageForPeer(entity.Id, entity.Id) bool
	ContainsMessage(entity.Id) bool
	Hold(entity.Id, entity.Id)
	Reset()
}

type entry struct {
	receivers map[entity.Id]time.Time
}

type messageCache struct {
	entries map[entity.Id]entry
}

const timeToLive = 10 * time.Second

func NewMessageCache() MessageCache {
	return &messageCache{entries: make(map[entity.Id]entry)}
}

func (c *messageCache) ContainsMessageForPeer(messageId entity.Id, peerId entity.Id) bool {
	c.clearOutdated()
	if entry, hasMessage := c.entries[messageId]; !hasMessage {
		return false
	} else if deadline, hasPeer := entry.receivers[peerId]; !hasPeer {
		return false
	} else {
		return deadline.After(time.Now())
	}
}

func (c *messageCache) ContainsMessage(messageId entity.Id) bool {
	c.clearOutdated()
	_, hasMessage := c.entries[messageId]
	return hasMessage
}

func (c *messageCache) Hold(messageId entity.Id, peerId entity.Id) {
	if item, hasMessage := c.entries[messageId]; !hasMessage {
		d := time.Now().Add(timeToLive)
		r := map[entity.Id]time.Time{peerId: d}
		e := entry{receivers: r}
		c.entries[messageId] = e
	} else if _, hasPeer := item.receivers[peerId]; !hasPeer {
		d := time.Now().Add(timeToLive)
		item.receivers[peerId] = d
	}
}

func (c *messageCache) Reset() {
	clear(c.entries)
}

func (c *messageCache) clearOutdated() {
	now := time.Now()
	for messageId, entry := range c.entries {
		for peerId, deadline := range entry.receivers {
			if deadline.Before(now) {
				delete(entry.receivers, peerId)
			}
		}
		if len(entry.receivers) == 0 {
			delete(c.entries, messageId)
		}
	}
}

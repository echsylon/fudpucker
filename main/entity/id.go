package entity

import (
	"errors"

	"github.com/google/uuid"
)

// Generic id specification
var ZeroId Id = Id(uuid.Nil)

type Id [16]byte

func NewRandomId() (Id, error) {
	if source, err := uuid.NewRandom(); err != nil {
		return ZeroId, err
	} else {
		return Id(source), nil
	}
}

func NewBytesId(bytes []byte) (Id, error) {
	if len(bytes) != 16 {
		return ZeroId, errors.New("unexpected data lengh")
	} else {
		return Id(bytes), nil
	}
}

func NewStringId(text string) Id {
	var source uuid.UUID
	if err := uuid.Validate(text); err != nil {
		source = uuid.NewSHA1(uuid.NameSpaceURL, []byte(text))
	} else {
		source = uuid.MustParse(text)
	}
	return Id(source)
}

func (i Id) String() string { return uuid.UUID(i).String() }
func (i Id) Bytes() []byte  { return i[:] }

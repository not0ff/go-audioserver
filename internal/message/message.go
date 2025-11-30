package message

import (
	"encoding/json"
)

type Message struct {
	Action  int    `json:"action"`
	Payload []byte `json:"payload"`
}

type Payload any

type PlayPayload struct {
	Id     int
	Format string
	Path   string
	Data   []byte
	Volume int
	Loop   bool
}

type IdPayload struct {
	Id int
}

func NewMessage() *Message {
	return &Message{}
}

func (m *Message) SetPayload(p Payload) error {
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	m.Payload = b
	return nil
}

func (m *Message) Decode(d []byte) error {
	if err := json.Unmarshal(d, m); err != nil {
		return err
	}
	return nil
}

func (m *Message) Encode() ([]byte, error) {
	d, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return d, nil
}

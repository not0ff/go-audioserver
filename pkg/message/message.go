package message

import (
	"encoding/json"
	"fmt"
)

type Message struct {
	Action  int    `json:"action,omitempty"`
	Payload []byte `json:"payload,omitempty"`
}

type Payload any

type PlayPayload struct {
	Id     int    `json:"id,omitempty"`
	Format string `json:"format,omitempty"`
	Path   string `json:"path,omitempty"`
	Data   []byte `json:"data,omitempty"`
	Volume int    `json:"volume,omitempty"`
	Loop   bool   `json:"loop,omitempty"`
}

func (p *PlayPayload) String() string {
	return fmt.Sprintf("{Id:%v Format:%s Path:%s Data:%v Volume:%v Loop:%v}", p.Id, p.Format, p.Path, len(p.Data), p.Volume, p.Loop)
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

package message

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

type Message struct {
	lenBuf [4]byte
	Data   []byte
}

func NewMessage() *Message {
	return &Message{}
}

func (m *Message) String() string {
	return fmt.Sprintf("Message{ [Length: %v][Data: %s] }", m.Len(), m.Data)
}

func (m *Message) Len() int {
	return len(m.Data)
}

func (m *Message) SetData(d []byte) {
	m.Data = d
}

func (m *Message) Write(w io.Writer) error {
	msgLen := m.Len()
	if msgLen > math.MaxUint32 {
		return errors.New("data size too big (exceeds uint32 length limit)")
	}

	binary.BigEndian.PutUint32(m.lenBuf[:], uint32(msgLen))

	if _, err := w.Write(m.lenBuf[:]); err != nil {
		return err
	}
	if _, err := w.Write(m.Data); err != nil {
		return err
	}

	return nil
}

func (m *Message) Read(r io.Reader) error {
	if _, err := io.ReadFull(r, m.lenBuf[:]); err != nil {
		return err
	}

	msgLen := int(binary.BigEndian.Uint32(m.lenBuf[:]))
	m.Data = make([]byte, msgLen)

	if _, err := io.ReadFull(r, m.Data); err != nil {
		return err
	}

	return nil
}

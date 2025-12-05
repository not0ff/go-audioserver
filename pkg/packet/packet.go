package packet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

type Packet struct {
	lenBuf [4]byte
	Data   []byte
}

func NewPacket() *Packet {
	return &Packet{}
}

func (p *Packet) String() string {
	return fmt.Sprintf("Packet{ [Length: %v][Data: %s] }", p.Len(), p.Data)
}

func (p *Packet) Len() int {
	return len(p.Data)
}

func (p *Packet) SetData(d []byte) {
	p.Data = d
}

// Writes encoded packet to provided writer
func (p *Packet) Write(w io.Writer) error {
	// Read and check data length
	msgLen := p.Len()
	if msgLen > math.MaxUint32 {
		return errors.New("data size too big (exceeds uint32 length limit)")
	}

	// Convert length to byte sequence
	binary.BigEndian.PutUint32(p.lenBuf[:], uint32(msgLen))

	// Write 4-byte length-prefix
	if _, err := w.Write(p.lenBuf[:]); err != nil {
		return err
	}
	// Write remaining bytes
	if _, err := w.Write(p.Data); err != nil {
		return err
	}

	return nil
}

// Reads packet from reader
func (p *Packet) Read(r io.Reader) error {
	// Read first 4-byte length-prefix
	if _, err := io.ReadFull(r, p.lenBuf[:]); err != nil {
		return err
	}

	// Allocate buffer for remaining data
	msgLen := int(binary.BigEndian.Uint32(p.lenBuf[:]))
	p.Data = make([]byte, msgLen)

	if _, err := io.ReadFull(r, p.Data); err != nil {
		return err
	}

	return nil
}

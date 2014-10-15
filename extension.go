package rain

import (
	"bytes"
	"encoding/binary"

	"github.com/zeebo/bencode"
)

type extensionHandshakeMessage struct {
	M            map[string]uint8 `bencode:"m"`
	MetadataSize uint32           `bencode:"metadata_size,omitempty"`
}

// Extension IDs
const (
	extensionHandshakeID uint8 = iota
	extensionMetadataID
)

var extensionMapping = map[string]uint8{
	"ut_metadata": extensionMetadataID,
}

func (p *peer) sendExtensionHandshake() error {
	m := extensionHandshakeMessage{
		M: extensionMapping,
	}
	p.log.Debugf("Sending extension handshake %+v", m)
	b, err := bencode.EncodeBytes(&m)
	if err != nil {
		return err
	}
	return p.sendExtensionMessage(extensionHandshakeID, b)
}

func (p *peer) sendExtensionMessage(id byte, payload []byte) error {
	msg := struct {
		Length      uint32
		BTID        byte
		ExtensionID byte
	}{
		Length:      uint32(len(payload)) + 2,
		BTID:        extensionID,
		ExtensionID: id,
	}
	p.log.Debugf("Sending extension message %+v", msg)
	buf := bytes.NewBuffer(make([]byte, 0, 6+len(payload)))
	err := binary.Write(buf, binary.BigEndian, msg)
	if err != nil {
		return err
	}
	_, err = buf.Write(payload)
	if err != nil {
		return err
	}
	return binary.Write(p.conn, binary.BigEndian, buf.Bytes())
}

package rain

import (
	"bytes"

	"github.com/zeebo/bencode"
)

const metadataPieceSize = 16 * 1024

// Metadata Extension Message Types
const (
	metadataRequest = iota
	metadataData
	metadataReject
)

func sendMetadataMessage(m *metadataMessage, p *peer, id uint8) error {
	var buf bytes.Buffer
	e := bencode.NewEncoder(&buf)
	err := e.Encode(m)
	if err != nil {
		return err
	}
	return p.sendExtensionMessage(id, buf.Bytes())
}

type metadataMessage struct {
	MessageType uint8  `bencode:"msg_type"`
	Piece       uint32 `bencode:"piece"`
}

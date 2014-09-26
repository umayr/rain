package piece

import (
	"bytes"
	"crypto/sha1"
	"io"
	"os"

	"github.com/cenkalti/rain/internal/torrent"
)

type Piece struct {
	index  uint32 // piece index in whole torrent
	hash   []byte
	length uint32   // last piece may not be complete
	files  sections // the place to write downloaded bytes
	blocks []Block
}

type Block struct {
	Begin  uint32
	Length uint32
}

func NewPieces(info *torrent.Info, osFiles []*os.File, blockSize uint32) []*Piece {
	var (
		fileIndex  int   // index of the current file in torrent
		fileLength int64 = info.GetFiles()[0].Length
		fileEnd    int64 = fileLength // absolute position of end of the file among all pieces
		fileOffset int64              // offset in file: [0, fileLength)
	)

	nextFile := func() {
		fileIndex++
		fileLength = info.GetFiles()[fileIndex].Length
		fileEnd += fileLength
		fileOffset = 0
	}
	fileLeft := func() int64 { return fileLength - fileOffset }

	// Construct pieces
	var total int64
	pieces := make([]*Piece, info.NumPieces)
	for i := uint32(0); i < info.NumPieces; i++ {
		p := &Piece{
			index: i,
			hash:  info.PieceHash(i),
		}

		// Construct p.files
		var pieceOffset uint32
		pieceLeft := func() uint32 { return info.PieceLength - pieceOffset }
		for left := pieceLeft(); left > 0; {
			n := uint32(minInt64(int64(left), fileLeft())) // number of bytes to write

			file := section{osFiles[fileIndex], fileOffset, int64(n)}
			p.files = append(p.files, file)

			left -= n
			p.length += n
			pieceOffset += n
			fileOffset += int64(n)
			total += int64(n)

			if total == info.TotalLength {
				break
			}
			if fileLeft() == 0 {
				nextFile()
			}
		}

		p.blocks = newBlocks(p.length, p.files, blockSize)
		pieces[i] = p
	}
	return pieces
}

func newBlocks(pieceLength uint32, files sections, blockSize uint32) []Block {
	div, mod := divMod32(pieceLength, blockSize)
	numBlocks := div
	if mod != 0 {
		numBlocks++
	}
	blocks := make([]Block, numBlocks)
	for j := uint32(0); j < div; j++ {
		blocks[j] = Block{
			Begin:  j * blockSize,
			Length: blockSize,
		}
	}
	if mod != 0 {
		blocks[numBlocks-1] = Block{
			Begin:  (numBlocks - 1) * blockSize,
			Length: uint32(mod),
		}
	}
	return blocks
}

func (p *Piece) Index() uint32                     { return p.index }
func (p *Piece) Blocks() []Block                   { return p.blocks }
func (p *Piece) Length() uint32                    { return p.length }
func (p *Piece) Hash() []byte                      { return p.hash }
func (p *Piece) Reader() io.Reader                 { return p.files.Reader() }
func (p *Piece) Write(b []byte) (n int, err error) { return p.files.Write(b) }

func (p *Piece) HashCheck() (ok bool, err error) {
	hash := sha1.New()
	_, err = io.CopyN(hash, p.files.Reader(), int64(p.Length()))
	if err != nil {
		return
	}
	if bytes.Equal(hash.Sum(nil), p.hash) {
		ok = true
	}
	return
}

func divMod32(a, b uint32) (uint32, uint32) { return a / b, a % b }

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
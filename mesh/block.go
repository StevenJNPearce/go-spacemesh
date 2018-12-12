package mesh

import (
	"github.com/google/uuid"
	"time"
)

type BlockID uint64
type LayerID uint64

var layerCounter LayerID = 0

type Block struct {
	id         BlockID
	LayerIndex LayerID
	Data       []byte
	Coin       bool
	Timestamp  time.Time
	ProVotes   uint64
	ConVotes   uint64
	BlockVotes map[BlockID]bool
}

func (b Block) Id() BlockID {
	return b.id
}

func (b Block) Layer() LayerID {
	return b.LayerIndex
}

func NewExistingBlock(id BlockID, layerIndex LayerID, data []byte) *Block {
	b := Block{
		id:         BlockID(id),
		BlockVotes: make(map[BlockID]bool),
		LayerIndex: LayerID(layerIndex),
		Data:       data,
	}
	return &b
}

func NewBlock(coin bool, data []byte, ts time.Time, layerId LayerID) *Block {
	b := Block{
		id:         BlockID(uuid.New().ID()),
		LayerIndex: layerId,
		BlockVotes: make(map[BlockID]bool),
		Timestamp:  ts,
		Data:       data,
		Coin:       coin,
		ProVotes:   0,
		ConVotes:   0,
	}
	return &b
}

type Layer struct {
	blocks []*Block
	index  LayerID
}

func (l *Layer) Index() LayerID {
	return l.index
}

func (l *Layer) Blocks() []*Block {
	return l.blocks
}

func (l *Layer) Hash() []byte {
	return []byte("some hash representing the layer")
}

func (l *Layer) AddBlock(block *Block) {
	block.LayerIndex = l.index
	l.blocks = append(l.blocks, block)
}

func NewLayer() *Layer {
	l := Layer{
		blocks: make([]*Block, 0),
		index:  layerCounter,
	}
	layerCounter++
	return &l
}

func NewExistingLayer(idx LayerID, blocks []*Block) *Layer {
	l := Layer{
		blocks: blocks,
		index:  idx,
	}
	return &l
}

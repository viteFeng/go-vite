package protocols

import (
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/log15"
	"math/big"
	"sync"
)

type Set interface {
	Has(interface{}) bool
	Add(interface{})
	Del(interface{})
	Count() uint
}

// used for mark request type
type reqFlag int

const (
	snapshotHeadersFlag reqFlag = iota
	accountblocksFlag
	snapshotBlocksFlag
	blocksFlag
)

type reqInfo struct {
	id   uint64
	flag reqFlag
	size uint64
}

func (r *reqInfo) Replay() {

}

const filterCap = 100000

// @section Peer for protocol handle, not p2p Peer.
type Peer struct {
	ts      Transport
	ID      string
	Head    types.Hash
	Height  *big.Int
	Version int
	Lock    sync.RWMutex

	// use this channel to ensure that only one goroutine send msg simultaneously.
	sending chan struct{}

	KnownBlocks Set

	Log log15.Logger

	term chan struct{}

	// wait for sending
	sendSnapshotBlock chan *ledger.SnapshotBlock
	sendAccountBlock  chan *ledger.AccountBlock

	// response performance
	Speed int
}

func newPeer() *Peer {
	return &Peer{
		sending:           make(chan struct{}, 1),
		Log:               log15.New("module", "net/peer"),
		term:              make(chan struct{}),
		sendSnapshotBlock: make(chan *ledger.SnapshotBlock),
		sendAccountBlock:  make(chan *ledger.AccountBlock),
		KnownBlocks:       NewCuckooSet(filterCap),
	}
}

func (p *Peer) HandShake() error {
	return nil
}

func (p *Peer) Update(head types.Hash, height *big.Int) {
	p.Lock.Lock()
	defer p.Lock.Unlock()

	p.Head = head
	p.Height = height
	p.Log.Info("update status", "ID", p.ID, "height", p.Height, "head", p.Head)
}

func (p *Peer) Altitude() (head types.Hash, height *big.Int) {
	p.Lock.RLock()
	defer p.Lock.RUnlock()

	return p.Head, p.Height
}

func (p *Peer) Broadcast() {

}

// response
func (p *Peer) SendSnapshotBlockHeaders() {

}

func (p *Peer) SendSnapshotBlockBodies() {

}

func (p *Peer) SendSnapshotBlocks() {

}

func (p *Peer) SendAccountBlocks() {

}

func (p *Peer) SendSubLedger() {

}

// request
func (p *Peer) RequestSnapshotHeaders() {

}

func (p *Peer) RequestSnapshotBodies() {

}

func (p *Peer) RequestSnapshotBlocks() {

}

func (p *Peer) RequestAccountBlocks() {

}

func (p *Peer) RequestSubLedger() {

}

func (p *Peer) Send(set MsgSet, code MsgCode, msg Serializable) {

}

func (p *Peer) Destroy() {
	select {
	case <-p.term:
	default:
		close(p.term)
	}
}

type PeerInfo struct {
	Addr   string
	Flag   int
	Head   string
	Height uint64
}

func (p *Peer) Info() *PeerInfo {
	return &PeerInfo{}
}

// @section PeerSet
type peerSet struct {
	peers map[string]*Peer
	rw    sync.RWMutex
}

func NewPeerSet() *peerSet {
	return &peerSet{
		peers: make(map[string]*Peer),
	}
}

func (m *peerSet) BestPeer() (best *Peer) {
	m.rw.RLock()
	defer m.rw.RUnlock()

	maxHeight := new(big.Int)
	for _, peer := range m.peers {
		cmp := peer.Height.Cmp(maxHeight)
		if cmp > 0 {
			maxHeight = peer.Height
			best = peer
		}
	}

	return
}

func (m *peerSet) Has(id string) bool {
	_, ok := m.peers[id]
	return ok
}

func (m *peerSet) Add(peer *Peer) {
	m.rw.Lock()
	m.peers[peer.ID] = peer
	m.rw.Unlock()
}

func (m *peerSet) Del(peer *Peer) {
	m.rw.Lock()
	delete(m.peers, peer.ID)
	m.rw.Unlock()
}

func (m *peerSet) Count() int {
	m.rw.RLock()
	defer m.rw.RUnlock()

	return len(m.peers)
}

func (m *peerSet) Pick(height *big.Int) (peers []*Peer) {
	m.rw.RLock()
	defer m.rw.RUnlock()

	for _, p := range m.peers {
		if p.Height.Cmp(height) > 0 {
			peers = append(peers, p)
		}
	}
	return
}

func (m *peerSet) Info() (info []*PeerInfo) {
	m.rw.RLock()
	defer m.rw.RUnlock()

	for _, peer := range m.peers {
		info = append(info, peer.Info())
	}

	return
}

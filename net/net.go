package net

import (
	"github.com/seiflotfy/cuckoofilter"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/ledger"
	"log"
	"math/big"
	"net"
	"sync"
	"time"
)

type Config struct {
	NetID uint64
}

type Net struct {
	*Config
	start         time.Time
	peers         *peerSet
	snapshotFeed  *snapshotBlockFeed
	accountFeed   *accountBlockFeed
	term          chan struct{}
	pool          *reqPool
	FromHeight    *big.Int
	TargetHeight  *big.Int
	syncState     SyncState
	slock         sync.RWMutex // use for syncState change
	SyncStartHook func(*big.Int, *big.Int)
	SyncDoneHook  func(*big.Int, *big.Int)
	SyncErrHook   func(*big.Int, *big.Int)
	stateFeed     *SyncStateFeed
	SnapshotChain BlockChain
	blockRecord   *cuckoofilter.CuckooFilter // record blocks has retrieved from network
}

func New(cfg *Config) *Net {
	n := &Net{
		Config:       cfg,
		peers:        NewPeerSet(),
		snapshotFeed: new(snapshotBlockFeed),
		accountFeed:  new(accountBlockFeed),
		stateFeed:    new(SyncStateFeed),
		term:         make(chan struct{}),
		pool:         new(reqPool),
		blockRecord:  cuckoofilter.NewCuckooFilter(10000),
	}

	return n
}

func (n *Net) Start() {
	n.start = time.Now()
}

func (n *Net) Stop() {
	select {
	case <-n.term:
	default:
		close(n.term)
	}
}

func (n *Net) Syncing() bool {
	n.slock.RLock()
	defer n.slock.RUnlock()
	return n.syncState == Syncing
}

func (n *Net) SetSyncState(st SyncState) {
	n.slock.Lock()
	defer n.slock.Unlock()
	n.syncState = st
	n.stateFeed.Notify(st)
}

func (n *Net) ReceiveConn(conn net.Conn) {
	select {
	case <-n.term:
	default:
	}

}

func (n *Net) HandlePeer(p *Peer) {
	head, err := n.SnapshotChain.GetLatestSnapshotBlock()
	if err != nil {
		log.Fatal("cannot get current block", err)
	}

	genesis, err := n.SnapshotChain.GetGenesesBlock()
	if err != nil {
		log.Fatal("cannot get genesis block", err)
	}

	err := p.Handshake(n.NetID, head.Height, head.Hash, genesis.Hash)
	if err != nil {
		return
	}
}

func (n *Net) BroadcastSnapshotBlocks(blocks []*ledger.SnapshotBlock, propagate bool) {

}

func (n *Net) BroadcastAccountBlocks(blocks []*ledger.AccountBlock, propagate bool) {

}

type snap struct {
	From    types.Hash
	To      types.Hash
	Count   uint64
	Forward bool
	Step    int
}

func (n *Net) FetchSnapshotBlocks(s *snap) {

}

type ac map[types.Address]*snap

func (n *Net) FetchAccountBlocks(a ac) {

}

func (n *Net) SubscribeAccountBlock(fn func(block *ledger.AccountBlock)) (subId int) {
	return n.accountFeed.Sub(fn)
}

func (n *Net) UnsubscribeAccountBlock(subId int) {
	n.accountFeed.Unsub(subId)
}

func (n *Net) receiveAccountBlock(block *ledger.AccountBlock) {
	n.accountFeed.Notify(block)
}

func (n *Net) SubscribeSnapshotBlock(fn func(block *ledger.SnapshotBlock)) (subId int) {
	return n.snapshotFeed.Sub(fn)
}

func (n *Net) UnsubscribeSnapshotBlock(subId int) {
	n.snapshotFeed.Unsub(subId)
}

func (n *Net) receiveSnapshotBlock(block *ledger.SnapshotBlock) {
	n.snapshotFeed.Notify(block)
}

func (n *Net) SubscribeSyncStatus(fn func(SyncState)) (subId int) {
	return n.stateFeed.Sub(fn)
}

func (n *Net) UnsubscribeSyncStatus(subId int) {
	n.stateFeed.Unsub(subId)
}

// get current netInfo (peers, syncStatus, ...)
func (n *Net) Status() *NetStatus {
	running := true
	select {
	case <-n.term:
		running = false
	default:
	}

	return &NetStatus{
		Peers:      n.peers.Info(),
		Running:    running,
		Uptime:     time.Now().Sub(n.start),
		SyncStatus: n.syncState.String(),
	}
}

type NetStatus struct {
	Peers      []*PeerInfo
	SyncStatus string
	Uptime     time.Duration
	Running    bool
}
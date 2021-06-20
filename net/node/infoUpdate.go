package node

import (
	. "GoOnchain/net/message"
	. "GoOnchain/net/protocol"
	"time"
)

func keepAlive(from *Noder, dst *Noder) {
	// Need move to node function or keep here?
}

func (node node) GetBlkHdrs() {
	for _, n := range node.local.List {
		h1 := n.GetHeight()
		h2:= node.local.GetLedger().GetLocalBlockChainHeight()
		if (node.GetState() == ESTABLISH) && (h1 > uint64(h2)) {
			buf, _ := NewMsg("getheaders", node.local)
			go node.Tx(buf)
		}
	}
}

func (node *node) SyncBlk() {
	headerHeight := ledger.DefaultLedger.Store.GetHeaderHeight()
	currentBlkHeight := ledger.DefaultLedger.Blockchain.BlockHeight
	if currentBlkHeight >= headerHeight {
		return
	}
	var dValue int32
	var reqCnt uint32
	var i uint32
	noders := node.local.GetNeighborNoder()
	for _, n := range noders {
		count := MAXREQBLKONCE - uint32(n.GetFlightHeightCnt())
		dValue = int32(headerHeight - currentBlkHeight - reqCnt)
		for i = 1; i <= count && dValue >= 0; i++ {
			hash := ledger.DefaultLedger.Store.GetHeaderHashByHeight(currentBlkHeight + reqCnt)
			ReqBlkData(n, hash)
			n.StoreFlightHeight(currentBlkHeight + reqCnt)
			reqCnt++
			dValue--
		}
	}
}

func (node *node) SendPingToNbr() {
	noders := node.local.GetNeighborNoder()
	for _, n := range noders {
		if n.GetState() == ESTABLISH {
			buf, err := NewPingMsg()
			if err != nil {
				log.Error("failed build a new ping message")
			} else {
				go n.Tx(buf)
			}
		}
	}
}

func (node *node) HeartBeatMonitor() {
	log.Info("heart beat")
	noders := node.local.GetNeighborNoder()
	for _, n := range noders {
		if n.GetState() == ESTABLISH {
			t := n.GetLastCcntmact()
			if time.Since(t).Seconds() > KEEPALIVETIMEOUT {
				log.Warn("keepalive timeout!!!")
				n.SetState(INACTIVITY)
				//n.CloseConn()
			}
		}
	}
}

func (node node) ReqNeighborList() {
	buf, _ := NewMsg("getaddr", node.local)
	go node.Tx(buf)
}

// Fixme the Nodes should be a parameter
func (node node) updateNodeInfo() {
	ticker := time.NewTicker(time.Second * PERIODUPDATETIME)
	quit := make(chan struct{})
	for {
		timer := time.NewTimer(time.Second * HEARTBEAT)
		select {
		case <-ticker.C:
			//GetHeaders process haven't finished yet. So comment it now.
			node.SendPingToNbr()
			node.GetBlkHdrs()
			node.SyncBlk()
		case <-quit:
			ticker.Stop()
			return
		case <-timer.C:
			node.HeartBeatMonitor()
		}
	}
	// TODO when to close the timer
	//close(quit)
}

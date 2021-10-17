package actor

import (
	"github.com/cntmID/eventbus/actor"
	"github.com/Ontology/net/protocol"
	"github.com/Ontology/common/log"
)

var NetServerPid *actor.PID
var node protocol.Noder
type MsgActor struct{}

type GetConnectionCntReq struct {
}
type GetConnectionCntRsp struct {
	Cnt uint
}

func (state *MsgActor) Receive(ccntmext actor.Ccntmext) {
	switch ccntmext.Message().(type) {
	case *GetConnectionCntReq:
		connectionCnt := node.GetConnectionCnt()
		ccntmext.Sender().Request(&GetConnectionCntRsp{Cnt: connectionCnt}, ccntmext.Self())
	default:
		err := node.Xmit(ccntmext.Message())
		if nil != err {
			log.Error("Error Xmit message ", err.Error())
		}
	}
}

func init() {
	props := actor.FromProducer(func() actor.Actor { return &MsgActor{} })
	NetServerPid = actor.Spawn(props)
}

func SetNode(netNode protocol.Noder){
	node = netNode
}

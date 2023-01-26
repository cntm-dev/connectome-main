/*
 * Copyright (C) 2018 The cntmology Authors
 * This file is part of The cntmology library.
 *
 * The cntmology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The cntmology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * alcntm with The cntmology.  If not, see <http://www.gnu.org/licenses/>.
 */

package nodeinfo

import (
	"time"

	"github.com/cntmio/cntmology/common/config"
	"github.com/cntmio/cntmology/core/ledger"
	"github.com/cntmio/cntmology/p2pserver/net/netserver"
	p2p "github.com/cntmio/cntmology/p2pserver/net/protocol"
	"github.com/cntmio/cntmology/p2pserver/protocols"
	prom "github.com/prometheus/client_golang/prometheus"
)

var (
	nodePortMetric = prom.NewGauge(prom.GaugeOpts{
		Name: "cntmology_nodeport",
		Help: "cntmology node port",
	})

	blockHeightMetric = prom.NewGauge(prom.GaugeOpts{
		Name: "cntmology_block_height",
		Help: "cntmology blockchain block height",
	})

	inboundsCountMetric = prom.NewGauge(prom.GaugeOpts{
		Name: "cntmology_p2p_inbounds_count",
		Help: "cntmology p2p inbloud count",
	})

	outboundsCountMetric = prom.NewGauge(prom.GaugeOpts{
		Name: "cntmology_p2p_outbounds_count",
		Help: "cntmology p2p outbloud count",
	})

	peerStatusMetric = prom.NewGaugeVec(prom.GaugeOpts{
		Name: "cntmology_p2p_peer_status",
		Help: "cntmology peer info",
	}, []string{"ip", "id"})

	reconnectCountMetric = prom.NewGauge(prom.GaugeOpts{
		Name: "cntmology_p2p_reconnect_count",
		Help: "cntmology p2p reconnect count",
	})
)

var (
	metrics = []prom.Collector{nodePortMetric, blockHeightMetric, inboundsCountMetric,
		outboundsCountMetric, peerStatusMetric, reconnectCountMetric}
)

func initMetric() error {
	for _, curMetric := range metrics {
		if err := prom.Register(curMetric); err != nil {
			return err
		}
	}

	return nil
}

func metricUpdate(n p2p.P2P) {
	nodePortMetric.Set(float64(config.DefConfig.P2PNode.NodePort))

	blockHeightMetric.Set(float64(ledger.DefLedger.GetCurrentBlockHeight()))

	ns, ok := n.(*netserver.NetServer)
	if !ok {
		return
	}

	inboundsCountMetric.Set(float64(ns.ConnectCcntmroller().InboundsCount()))
	outboundsCountMetric.Set(float64(ns.ConnectCcntmroller().OutboundsCount()))

	peers := ns.GetNeighbors()
	for _, curPeer := range peers {
		id := curPeer.GetID()

		// label: IP PeedID
		peerStatusMetric.WithLabelValues(curPeer.GetAddr(), id.ToHexString()).Set(float64(curPeer.GetHeight()))
	}

	pt := ns.Protocol()
	mh, ok := pt.(*protocols.MsgHandler)
	if !ok {
		return
	}

	reconnectCountMetric.Set(float64(mh.ReconnectService().ReconnectCount()))
}

func updateMetric(n p2p.P2P) {
	tk := time.NewTicker(time.Minute)
	defer tk.Stop()
	for {
		select {
		case <-tk.C:
			metricUpdate(n)
		}
	}
}

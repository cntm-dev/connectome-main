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

package node

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	. "github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	"github.com/Ontology/events"
	msg "github.com/Ontology/net/message"
	. "github.com/Ontology/net/protocol"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type link struct {
	addr  string    // The address of the node
	conn  net.Conn  // Connect socket with the peer node
	port  uint16    // The server port of the node
	time  time.Time // The latest time the node activity
	rxBuf struct {  // The RX buffer of this node to solve mutliple packets problem
		p   []byte
		len int
	}
	connCnt uint64 // The connection count
}

// Shrinking the buf to the exactly reading in byte length
//@Return @1 the start header of next message, the left length of the next message
func unpackNodeBuf(node *node, buf []byte) {
	var msgLen int
	var msgBuf []byte
	if node.rxBuf.p == nil {
		if len(buf) < MSGHDRLEN {
			log.Warn("Unexpected size of received message")
			errors.New("Unexpected size of received message")
			return
		}
		// FIXME Check the payload < 0 error case
		//fmt.Printf("The Rx msg payload is %d\n", PayloadLen(buf))
		msgLen = PayloadLen(buf) + MSGHDRLEN
	} else {
		msgLen = node.rxBuf.len
	}

	//fmt.Printf("The msg length is %d, buf len is %d\n", msgLen, len(buf))
	if len(buf) == msgLen {
		msgBuf = append(node.rxBuf.p, buf[:]...)
		go HandleNodeMsg(node, msgBuf, len(msgBuf))
		node.rxBuf.p = nil
		node.rxBuf.len = 0
	} else if len(buf) < msgLen {
		node.rxBuf.p = append(node.rxBuf.p, buf[:]...)
		node.rxBuf.len = msgLen - len(buf)
	} else {
		msgBuf = append(node.rxBuf.p, buf[0:msgLen]...)
		go msg.HandleNodeMsg(node, msgBuf, len(msgBuf))
		node.rxBuf.p = nil
		node.rxBuf.len = 0

		unpackNodeBuf(node, buf[msgLen:])
	}
}

func (node *node) rx() {
	conn := node.getConn()
	buf := make([]byte, MAXBUFLEN)
	for {
		len, err := conn.Read(buf[0:(MAXBUFLEN - 1)])
		buf[MAXBUFLEN-1] = 0 //Prevent overflow
		switch err {
		case nil:
			unpackNodeBuf(node, buf[0:len])
			//go handleNodeMsg(node, buf, len)
			break
		case io.EOF:
			//fmt.Println("Reading EOF of network conn")
			break
		default:
			log.Error("Read connetion error ", err)
			goto disconnect
		}
	}

disconnect:
	err := conn.Close()
	return err
}

func printIPAddr() {
	host, _ := os.Hostname()
	addrs, _ := net.LookupIP(host)
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			log.Info("IPv4: ", ipv4)
		}
	}
}

func (link *link) CloseConn() {
	link.conn.Close()
}

func (n *node) initConnection() {
	isTls := Parameters.IsTLS
	var listener net.Listener
	var err error
	if isTls {
		listener, err = initTlsListen()
		if err != nil {
			log.Error("TLS listen failed")
			return
		}
	} else {
		listener, err = initNonTlsListen()
		if err != nil {
			log.Error("non TLS listen failed")
			return
		}
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error("Error accepting ", err.Error())
			return
		}
		log.Info("Remote node connect with ", conn.RemoteAddr(), conn.LocalAddr())

		n.link.connCnt++

		node := NewNode()
		node.addr, err = parseIPaddr(conn.RemoteAddr().String())
		node.local = n
		node.conn = conn
		go node.rx()
	}
	//TODO Release the net listen resouce
}

func initNonTlsListen() (net.Listener, error) {
	log.Debug()
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(Parameters.NodePort))
	if err != nil {
		log.Error("Error listening\n", err.Error())
		return nil, err
	}
	return listener, nil
}

func initTlsListen() (net.Listener, error) {
	CertPath := Parameters.CertPath
	KeyPath := Parameters.KeyPath
	CAPath := Parameters.CAPath

	// load cert
	cert, err := tls.LoadX509KeyPair(CertPath, KeyPath)
	if err != nil {
		log.Error("load keys fail", err)
		return nil, err
	}
	// load root ca
	caData, err := ioutil.ReadFile(CAPath)
	if err != nil {
		log.Error("read ca fail", err)
		return nil, err
	}
	pool := x509.NewCertPool()
	ret := pool.AppendCertsFromPEM(caData)
	if !ret {
		return nil, errors.New("failed to parse root certificate")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    pool,
	}

	log.Info("TLS listen port is ", strconv.Itoa(Parameters.NodePort))
	listener, err := tls.Listen("tcp", ":"+strconv.Itoa(Parameters.NodePort), tlsConfig)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return listener, nil
}

func parseIPaddr(s string) (string, error) {
	i := strings.Index(s, ":")
	if i < 0 {
		log.Warn("Split IP address&port error")
		return s, errors.New("Split IP address&port error")
	}
	return s[:i], nil
}

func (node *node) Connect(nodeAddr string) {
	node.chF <- func() error {
		common.Trace()
		isTls := Parameters.IsTLS
		var conn net.Conn
		var err error
		if isTls {
			conn, err = TLSDial(nodeAddr)
			if err != nil {
				log.Error("TLS connect failed: ", err)
				return nil
			}
		} else {
			conn, err = NonTLSDial(nodeAddr)
			if err != nil {
				log.Error("non TLS connect failed:", err)
				return nil
			}
		}
		node.link.connCnt++

		n := NewNode()
		n.conn = conn
		n.addr, err = parseIPaddr(conn.RemoteAddr().String())
		n.local = node

		log.Info(fmt.Sprintf("Connect node %s connect with %s with %s",
			conn.LocalAddr().String(), conn.RemoteAddr().String(),
			conn.RemoteAddr().Network()))
		go n.rx()

		time.Sleep(2 * time.Second)
		// FIXME is there any timing race with rx
		buf, _ := NewVersion(node)
		n.Tx(buf)
		return nil
	}
}

func NonTLSDial(nodeAddr string) (net.Conn, error) {
	log.Debug()
	conn, err := net.DialTimeout("tcp", nodeAddr, time.Second * DIALTIMEOUT)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func TLSDial(nodeAddr string) (net.Conn, error) {
	CertPath := Parameters.CertPath
	KeyPath := Parameters.KeyPath
	CAPath := Parameters.CAPath

	clientCertPool := x509.NewCertPool()

	cacert, err := ioutil.ReadFile(CAPath)
	cert, err := tls.LoadX509KeyPair(CertPath, KeyPath)
	if err != nil {
		return nil, err
	}

	ret := clientCertPool.AppendCertsFromPEM(cacert)
	if !ret {
		return nil, errors.New("failed to parse root certificate")
	}

	conf := &tls.Config{
		RootCAs:      clientCertPool,
		Certificates: []tls.Certificate{cert},
	}

	var dialer net.Dialer
	dialer.Timeout = time.Second * DIALTIMEOUT
	conn, err := tls.DialWithDialer(&dialer, "tcp", nodeAddr, conf)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// TODO construct a TX channel and other application just drop the message to the channel
func (node node) Tx(buf []byte) {
	//node.chF <- func() error {
	common.Trace()
	str := hex.EncodeToString(buf)
	log.Debug(fmt.Sprintf("TX buf length: %d\n%s", len(buf), str))

	_, err := node.conn.Write(buf)
	if err != nil {
		log.Error("Error sending messge to peer node ", err.Error())
	}
	//return err
	//}
}

// func (net net) Xmit(inv Inventory) error {
// 	//if (!KnownHashes.Add(inventory.Hash)) return false;
// 	t := inv.Type()
// 	switch t {
// 	case BLOCK:
//                 if (Blockchain.Default == null) {
// 			return false
// 		}
//                 Block block = (Block)inventory;
//                 if (Blockchain.Default.CcntmainsBlock(block.Hash)) {
// 			return false;
// 		}
//                 if (!Blockchain.Default.AddBlock(block)) {
// 			return false;
// 		}
// 	case TRANSACTION:
// 		if (!AddTransaction((Transaction)inventory)) {
// 			return false
// 		}
// 	case CONSENSUS:
//                 if (!inventory.Verify()) {
// 			return false
// 		}
// 	default:
// 		fmt.Print("Unknow inventory type/n")
// 		return errors.New("Unknow inventory type/n")
// 	}

// 	RelayCache.Add(inventory);
// 	foreach (RemoteNode node in connectedPeers)
// 	relayed |= node.Relay(inventory);
// 	NewInventory.Invoke(this, inventory);
// }

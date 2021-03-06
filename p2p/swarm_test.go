package p2p

import (
	"github.com/spacemeshos/go-spacemesh/p2p/connectionpool"
	"github.com/spacemeshos/go-spacemesh/p2p/dht"
	"testing"
	"time"

	"context"
	"errors"
	"github.com/gogo/protobuf/proto"
	"github.com/spacemeshos/go-spacemesh/crypto"
	"github.com/spacemeshos/go-spacemesh/p2p/config"
	"github.com/spacemeshos/go-spacemesh/p2p/message"
	"github.com/spacemeshos/go-spacemesh/p2p/net"
	"github.com/spacemeshos/go-spacemesh/p2p/node"
	"github.com/spacemeshos/go-spacemesh/p2p/pb"
	"github.com/spacemeshos/go-spacemesh/p2p/service"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"sync"
)

type cpoolMock struct {
	f func(address string, pk crypto.PublicKey) (net.Connection, error)
}

func (cp *cpoolMock) GetConnection(address string, pk crypto.PublicKey) (net.Connection, error) {
	if cp.f != nil {
		return cp.f(address, pk)
	}
	return net.NewConnectionMock(pk), nil
}

func (cp *cpoolMock) RemoteConnectionsChannel() chan net.NewConnectionEvent {
	return make(chan net.NewConnectionEvent)
}

func p2pTestInstance(t testing.TB, config config.Config) *swarm {
	port, err := node.GetUnboundedPort()
	assert.NoError(t, err, "Error getting a port", err)
	config.TCPPort = port
	p, err := newSwarm(context.TODO(), config, true, true)
	assert.NoError(t, err, "Error creating p2p stack, err: %v", err)
	assert.NotNil(t, p)
	p.Start()
	return p
}

func p2pTestNoStart(t testing.TB, config config.Config) *swarm {
	port, err := node.GetUnboundedPort()
	assert.NoError(t, err, "Error getting a port", err)
	config.TCPPort = port
	p, err := newSwarm(context.TODO(), config, true, true)
	assert.NoError(t, err, "Error creating p2p stack, err: %v", err)
	assert.NotNil(t, p)
	return p
}

const exampleProtocol = "EX"
const examplePayload = "Example"

func TestNew(t *testing.T) {
	s, err := New(context.TODO(), config.DefaultConfig())
	assert.NoError(t, err, err)
	err = s.Start()
	assert.NoError(t, err, err)
	assert.NotNil(t, s, "its nil")
	s.Shutdown()
}

func Test_newSwarm(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.TCPPort = int(crypto.GetRandomUserPort())
	s, err := newSwarm(context.TODO(), cfg, true, false)
	assert.NoError(t, err)
	err = s.Start()
	assert.NoError(t, err, err)
	assert.NotNil(t, s)
	s.Shutdown()
}

func TestSwarm_Shutdown(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.TCPPort = int(crypto.GetRandomUserPort())
	s, err := newSwarm(context.TODO(), cfg, true, false)
	assert.NoError(t, err)
	err = s.Start()
	assert.NoError(t, err, err)
	s.Shutdown()

	select {
	case _, ok := <-s.shutdown:
		assert.False(t, ok)
	case <-time.After(1 * time.Second):
		t.Error("Failed to shutdown")
	}
}

func TestSwarm_ShutdownNoStart(t *testing.T) {
	s, err := newSwarm(context.TODO(), config.DefaultConfig(), true, false)
	assert.NoError(t, err)
	s.Shutdown()
}

func TestSwarm_RegisterProtocolNoStart(t *testing.T) {
	s, err := newSwarm(context.TODO(), config.DefaultConfig(), true, false)
	msgs := s.RegisterProtocol("Anton")
	assert.NotNil(t, msgs)
	assert.NoError(t, err)
	s.Shutdown()
}

func TestSwarm_processMessage(t *testing.T) {
	s := swarm{}
	s.lNode, _ = node.GenerateTestNode(t)
	r := node.GenerateRandomNodeData()
	c := &net.ConnectionMock{}
	c.SetRemotePublicKey(r.PublicKey())
	ime := net.IncomingMessageEvent{Message: []byte("0"), Conn: c}
	s.processMessage(ime) // should error

	assert.True(t, c.Closed())
}

func TestSwarm_authAuthor(t *testing.T) {
	// create a message

	priv, pub, err := crypto.GenerateKeyPair()

	assert.NoError(t, err, err)
	assert.NotNil(t, priv)
	assert.NotNil(t, pub)

	pm := &pb.ProtocolMessage{
		Metadata: message.NewProtocolMessageMetadata(pub, exampleProtocol),
		Data:     &pb.ProtocolMessage_Payload{[]byte(examplePayload)},
	}
	ppm, err := proto.Marshal(pm)
	assert.NoError(t, err, "cant marshal msg ", err)

	// sign it
	s, err := priv.Sign(ppm)
	assert.NoError(t, err, "cant sign ", err)

	pm.Metadata.MsgSign = s

	vererr := message.AuthAuthor(pm)
	assert.NoError(t, vererr)

	priv2, pub2, err := crypto.GenerateKeyPair()

	assert.NoError(t, err, err)
	assert.NotNil(t, priv2)
	assert.NotNil(t, pub2)

	s, err = priv2.Sign(ppm)
	assert.NoError(t, err, "cant sign ", err)

	pm.Metadata.MsgSign = s

	vererr = message.AuthAuthor(pm)
	assert.Error(t, vererr)
}

func TestSwarm_SignAuth(t *testing.T) {
	n, _ := node.GenerateTestNode(t)
	pm := &pb.ProtocolMessage{
		Metadata: message.NewProtocolMessageMetadata(n.PublicKey(), exampleProtocol),
		Data:     &pb.ProtocolMessage_Payload{[]byte(examplePayload)},
	}

	err := message.SignMessage(n.PrivateKey(), pm)
	assert.NoError(t, err)

	err = message.AuthAuthor(pm)

	assert.NoError(t, err)
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func Test_ConnectionBeforeMessage(t *testing.T) {

	numNodes := 5
	var wg sync.WaitGroup

	p2 := p2pTestInstance(t, config.DefaultConfig())
	c2 := p2.RegisterProtocol(exampleProtocol)

	go func () {
		for {
			msg := <-c2 // immediate response will probably trigger GetConnection fast
			p2.SendMessage(msg.Sender().PublicKey().String(), exampleProtocol, []byte("RESP"))
			wg.Done()
		}
	}()

	oldCpool := p2.cPool.(*connectionpool.ConnectionPool)

	//called := make(chan struct{}, numNodes)
	cpm := new(cpoolMock)
	cpm.f = func(address string, pk crypto.PublicKey) (net.Connection, error) {
		c, err := oldCpool.GetConnectionIfExists(pk)
		if err != nil {
			t.Fatal("Didn't get connection yet while SendMessage called GetConnection")
		}
		return c,nil
	}


	p2.cPool = cpm

	payload := []byte(RandString(10))


	for i := 0; i < numNodes; i++ {
		wg.Add(1)
		go func() {
			p1 := p2pTestInstance(t, config.DefaultConfig())
			_ = p1.RegisterProtocol(exampleProtocol)
			p1.dht.Update(p2.lNode.Node)
			err := p1.SendMessage(p2.lNode.PublicKey().String(), exampleProtocol, payload)
			assert.NoError(t, err)
		}()
	}
	wg.Wait()

}

func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func sendDirectMessage(t *testing.T, sender *swarm, recvPub string, inChan chan service.Message, checkpayload bool) {
	payload := []byte(RandString(10))
	err := sender.SendMessage(recvPub, exampleProtocol, payload)
	assert.NoError(t, err)
	select {
	case msg := <-inChan:
		if checkpayload {
			assert.Equal(t, msg.Bytes(), payload)
		}
		assert.Equal(t, msg.Sender().String(), sender.lNode.String())
		break
	case <-time.After(2 * time.Second):
		t.Error("Took too much time to recieve")
	}
}

func TestSwarm_RoundTrip(t *testing.T) {
	p1 := p2pTestInstance(t, config.DefaultConfig())
	p2 := p2pTestInstance(t, config.DefaultConfig())

	exchan1 := p1.RegisterProtocol(exampleProtocol)
	assert.Equal(t, exchan1, p1.protocolHandlers[exampleProtocol])
	exchan2 := p2.RegisterProtocol(exampleProtocol)
	assert.Equal(t, exchan2, p2.protocolHandlers[exampleProtocol])

	p2.dht.Update(p1.lNode.Node)

	sendDirectMessage(t, p2, p1.lNode.PublicKey().String(), exchan1, true)
	sendDirectMessage(t, p1, p2.lNode.PublicKey().String(), exchan2, true)
}

func TestSwarm_MultipleMessages(t *testing.T) {
	p1 := p2pTestInstance(t, config.DefaultConfig())
	p2 := p2pTestInstance(t, config.DefaultConfig())

	exchan1 := p1.RegisterProtocol(exampleProtocol)
	assert.Equal(t, exchan1, p1.protocolHandlers[exampleProtocol])
	exchan2 := p2.RegisterProtocol(exampleProtocol)
	assert.Equal(t, exchan2, p2.protocolHandlers[exampleProtocol])

	err := p2.SendMessage(p1.lNode.String(), exampleProtocol, []byte(examplePayload))
	assert.Error(t, err, "ERR") // should'nt be in routing table
	p2.dht.Update(p1.lNode.Node)

	var wg sync.WaitGroup
	for i := 0; i < 500; i++ {
		wg.Add(1)
		go func() { sendDirectMessage(t, p2, p1.lNode.String(), exchan1, false); wg.Done() }()
	}
	wg.Wait()
}

func TestSwarm_MultipleMessagesFromMultipleSenders(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.SwarmConfig.Gossip = false
	cfg.SwarmConfig.Bootstrap = false

	p1 := p2pTestInstance(t, cfg)

	exchan1 := p1.RegisterProtocol(exampleProtocol)
	assert.Equal(t, exchan1, p1.protocolHandlers[exampleProtocol])

	pend := make(map[string]chan struct{})
	var mu sync.Mutex
	var wg sync.WaitGroup

	go func() {
		for {
			msg := <-exchan1
			sender := msg.Sender().PublicKey().String()
			mu.Lock()
			c, ok := pend[sender]
			if !ok {
				t.FailNow()
			}
			close(c)
			delete(pend, sender)
			mu.Unlock()
			wg.Done()
		}
	}()

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			p := p2pTestInstance(t, cfg)
			p.dht.Update(p1.LocalNode().Node)
			mychan := make(chan struct{})
			mu.Lock()
			pend[p.lNode.Node.PublicKey().String()] = mychan
			mu.Unlock()

			payload := []byte(RandString(10))
			err := p.SendMessage(p1.lNode.PublicKey().String(), exampleProtocol, payload)
			assert.NoError(t, err)
		}()
	}
	wg.Wait()
}

func TestSwarm_RegisterProtocol(t *testing.T) {
	const numPeers = 100
	nodechan := make(chan *swarm)
	cfg := config.DefaultConfig()
	for i := 0; i < numPeers; i++ {
		go func() { // protocols are registered before starting the node and read after that.
			// there ins't an actual need to sync them.
			nod := p2pTestInstance(t, cfg)
			nod.RegisterProtocol(exampleProtocol) // this is example
			nodechan <- nod
		}()
	}
	i := 0
	for r := range nodechan {
		_, ok := r.protocolHandlers[exampleProtocol]
		assert.True(t, ok)
		_, ok = r.protocolHandlers["/dht/1.0/find-node/"]
		assert.True(t, ok)
		i++
		if i == numPeers {
			close(nodechan)
		}
	}
}

func TestSwarm_onRemoteClientMessage(t *testing.T) {
	cfg := config.DefaultConfig()
	id, err := node.NewNodeIdentity(cfg, "0.0.0.0:0000", false)
	assert.NoError(t, err, "we cant make node ?")

	p := p2pTestNoStart(t, cfg)
	nmock := new(net.ConnectionMock)
	nmock.SetRemotePublicKey(id.PublicKey())

	// Test bad format
	imc := net.IncomingMessageEvent{nmock, nil}
	err = p.onRemoteClientMessage(imc)
	assert.Equal(t, err, ErrBadFormat1)

	// Test No Session
	imc.Message = []byte("test")

	err = p.onRemoteClientMessage(imc)
	assert.Equal(t, err, ErrNoSession)

	//Test bad session
	session := &net.SessionMock{}
	session.SetDecrypt(nil, errors.New("fail"))
	imc.Conn.SetSession(session)

	err = p.onRemoteClientMessage(imc)
	assert.Equal(t, err, ErrFailDecrypt)

	//// Test bad format again
	session.SetDecrypt([]byte("wont_format_fo_protocol_message"), nil)

	err = p.onRemoteClientMessage(imc)
	assert.Equal(t, err, ErrBadFormat2)

	// Test bad auth sign
	goodmsg := &pb.ProtocolMessage{
		Metadata: message.NewProtocolMessageMetadata(id.PublicKey(), exampleProtocol), // not signed
		Data:     &pb.ProtocolMessage_Payload{[]byte(examplePayload)},
	}

	goodbin, _ := proto.Marshal(goodmsg)

	imc.Message = goodbin
	session.SetDecrypt(goodbin, nil)

	err = p.onRemoteClientMessage(imc)
	assert.Equal(t, ErrAuthAuthor, err)

	goodmsg.Metadata.Timestamp = time.Now().Add(-time.Hour).Unix()
	err = message.SignMessage(id.PrivateKey(), goodmsg)
	assert.NoError(t, err)
	nosynced, _ := proto.Marshal(goodmsg)
	session.SetDecrypt(nosynced, nil)
	// Test out of sync
	imc.Message = nosynced

	err = p.onRemoteClientMessage(imc)
	assert.Equal(t, ErrOutOfSync, err)

	// Test no protocol
	goodmsg.Metadata.Timestamp = time.Now().Unix()
	goodmsg.Metadata.MsgSign = nil
	err = message.SignMessage(id.PrivateKey(), goodmsg)
	assert.NoError(t, err, err)

	goodbin, _ = proto.Marshal(goodmsg)
	imc.Message = goodbin
	session.SetDecrypt(goodbin, nil)

	err = p.onRemoteClientMessage(imc)
	assert.Equal(t, err, ErrNoProtocol)

	// Test no err

	var wg sync.WaitGroup
	c := p.RegisterProtocol(exampleProtocol)
	go func() {
		ti := time.After(1 * time.Second)
		select {
		case <-c:
			wg.Done()
			break
		case <-ti:
			t.Error("Didn't get message in time")

		}
	}()
	wg.Add(1)
	err = p.onRemoteClientMessage(imc)
	assert.NoError(t, err)
	wg.Wait()

}

func assertNewPeerEvent(t *testing.T, peer crypto.PublicKey, connChan <-chan crypto.PublicKey) {
	select {
	case newPeer := <-connChan:
		assert.Equal(t, peer.String(), newPeer.String())
	default:
		assert.Error(t, errors.New("no new peer event"))
	}
}

func assertNewPeerEvents(t *testing.T, expCount int, connChan <-chan crypto.PublicKey) {
	//var actCount int
	//loop:
	//for {
	//	select {
	//	case _ = <-connChan:
	//		actCount++
	//	default:
	//		break loop
	//	}
	//}
	assert.Equal(t, expCount, len(connChan))
}

func assertNoNewPeerEvent(t *testing.T, eventChan <-chan crypto.PublicKey) {
	select {
	case newPeer := <-eventChan:
		assert.Error(t, errors.New("unexpected new peer event, peer "+newPeer.String()))
	default:
		return
	}
}

func assertNewDisconnectedPeerEvent(t *testing.T, peer crypto.PublicKey, discChan <-chan crypto.PublicKey) {
	select {
	case newPeer := <-discChan:
		assert.Equal(t, peer.String(), newPeer.String())
	default:
		assert.Error(t, errors.New("no new peer event"))
	}
}

func assertNoNewDisconnectedPeerEvent(t *testing.T, eventChan <-chan crypto.PublicKey) {
	select {
	case newPeer := <-eventChan:
		assert.Error(t, errors.New("unexpected new peer event, peer "+newPeer.String()))
	default:
		return
	}
}

func drainPeerEvents(eventChan <-chan crypto.PublicKey) {
loop:
	for {
		select {
		case <-eventChan:
			continue loop
		default:
			break loop
		}
	}
}

func Test_Swarm_getMorePeers(t *testing.T) {
	// test normal flow
	numpeers := 3
	cfg := config.DefaultConfig()
	cfg.SwarmConfig.Bootstrap = false
	cfg.SwarmConfig.Gossip = false
	cfg.SwarmConfig.RandomConnections = numpeers
	n := p2pTestNoStart(t, cfg)

	conn, _ := n.SubscribePeerEvents()

	res := n.getMorePeers(0) // this should'nt work
	assert.Equal(t, res, 0)
	assertNoNewPeerEvent(t, conn)

	mdht := new(dht.MockDHT)
	n.dht = mdht
	// this will return 0 peers because SelectPeers returns empty array when not set

	res = n.getMorePeers(10)
	assert.Equal(t, res, 0)
	assertNoNewPeerEvent(t, conn)

	testNode := node.GenerateRandomNodeData()
	mdht.SelectPeersFunc = func(qty int) []node.Node {
		return []node.Node{testNode}
	}

	cpm := new(cpoolMock)

	// test connection error
	cpm.f = func(address string, pk crypto.PublicKey) (net.Connection, error) {
		return nil, errors.New("can't make connection")
	}

	n.cPool = cpm
	res = n.getMorePeers(1) // this should'nt work
	assert.Equal(t, res, 0)
	cpm.f = nil // for next tests
	assertNoNewPeerEvent(t, conn)

	res = n.getMorePeers(1)
	assert.Equal(t, 1, res)
	assert.Equal(t, len(n.outpeers), 1)
	assert.True(t, n.hasOutgoingPeer(testNode.PublicKey()))
	assertNewPeerEvents(t, 1, conn)
	assertNewPeerEvent(t, testNode.PublicKey(), conn)

	drainPeerEvents(conn)

	//todo remove the peer instead of counting plus one
	//
	//
	mdht.SelectPeersFunc = func(qty int) []node.Node {
		return node.GenerateRandomNodesData(qty)
	}

	res = n.getMorePeers(numpeers)
	assert.Equal(t, res, numpeers)
	assert.Equal(t, len(n.outpeers), numpeers+1) // there's already one inside
	assertNewPeerEvents(t, numpeers, conn)
	drainPeerEvents(conn) // so they wont interrupt next test
	//test inc peer
	nd := node.GenerateRandomNodeData()
	n.addIncomingPeer(nd.PublicKey())

	assert.True(t, n.hasIncomingPeer(nd.PublicKey()))
	assertNewPeerEvents(t, 1, conn)
	assertNewPeerEvent(t, nd.PublicKey(), conn)

	//test replacing inc peer
	//
	mdht.SelectPeersFunc = func(count int) []node.Node {
		some := node.GenerateRandomNodesData(count - 1)
		some = append(some, nd)
		return some
	}

	res = n.getMorePeers(numpeers)
	assert.Equal(t, res, numpeers-1)
	assert.False(t, n.hasOutgoingPeer(nd.PublicKey()))
	assert.True(t, n.hasIncomingPeer(nd.PublicKey()))
}

func TestNeighborhood_Initial(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.SwarmConfig.RandomConnections = 3
	cfg.SwarmConfig.Gossip = true
	cfg.SwarmConfig.Bootstrap = false

	p := p2pTestNoStart(t, cfg)
	mdht := new(dht.MockDHT)
	mdht.SelectPeersFunc = func(qty int) []node.Node {
		return node.GenerateRandomNodesData(qty)
	}

	p.dht = mdht

	err := p.Start()
	assert.NoError(t, err)
	ti := time.After(time.Millisecond)
	select {
	case <-p.initial:
		t.Error("Start succeded")
	case <-ti:
		break
	}

	p.Shutdown()

	p = p2pTestNoStart(t, cfg)
	p.dht = mdht
	cpm := new(cpoolMock)
	cpm.f = func(address string, pk crypto.PublicKey) (net.Connection, error) {
		return net.NewConnectionMock(pk), nil
	}
	p.cPool = cpm

	err = p.Start()
	assert.NoError(t, err)
	ti = time.After(time.Second * 1)
	select {
	case <-p.initial:
		break
	case <-ti:
		t.Error("Start succeded")
	}
}

func TestNeighborhood_Disconnect(t *testing.T) {
	n := p2pTestNoStart(t, config.DefaultConfig())
	_, disc := n.SubscribePeerEvents()
	rnd := node.GenerateRandomNodeData()
	n.addIncomingPeer(rnd.PublicKey())

	assert.True(t, n.hasIncomingPeer(rnd.PublicKey()))
	n.Disconnect(rnd.PublicKey())
	assertNewDisconnectedPeerEvent(t, rnd.PublicKey(), disc)
	ti := time.After(time.Millisecond)
	select {
	case <-n.morePeersReq:
		t.Error("got more peers on inbound")
	case <-ti:
		break
	}
	assert.False(t, n.hasIncomingPeer(rnd.PublicKey()))

	// manualy add an incoming peer
	rnd2 := node.GenerateRandomNodeData()
	n.outpeers[rnd2.PublicKey().String()] = rnd2.PublicKey() // no need to lock nothing's happening
	go n.Disconnect(rnd2.PublicKey())
	ti = time.After(time.Millisecond)
	select {
	case <-n.morePeersReq:
		break
	case <-ti:
		t.Error("didnt get morepeers")
	}
	assertNewDisconnectedPeerEvent(t, rnd2.PublicKey(), disc)
}

func TestSwarm_AddIncomingPeer(t *testing.T) {
	p := p2pTestInstance(t, config.DefaultConfig())
	rnd := node.GenerateRandomNodeData()
	p.addIncomingPeer(rnd.PublicKey())

	p.inpeersMutex.RLock()
	peer, ok := p.inpeers[rnd.PublicKey().String()]
	p.inpeersMutex.RUnlock()

	assert.True(t, ok)
	assert.NotNil(t, peer)
}

package node

import (
	"context"
	"testing"
	"time"

	peerstore "gx/ipfs/QmZR2XWVVBCtbgBWnQhWk2xcQfaR3W8faQPriAiaaj7rsr/go-libp2p-peerstore"

	"github.com/filecoin-project/go-filecoin/core"
	"github.com/filecoin-project/go-filecoin/types"
	"github.com/stretchr/testify/assert"
)

func connect(t *testing.T, nd1, nd2 *Node) {
	t.Helper()
	pinfo := peerstore.PeerInfo{
		ID:    nd2.Host.ID(),
		Addrs: nd2.Host.Addrs(),
	}

	if err := nd1.Host.Connect(context.Background(), pinfo); err != nil {
		t.Fatal(err)
	}
}

func stopNodes(nds []*Node) {
	for _, nd := range nds {
		nd.Stop()
	}
}

func startNodes(t *testing.T, nds []*Node) {
	t.Helper()
	for _, nd := range nds {
		if err := nd.Start(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestBlockPropTwoNodes(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	assert := assert.New(t)

	nodes := MakeNodesUnstarted(t, 2, false, true)
	startNodes(t, nodes)
	defer stopNodes(nodes)
	connect(t, nodes[0], nodes[1])

	baseBlk := core.RequireBestBlock(nodes[0].ChainMgr, t)
	nextBlk := &types.Block{
		Parents:           types.NewSortedCidSet(baseBlk.Cid()),
		Height:            types.Uint64(1),
		ParentWeightNum:   types.Uint64(10),
		ParentWeightDenom: types.Uint64(1),
		StateRoot:         baseBlk.StateRoot,
	}

	// Wait for network connection notifications to propagate
	time.Sleep(time.Millisecond * 75)

	assert.NoError(nodes[0].AddNewBlock(ctx, nextBlk))

	time.Sleep(time.Millisecond * 75)

	otherBest := core.RequireBestBlock(nodes[1].ChainMgr, t)
	assert.Equal(otherBest.Cid(), nextBlk.Cid(), "Blocks not equal: %#+v, %#+v", otherBest, nextBlk)
}

func TestChainSync(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	assert := assert.New(t)

	nodes := MakeNodesUnstarted(t, 2, false, true)
	startNodes(t, nodes)
	defer stopNodes(nodes)

	baseBlk := core.RequireBestBlock(nodes[0].ChainMgr, t)
	nextBlk1 := &types.Block{
		Parents:           types.NewSortedCidSet(baseBlk.Cid()),
		Height:            types.Uint64(1),
		ParentWeightNum:   types.Uint64(10),
		ParentWeightDenom: types.Uint64(1),
		StateRoot:         baseBlk.StateRoot,
	}
	nextBlk2 := &types.Block{
		Parents:           types.NewSortedCidSet(nextBlk1.Cid()),
		Height:            types.Uint64(2),
		ParentWeightNum:   types.Uint64(20),
		ParentWeightDenom: types.Uint64(1),
		StateRoot:         baseBlk.StateRoot,
	}
	nextBlk3 := &types.Block{
		Parents:           types.NewSortedCidSet(nextBlk2.Cid()),
		Height:            types.Uint64(3),
		ParentWeightNum:   types.Uint64(30),
		ParentWeightDenom: types.Uint64(1),
		StateRoot:         baseBlk.StateRoot,
	}

	assert.NoError(nodes[0].AddNewBlock(ctx, nextBlk1))
	assert.NoError(nodes[0].AddNewBlock(ctx, nextBlk2))
	assert.NoError(nodes[0].AddNewBlock(ctx, nextBlk3))

	connect(t, nodes[0], nodes[1])

	time.Sleep(time.Millisecond * 50)
	otherBest := core.RequireBestBlock(nodes[1].ChainMgr, t)
	assert.Equal(otherBest.Cid(), nextBlk3.Cid())
}
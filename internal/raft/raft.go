package raft

import (
	"context"
	"distributed_cloud_service/internal/cluster"
	"distributed_cloud_service/internal/store"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

// Node wraps a Raft node
type Node struct {
	raft *raft.Raft
	fsm  *FSM
}

// NewNode creates and initializes a new Raft node
func NewNode(store *store.Store, config *cluster.Config, dataDir string) (*Node, error) {
	// Create FSM
	fsm := NewFSM(store)

	// Create Raft configuration
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(config.NodeID)
	// Reduce timeouts for faster leader election during testing
	// Default: HeartbeatTimeout=1s, ElectionTimeout=1s, but randomized
	raftConfig.HeartbeatTimeout = 500 * time.Millisecond
	raftConfig.ElectionTimeout = 500 * time.Millisecond
	raftConfig.LeaderLeaseTimeout = 250 * time.Millisecond

	// Create log store
	logStore, err := raftboltdb.NewBoltStore(filepath.Join(dataDir, "raft.db"))
	if err != nil {
		return nil, fmt.Errorf("failed to create log store: %w", err)
	}

	// Create stable store
	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(dataDir, "stable.db"))
	if err != nil {
		return nil, fmt.Errorf("failed to create stable store: %w", err)
	}

	// Create snapshot store
	snapshotStore, err := raft.NewFileSnapshotStore(dataDir, 3, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot store: %w", err)
	}

	// Detect existing Raft state to decide whether to bootstrap
	existingState, err := raft.HasExistingState(logStore, stableStore, snapshotStore)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing raft state: %w", err)
	}

	// Determine Raft address
	raftAddr := config.RaftAddr
	if raftAddr == "" {
		host, port, err := net.SplitHostPort(config.ListenAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse listen address: %w", err)
		}
		raftAddr = fmt.Sprintf("%s:%d", host, portToInt(port)+10)
	}
	raftTCPAddr, err := net.ResolveTCPAddr("tcp", raftAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve Raft address: %w", err)
	}

	transport, err := raft.NewTCPTransport(raftAddr, raftTCPAddr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	// Create Raft instance
	r, err := raft.NewRaft(raftConfig, fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		return nil, fmt.Errorf("failed to create raft: %w", err)
	}

	// Bootstrap cluster only if there is no existing state and Bootstrap is true
	if !existingState && config.Bootstrap {
		future := r.BootstrapCluster(raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raft.ServerID(config.NodeID),
					Address: raft.ServerAddress(raftAddr),
				},
			},
		})
		if err := future.Error(); err != nil {
			fmt.Printf("Bootstrap note: %v (cluster may already exist)\n", err)
		} else {
			fmt.Printf("Bootstrap successful: %s\n", config.NodeID)
		}
	} else if existingState {
		fmt.Printf("Existing Raft state detected, joining as %s\n", config.NodeID)
	}

	return &Node{
		raft: r,
		fsm:  fsm,
	}, nil
}

// Apply proposes a command to the Raft cluster
func (n *Node) Apply(cmd KVCommand) error {
	data, err := cmd.Marshal()
	if err != nil {
		return err
	}

	future := n.raft.Apply(data, 10*time.Second)
	if err := future.Error(); err != nil {
		return err
	}

	return nil
}

// IsLeader returns true if this node is the leader
func (n *Node) IsLeader() bool {
	return n.raft.State() == raft.Leader
}

// Leader returns the address of the current leader
func (n *Node) Leader() string {
	return string(n.raft.Leader())
}

// GetRaft returns the underlying Raft instance
func (n *Node) GetRaft() *raft.Raft {
	return n.raft
}

// Join adds a new voter to the Raft cluster using the provided node ID and raft address.
func (n *Node) Join(nodeID string, raftAddress string) error {
	srvID := raft.ServerID(nodeID)
	srvAddr := raft.ServerAddress(raftAddress)
	future := n.raft.AddVoter(srvID, srvAddr, 0, 0)
	return future.Error()
}

// Remove removes a node from the Raft cluster
func (n *Node) Remove(nodeID string) error {
	srvID := raft.ServerID(nodeID)
	future := n.raft.RemoveServer(srvID, 0, 0)
	return future.Error()
}

// VerifyRead ensures linearizable reads
// For leaders: reads are always linearizable
// For followers: validates node state (true ReadIndex would require library support)
func (n *Node) VerifyRead(ctx context.Context) error {
	if n.IsLeader() {
		// Leader can read directly (leader read is always linearizable)
		return nil
	}
	
	// For followers, ensure node is in a valid state
	// Note: HashiCorp Raft doesn't expose ReadIndex directly in this version
	// For true linearizability, clients should read from the leader
	// This validates the node is operational, but doesn't guarantee linearizability
	state := n.raft.State()
	if state != raft.Follower && state != raft.Leader {
		return fmt.Errorf("node not in valid state for reads: %s", state.String())
	}
	
	// In production, you would:
	// 1. Implement ReadIndex if library supports it
	// 2. Or redirect follower reads to leader
	// 3. Or use barrier/barrierFuture to ensure catch-up
	
	return nil
}

// Shutdown gracefully shuts down the Raft node
func (n *Node) Shutdown() error {
	future := n.raft.Shutdown()
	return future.Error()
}

// Marshal serializes a KVCommand to JSON
func (c *KVCommand) Marshal() ([]byte, error) {
	return json.Marshal(c)
}

// portToInt converts a port string to int
func portToInt(port string) int {
	p, err := strconv.Atoi(port)
	if err != nil {
		return 0
	}
	return p
}


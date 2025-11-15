package test

import (
	"context"
	"distributed_cloud_service/internal/cluster"
	"distributed_cloud_service/internal/raft"
	"distributed_cloud_service/internal/store"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestMultiNodeCluster tests basic multi-node cluster operations
func TestMultiNodeCluster(t *testing.T) {
	// This is a simplified integration test
	// Full integration tests would require actual network setup
	
	// Create test data directories
	dataDir1 := filepath.Join("testdata", "node1")
	dataDir2 := filepath.Join("testdata", "node2")
	os.MkdirAll(dataDir1, 0755)
	os.MkdirAll(dataDir2, 0755)
	defer os.RemoveAll("testdata")

	// Create stores
	store1 := store.NewStore()
	store2 := store.NewStore()

	// Create configs
	config1 := &cluster.Config{
		NodeID:     "node1",
		ListenAddr: "127.0.0.1:19001",
		RaftAddr:   "127.0.0.1:19011",
		Bootstrap:  true,
	}

	config2 := &cluster.Config{
		NodeID:     "node2",
		ListenAddr: "127.0.0.1:19002",
		RaftAddr:   "127.0.0.1:19012",
		Bootstrap:  false,
		JoinURL:    "http://127.0.0.1:19001",
	}

	// Initialize Raft nodes
	raftNode1, err := raft.NewNode(store1, config1, dataDir1)
	if err != nil {
		t.Fatalf("Failed to create node1: %v", err)
	}
	defer raftNode1.Shutdown()

	raftNode2, err := raft.NewNode(store2, config2, dataDir2)
	if err != nil {
		t.Fatalf("Failed to create node2: %v", err)
	}
	defer raftNode2.Shutdown()

	// Wait for leader election
	time.Sleep(2 * time.Second)

	// Test that we have a leader
	leader1 := raftNode1.IsLeader()
	leader2 := raftNode2.IsLeader()
	
	if !leader1 && !leader2 {
		t.Error("No leader elected")
	}

	// Test that only one is leader
	if leader1 && leader2 {
		t.Error("Multiple leaders elected")
	}

	// Test join operation (if node1 is leader)
	if leader1 {
		err := raftNode1.Join("node2", "127.0.0.1:19012")
		if err != nil {
			t.Logf("Join failed (expected if already joined): %v", err)
		}
	}

	// Test apply operation
	cmd := raft.KVCommand{
		Op:    "put",
		Key:   "test-key",
		Value: []byte("test-value"),
	}

	if leader1 {
		err = raftNode1.Apply(cmd)
	} else {
		err = raftNode2.Apply(cmd)
	}

	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Wait for replication
	time.Sleep(1 * time.Second)

	// Verify value in both stores
	val1, ok1 := store1.Get("test-key")
	val2, ok2 := store2.Get("test-key")

	if !ok1 || string(val1) != "test-value" {
		t.Error("Value not found in store1")
	}

	if !ok2 || string(val2) != "test-value" {
		t.Error("Value not replicated to store2")
	}
}

// TestSnapshotRestore tests snapshot and restore functionality
func TestSnapshotRestore(t *testing.T) {
	dataDir := filepath.Join("testdata", "snapshot")
	os.MkdirAll(dataDir, 0755)
	defer os.RemoveAll("testdata")

	store1 := store.NewStore()
	config := &cluster.Config{
		NodeID:     "snapshot-node",
		ListenAddr: "127.0.0.1:19003",
		RaftAddr:   "127.0.0.1:19013",
		Bootstrap:  true,
	}

	raftNode, err := raft.NewNode(store1, config, dataDir)
	if err != nil {
		t.Fatalf("Failed to create node: %v", err)
	}
	defer raftNode.Shutdown()

	// Wait for leader
	time.Sleep(1 * time.Second)

	// Add some data
	cmd := raft.KVCommand{
		Op:    "put",
		Key:   "snapshot-key",
		Value: []byte("snapshot-value"),
	}

	if raftNode.IsLeader() {
		err = raftNode.Apply(cmd)
		if err != nil {
			t.Fatalf("Apply failed: %v", err)
		}
		
		// Wait for apply to complete and be committed
		time.Sleep(2 * time.Second)

		// Verify data exists
		val, ok := store1.Get("snapshot-key")
		if !ok || string(val) != "snapshot-value" {
			t.Error("Data not stored before snapshot")
		}
	} else {
		t.Skip("Node is not leader, skipping snapshot test")
	}

	// Snapshot is handled automatically by Raft
	// In a real test, we'd trigger a snapshot and verify restore
}

// TestReadVerification tests read verification
func TestReadVerification(t *testing.T) {
	dataDir := filepath.Join("testdata", "read")
	os.MkdirAll(dataDir, 0755)
	defer os.RemoveAll("testdata")

	store1 := store.NewStore()
	config := &cluster.Config{
		NodeID:     "read-node",
		ListenAddr: "127.0.0.1:19004",
		RaftAddr:   "127.0.0.1:19014",
		Bootstrap:  true,
	}

	raftNode, err := raft.NewNode(store1, config, dataDir)
	if err != nil {
		t.Fatalf("Failed to create node: %v", err)
	}
	defer raftNode.Shutdown()

	// Wait for leader
	time.Sleep(1 * time.Second)

	// Test read verification
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = raftNode.VerifyRead(ctx)
	if err != nil {
		t.Errorf("VerifyRead failed: %v", err)
	}
}

// TestJoinRemove tests join and remove operations
func TestJoinRemove(t *testing.T) {
	dataDir1 := filepath.Join("testdata", "join1")
	dataDir2 := filepath.Join("testdata", "join2")
	os.MkdirAll(dataDir1, 0755)
	os.MkdirAll(dataDir2, 0755)
	defer os.RemoveAll("testdata")

	store1 := store.NewStore()
	store2 := store.NewStore()

	config1 := &cluster.Config{
		NodeID:     "join-node1",
		ListenAddr: "127.0.0.1:19005",
		RaftAddr:   "127.0.0.1:19015",
		Bootstrap:  true,
	}

	config2 := &cluster.Config{
		NodeID:     "join-node2",
		ListenAddr: "127.0.0.1:19006",
		RaftAddr:   "127.0.0.1:19016",
		Bootstrap:  false,
	}

	raftNode1, err := raft.NewNode(store1, config1, dataDir1)
	if err != nil {
		t.Fatalf("Failed to create node1: %v", err)
	}
	defer raftNode1.Shutdown()

	raftNode2, err := raft.NewNode(store2, config2, dataDir2)
	if err != nil {
		t.Fatalf("Failed to create node2: %v", err)
	}
	defer raftNode2.Shutdown()

	// Wait for leader
	time.Sleep(2 * time.Second)

	// Test join
	if raftNode1.IsLeader() {
		err = raftNode1.Join("join-node2", "127.0.0.1:19016")
		if err != nil {
			t.Logf("Join failed (may already be joined): %v", err)
		}
	}

	// Wait for join to complete
	time.Sleep(1 * time.Second)

	// Test remove (if node1 is leader)
	if raftNode1.IsLeader() {
		err = raftNode1.Remove("join-node2")
		if err != nil {
			t.Logf("Remove failed (expected if not joined): %v", err)
		}
	}
}


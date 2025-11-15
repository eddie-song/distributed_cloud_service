package raft

import (
	"distributed_cloud_service/internal/store"
	"encoding/json"
	"testing"

	"github.com/hashicorp/raft"
)

func TestFSM_Apply(t *testing.T) {
	kvStore := store.NewStore()
	fsm := NewFSM(kvStore)

	// Test PUT operation
	putCmd := KVCommand{
		Op:    "put",
		Key:   "test-key",
		Value: []byte("test-value"),
	}
	putData, _ := json.Marshal(putCmd)
	
	logEntry := &raft.Log{
		Index: 1,
		Term:  1,
		Type:  raft.LogCommand,
		Data:  putData,
	}

	result := fsm.Apply(logEntry)
	if result != nil {
		t.Errorf("Apply() returned error: %v", result)
	}

	// Verify value was stored
	val, ok := kvStore.Get("test-key")
	if !ok {
		t.Fatal("Key not found after Apply(PUT)")
	}
	if string(val) != "test-value" {
		t.Errorf("Expected 'test-value', got '%s'", string(val))
	}

	// Test DELETE operation
	deleteCmd := KVCommand{
		Op:  "delete",
		Key: "test-key",
	}
	deleteData, _ := json.Marshal(deleteCmd)
	
	logEntry2 := &raft.Log{
		Index: 2,
		Term:  1,
		Type:  raft.LogCommand,
		Data:  deleteData,
	}

	result = fsm.Apply(logEntry2)
	if result != nil {
		t.Errorf("Apply() returned error: %v", result)
	}

	// Verify key was deleted
	_, ok = kvStore.Get("test-key")
	if ok {
		t.Error("Key still exists after Apply(DELETE)")
	}

	// Test unknown operation
	unknownCmd := KVCommand{
		Op:  "unknown",
		Key: "test-key",
	}
	unknownData, _ := json.Marshal(unknownCmd)
	
	logEntry3 := &raft.Log{
		Index: 3,
		Term:  1,
		Type:  raft.LogCommand,
		Data:  unknownData,
	}

	result = fsm.Apply(logEntry3)
	if result == nil {
		t.Error("Apply() should return error for unknown operation")
	}
}

func TestFSM_Snapshot(t *testing.T) {
	kvStore := store.NewStore()
	fsm := NewFSM(kvStore)

	// Add some data
	kvStore.Put("key1", []byte("value1"))
	kvStore.Put("key2", []byte("value2"))

	// Create snapshot
	snapshot, err := fsm.Snapshot()
	if err != nil {
		t.Fatalf("Snapshot() failed: %v", err)
	}
	if snapshot == nil {
		t.Fatal("Snapshot() returned nil")
	}
}

func TestFSM_Restore(t *testing.T) {
	kvStore := store.NewStore()
	fsm := NewFSM(kvStore)

	// Create state to restore
	state := map[string][]byte{
		"key1": []byte("value1"),
		"key2": []byte("value2"),
	}
	stateData, _ := json.Marshal(state)

	// Create a mock ReadCloser
	reader := &mockReadCloser{data: stateData}

	// Restore
	err := fsm.Restore(reader)
	if err != nil {
		t.Fatalf("Restore() failed: %v", err)
	}

	// Verify restored data
	val1, ok := kvStore.Get("key1")
	if !ok || string(val1) != "value1" {
		t.Error("key1 not restored correctly")
	}

	val2, ok := kvStore.Get("key2")
	if !ok || string(val2) != "value2" {
		t.Error("key2 not restored correctly")
	}
}

// mockReadCloser implements io.ReadCloser for testing
type mockReadCloser struct {
	data []byte
	pos  int
}

func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	if m.pos >= len(m.data) {
		return 0, nil // EOF
	}
	n = copy(p, m.data[m.pos:])
	m.pos += n
	return n, nil
}

func (m *mockReadCloser) Close() error {
	return nil
}


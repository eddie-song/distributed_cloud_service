package http

import (
	"bytes"
	"context"
	"distributed_cloud_service/internal/raft"
	"distributed_cloud_service/internal/store"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockRaftNode implements RaftNode interface for testing
type mockRaftNode struct {
	isLeader bool
	leader   string
	store    *store.Store
}

func (m *mockRaftNode) IsLeader() bool {
	return m.isLeader
}

func (m *mockRaftNode) Leader() string {
	return m.leader
}

func (m *mockRaftNode) Apply(cmd raft.KVCommand) error {
	// Apply directly to store for testing
	switch cmd.Op {
	case "put":
		m.store.Put(cmd.Key, cmd.Value)
	case "delete":
		m.store.Delete(cmd.Key)
	}
	return nil
}

func (m *mockRaftNode) VerifyRead(ctx context.Context) error {
	return nil
}

func TestHandlePut(t *testing.T) {
	kvStore := store.NewStore()
	mockRaft := &mockRaftNode{isLeader: true, store: kvStore}
	server := NewServer(kvStore, mockRaft)

	// Test successful PUT
	req := httptest.NewRequest("PUT", "/kv/test-key", bytes.NewBufferString("test-value"))
	w := httptest.NewRecorder()
	server.HandlePut(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}

	// Verify value was stored
	val, ok := kvStore.Get("test-key")
	if !ok {
		t.Fatal("Key not found after PUT")
	}
	if string(val) != "test-value" {
		t.Errorf("Expected 'test-value', got '%s'", string(val))
	}
}

func TestHandlePut_NotLeader(t *testing.T) {
	kvStore := store.NewStore()
	mockRaft := &mockRaftNode{isLeader: false, leader: "127.0.0.1:9001", store: kvStore}
	server := NewServer(kvStore, mockRaft)

	req := httptest.NewRequest("PUT", "/kv/test-key", bytes.NewBufferString("test-value"))
	w := httptest.NewRecorder()
	server.HandlePut(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	leader := w.Header().Get("X-Leader")
	if leader != "127.0.0.1:9001" {
		t.Errorf("Expected X-Leader header '127.0.0.1:9001', got '%s'", leader)
	}
}

func TestHandleGet(t *testing.T) {
	kvStore := store.NewStore()
	kvStore.Put("test-key", []byte("test-value"))
	mockRaft := &mockRaftNode{isLeader: true, store: kvStore}
	server := NewServer(kvStore, mockRaft)

	req := httptest.NewRequest("GET", "/kv/test-key", nil)
	w := httptest.NewRecorder()
	server.HandleGet(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != "test-value" {
		t.Errorf("Expected body 'test-value', got '%s'", w.Body.String())
	}
}

func TestHandleGet_NotFound(t *testing.T) {
	kvStore := store.NewStore()
	mockRaft := &mockRaftNode{isLeader: true, store: kvStore}
	server := NewServer(kvStore, mockRaft)

	req := httptest.NewRequest("GET", "/kv/nonexistent", nil)
	w := httptest.NewRecorder()
	server.HandleGet(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestHandleDelete(t *testing.T) {
	kvStore := store.NewStore()
	kvStore.Put("test-key", []byte("test-value"))
	mockRaft := &mockRaftNode{isLeader: true, store: kvStore}
	server := NewServer(kvStore, mockRaft)

	req := httptest.NewRequest("DELETE", "/kv/test-key", nil)
	w := httptest.NewRecorder()
	server.HandleDelete(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}

	// Verify key was deleted
	_, ok := kvStore.Get("test-key")
	if ok {
		t.Error("Key still exists after DELETE")
	}
}

func TestHandleDelete_NotFound(t *testing.T) {
	kvStore := store.NewStore()
	mockRaft := &mockRaftNode{isLeader: true, store: kvStore}
	server := NewServer(kvStore, mockRaft)

	req := httptest.NewRequest("DELETE", "/kv/nonexistent", nil)
	w := httptest.NewRecorder()
	server.HandleDelete(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}


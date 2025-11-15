package http

import (
	"context"
	"distributed_cloud_service/internal/raft"
	"distributed_cloud_service/internal/store"
	"io"
	"net/http"
	"time"
)

// Server handles HTTP requests for the key-value store
type Server struct {
	store *store.Store
	raft  RaftNode
}

// NewServer creates a new HTTP server
func NewServer(s *store.Store, r RaftNode) *Server {
	return &Server{
		store: s,
		raft:  r,
	}
}

// HandlePut handles PUT /kv/{key} requests
func (s *Server) HandlePut(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Only leader can accept writes
	if !s.raft.IsLeader() {
		leader := s.raft.Leader()
		if leader != "" {
			w.Header().Set("X-Leader", leader)
		}
		http.Error(w, "Not the leader", http.StatusBadRequest)
		return
	}

	key := r.URL.Path[len("/kv/"):]
	if key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	val, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Propose command to Raft
	cmd := raft.KVCommand{
		Op:    "put",
		Key:   key,
		Value: val,
	}

	if err := s.raft.Apply(cmd); err != nil {
		http.Error(w, "Failed to apply command: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleGet handles GET /kv/{key} requests with linearizable reads
func (s *Server) HandleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Path[len("/kv/"):]
	if key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	// Ensure linearizable read using ReadIndex (for followers)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	
	if err := s.raft.VerifyRead(ctx); err != nil {
		http.Error(w, "Failed to verify read: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Now safe to read from local store
	val, ok := s.store.Get(key)
	if !ok {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	w.Write(val)
}

// HandleDelete handles DELETE /kv/{key} requests
func (s *Server) HandleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Only leader can accept writes
	if !s.raft.IsLeader() {
		leader := s.raft.Leader()
		if leader != "" {
			w.Header().Set("X-Leader", leader)
		}
		http.Error(w, "Not the leader", http.StatusBadRequest)
		return
	}

	key := r.URL.Path[len("/kv/"):]
	if key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	// Check if key exists first (read from store is okay)
	_, ok := s.store.Get(key)
	if !ok {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}

	// Propose command to Raft
	cmd := raft.KVCommand{
		Op:  "delete",
		Key: key,
	}

	if err := s.raft.Apply(cmd); err != nil {
		http.Error(w, "Failed to apply command: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}


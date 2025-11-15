package raft

import (
	"distributed_cloud_service/internal/store"
	"encoding/json"
	"io"

	"github.com/hashicorp/raft"
)

// KVCommand represents a command to be applied via Raft
type KVCommand struct {
	Op    string `json:"op"`    // "put", "delete"
	Key   string `json:"key"`
	Value []byte `json:"value,omitempty"`
}

// FSM is the finite state machine that applies commands to the store
type FSM struct {
	store *store.Store
}

// NewFSM creates a new FSM
func NewFSM(s *store.Store) *FSM {
	return &FSM{
		store: s,
	}
}

// Apply applies a Raft log entry to the FSM
func (f *FSM) Apply(logEntry *raft.Log) interface{} {
	var cmd KVCommand
	if err := json.Unmarshal(logEntry.Data, &cmd); err != nil {
		return err
	}

	switch cmd.Op {
	case "put":
		f.store.Put(cmd.Key, cmd.Value)
		return nil
	case "delete":
		f.store.Delete(cmd.Key)
		return nil
	default:
		return "unknown command"
	}
}

// Snapshot returns a snapshot of the current state
func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	state := f.store.Dump()
	return &snapshot{state: state}, nil
}

// Restore restores from a snapshot
func (f *FSM) Restore(snapshot io.ReadCloser) error {
	defer snapshot.Close()
	var state map[string][]byte
	if err := json.NewDecoder(snapshot).Decode(&state); err != nil {
		return err
	}
	f.store.Load(state)
	return nil
}

// snapshot is a simple snapshot implementation
type snapshot struct{
	state map[string][]byte
}

func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	enc := json.NewEncoder(sink)
	if err := enc.Encode(s.state); err != nil {
		_ = sink.Cancel()
		return err
	}
	return sink.Close()
}

func (s *snapshot) Release() {}


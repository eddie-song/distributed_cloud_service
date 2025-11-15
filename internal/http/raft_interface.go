package http

import (
	"context"
	"distributed_cloud_service/internal/raft"
)

// RaftNode defines the interface for Raft operations needed by HTTP handlers
type RaftNode interface {
	IsLeader() bool
	Leader() string
	Apply(cmd raft.KVCommand) error
	VerifyRead(ctx context.Context) error
}


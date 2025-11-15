package cluster

import (
	"encoding/json"
	"net/http"
)

// StatusResponse represents cluster status information
type StatusResponse struct {
	NodeID  string   `json:"node_id"`
	Address string   `json:"address"`
	Peers   []string `json:"peers"`
}

// HandleStatus returns the current node's status
func (c *Cluster) HandleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	peerAddrs := make([]string, len(c.Peers))
	for i, peer := range c.Peers {
		peerAddrs[i] = peer.Address
	}

	response := StatusResponse{
		NodeID:  c.Self.ID,
		Address: c.Self.Address,
		Peers:   peerAddrs,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleMembers returns all cluster members
func (c *Cluster) HandleMembers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	members := c.GetMembers()
	memberList := make([]map[string]string, len(members))
	for i, member := range members {
		memberList[i] = map[string]string{
			"id":      member.ID,
			"address": member.Address,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"members": memberList,
	})
}


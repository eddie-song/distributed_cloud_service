package cluster

// Config represents a node's configuration
type Config struct {
	NodeID     string   `yaml:"node_id"`
	ListenAddr string   `yaml:"listen_addr"` // HTTP address
	RaftAddr   string   `yaml:"raft_addr"`   // Optional explicit Raft address
	Peers      []string `yaml:"peers"`
	Bootstrap  bool     `yaml:"bootstrap"`   // Only first node should set true
	JoinURL    string   `yaml:"join_url"`    // Leader HTTP base for auto-join (e.g., http://127.0.0.1:9001)
	AuthToken  string   `yaml:"auth_token"`  // Optional bearer token for write operations
}

// Node represents a node in the cluster
type Node struct {
	ID      string
	Address string
}

// Cluster manages membership and node information
type Cluster struct {
	Self   Node
	Peers  []Node
	Config *Config
}

// NewCluster creates a new cluster instance
func NewCluster(config *Config) *Cluster {
	cluster := &Cluster{
		Config: config,
		Self: Node{
			ID:      config.NodeID,
			Address: config.ListenAddr,
		},
		Peers: make([]Node, 0),
	}

	// Convert peer addresses to Node objects
	for _, addr := range config.Peers {
		cluster.Peers = append(cluster.Peers, Node{
			Address: addr,
		})
	}

	return cluster
}

// GetMembers returns all cluster members (self + peers)
func (c *Cluster) GetMembers() []Node {
	members := []Node{c.Self}
	members = append(members, c.Peers...)
	return members
}


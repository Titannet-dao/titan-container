package types

// NodeType node type
type NodeType int

const (
	NodeUnknown NodeType = iota

	NodeManager
	NodeProvider
)

func (n NodeType) String() string {
	switch n {
	case NodeManager:
		return "manager"
	case NodeProvider:
		return "provider"
	}

	return ""
}

// RunningNodeType represents the type of the running node.
var RunningNodeType NodeType

package api

import (
	"fmt"

	"github.com/Filecoin-Titan/titan-container/api/types"

	"golang.org/x/xerrors"
)

type Version uint32

func newVer(major, minor, patch uint8) Version {
	return Version(uint32(major)<<16 | uint32(minor)<<8 | uint32(patch))
}

// Ints returns (major, minor, patch) versions
func (ve Version) Ints() (uint32, uint32, uint32) {
	v := uint32(ve)
	return (v & majorOnlyMask) >> 16, (v & minorOnlyMask) >> 8, v & patchOnlyMask
}

func (ve Version) String() string {
	vmj, vmi, vp := ve.Ints()
	return fmt.Sprintf("%d.%d.%d", vmj, vmi, vp)
}

func (ve Version) EqMajorMinor(v2 Version) bool {
	return ve&minorMask == v2&minorMask
}

type NodeType int

const (
	NodeUnknown NodeType = iota

	NodeManager
	NodeProvider
)

var (
	ManagerAPIVersion0  = newVer(1, 0, 0)
	ProviderAPIVersion0 = newVer(1, 0, 0)
)

//nolint:varcheck,deadcode
const (
	majorMask = 0xff0000
	minorMask = 0xffff00
	patchMask = 0xffffff

	majorOnlyMask = 0xff0000
	minorOnlyMask = 0x00ff00
	patchOnlyMask = 0x0000ff
)

func VersionForType(nodeType types.NodeType) (Version, error) {
	switch nodeType {
	case types.NodeManager:
		return ManagerAPIVersion0, nil
	case types.NodeProvider:
		return ProviderAPIVersion0, nil

	default:
		return Version(0), xerrors.Errorf("unknown node type %d", nodeType)
	}
}

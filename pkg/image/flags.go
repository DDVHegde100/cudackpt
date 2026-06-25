package image

const VersionV1 = 1
const VersionV2 = 2

const (
	FlagNone       uint16 = 0
	FlagCompressed uint16 = 1 << 0
	FlagSparse     uint16 = 1 << 1
	FlagDedup      uint16 = 1 << 2
	FlagDelta      uint16 = 1 << 3
)

func HasFlag(flags uint16, bit uint16) bool {
	return flags&bit != 0
}

package tools

func MakeDev(major uint32, minor uint32) uint32 {
	var dev uint32
	dev = (major & 0x00000fff) << 8
	dev |= (major & 0xfffff000) << 32
	dev |= (minor & 0x000000ff) << 0
	dev |= (minor & 0xffffff00) << 12
	return dev
}

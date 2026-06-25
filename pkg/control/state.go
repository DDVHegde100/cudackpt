package control

func StateName(v uint32) string {
	switch v {
	case 0:
		return "idle"
	case 1:
		return "frozen"
	case 2:
		return "snapped"
	case 3:
		return "restored"
	case 4:
		return "running"
	default:
		return "unknown"
	}
}

package control

func restoreCandidates(primary int, shims []int) []int {
	var out []int
	if primary > 0 {
		out = append(out, primary)
	}
	for _, p := range shims {
		if p != primary {
			out = append(out, p)
		}
	}
	return out
}

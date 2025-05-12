package helper

func BoolToFloat(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

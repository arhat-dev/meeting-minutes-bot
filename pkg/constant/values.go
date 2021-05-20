package constant

var (
	trueVal  = true
	falseVal = false
)

func True() *bool {
	return &trueVal
}

func False() *bool {
	return &falseVal
}

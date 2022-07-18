package rt

type LoginFlow uint32

const (
	LoginFlow_None LoginFlow = 0

	LoginFlow_Token LoginFlow = 1 << (iota - 1)
	LoginFlow_Username
	LoginFlow_Password
	LoginFlow_TOTPCode
)

package numhelper

func SizeAlign[T UnsignedInteger](n, alignment T) T {
	return (n + alignment - 1) & ^(alignment - 1)
}

func SizeStart[T UnsignedInteger](n, alignment T) T {
	return n & ^(alignment - 1)
}

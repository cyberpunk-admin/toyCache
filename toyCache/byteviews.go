// Abstraction and encapsulation of cache values

package toyCache

// ByteView holds an immutable view of bytes
type ByteView struct {
	b []byte  // Supports storage of any type of data
}

// Len return the length of view
func (v ByteView) Len() int {
	return len(v.b)
}

// ByteSlice return a copy of the date as a byte slice
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// Strings return the data as a string
func (v ByteView) String() string{
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
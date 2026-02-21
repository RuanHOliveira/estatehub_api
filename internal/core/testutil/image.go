package testutil

// Retorna 512 bytes JPEG
func JpegMagicBytes() []byte {
	buf := make([]byte, 512)
	copy(buf, []byte{0xFF, 0xD8, 0xFF, 0xE0})
	return buf
}

// Retorna 512 bytes GIF 
func GifBytes() []byte {
	buf := make([]byte, 512)
	copy(buf, []byte("GIF89a"))
	return buf
}

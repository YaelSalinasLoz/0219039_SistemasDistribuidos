package log

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

// Binary encoding
var (
	enc = binary.BigEndian
)

// file limit
const (
	lenWidth = 8
)

// Store struct
type store struct {
	*os.File
	mu   sync.Mutex
	buf  *bufio.Writer // Read/Write I/O operation
	size uint64
}

func newStore(f *os.File) (*store, error) {
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}
	size := uint64(fi.Size())
	return &store{
		File: f,
		size: size,
		buf:  bufio.NewWriter(f),
	}, nil
}

// Add Store --> receive bytes
func (s *store) Append(p []byte) (n uint64, pos uint64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Determine the position in which I am going to write.
	pos = s.size
	//Write to a buffer first...
	if err := binary.Write(s.buf, enc, uint64(len(p))); err != nil {
		return 0, 0, err
	}
	// ...  and then write to the physical file.
	w, err := s.buf.Write(p)
	if err != nil {
		return 0, 0, err
	}
	// Get the new length of my file.
	w += lenWidth
	s.size += uint64(w)
	// Return the number of bytes that I wrote and the current position of my file.
	return uint64(w), pos, nil
}

func (s *store) Read(pos uint64) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Flush the data that I have not yet written in the file
	if err := s.buf.Flush(); err != nil {
		return nil, err
	}
	// How many bytes I have to read to reach my Store
	size := make([]byte, lenWidth)
	if _, err := s.File.ReadAt(size, int64(pos)); err != nil {
		return nil, err
	}
	b := make([]byte, enc.Uint64(size))
	// Get the Store of the desired position
	if _, err := s.File.ReadAt(b, int64(pos+lenWidth)); err != nil {
		return nil, err
	}
	return b, nil
}

// ReadAt --> Helper function
func (s *store) ReadAt(p []byte, off int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.buf.Flush(); err != nil {
		return 0, err
	}
	return s.File.ReadAt(p, off)
}

// Close
func (s *store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.buf.Flush()
	if err != nil {
		return err
	}
	return s.File.Close()
}

package log

import (
	"encoding/binary"
	"os"
	"sync"

	"github.com/tysonmote/gommap"
)

var (
	offWidth uint64 = 4
	posWidth uint64 = 8
	entWidth        = offWidth + posWidth
)

type index struct {
	file *os.File
	mmap gommap.MMap
	size uint64
	mu   sync.Mutex
}

func newIndex(f *os.File, c Config) (*index, error) {
	idx := &index{
		file: f,
	}

	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}
	idx.size = uint64(fi.Size())

	if err := os.Truncate(f.Name(), int64(c.Segment.MaxIndexBytes)); err != nil {
		return nil, err
	}

	if idx.mmap, err = gommap.Map(idx.file.Fd(), gommap.PROT_READ|gommap.PROT_WRITE, gommap.MAP_SHARED); err != nil {
		return nil, err
	}

	return idx, nil
}

func (i *index) Read(in int64) (out uint32, pos uint64, err error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.size == 0 {
		return 0, 0, os.ErrInvalid
	}

	if in == -1 {
		out = uint32((i.size / entWidth) - 1)
	} else {
		out = uint32(in)
	}

	pos = binary.BigEndian.Uint64(i.mmap[out*uint32(entWidth)+offWidth:])
	return out, pos, nil
}

func (i *index) Write(off uint32, pos uint64) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.size+entWidth > uint64(len(i.mmap)) {
		return os.ErrInvalid
	}

	binary.BigEndian.PutUint32(i.mmap[i.size:], off)
	binary.BigEndian.PutUint64(i.mmap[i.size+offWidth:], pos)
	i.size += entWidth
	return nil
}

func (i *index) Close() error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if err := i.mmap.Sync(gommap.MS_SYNC); err != nil {
		return err
	}

	if err := i.file.Sync(); err != nil {
		return err
	}

	if err := os.Truncate(i.file.Name(), int64(i.size)); err != nil {
		return err
	}

	return i.file.Close()
}

package database

import (
	"encoding/binary"
	"errors"
)

const SystemsRecordSize = 33

type System struct {
	hilbertKey uint64
	X, Y, Z    uint32
	Id         uint64
	StarClass  uint8
	nameOffset uint32
}

func (s *System) Marshal(buf []byte) error {
	if len(buf) < SystemsRecordSize {
		return errors.New("The provided buffer is not large enough")
	}

	binary.LittleEndian.PutUint64(buf[0:8], s.hilbertKey)
	binary.LittleEndian.PutUint32(buf[8:12], s.X)
	binary.LittleEndian.PutUint32(buf[12:16], s.Y)
	binary.LittleEndian.PutUint32(buf[16:20], s.Z)
	binary.LittleEndian.PutUint64(buf[20:28], s.Id)
	buf[28] = s.StarClass
	binary.LittleEndian.PutUint32(buf[29:33], s.nameOffset)

	return nil
}

func (s *System) Unmarshal(data []byte) error {
	if len(data) < SystemsRecordSize {
		return errors.New("The provided buffer is not large enough")
	}

	s.hilbertKey = binary.LittleEndian.Uint64(data[0:8])
	s.X = binary.LittleEndian.Uint32(data[8:12])
	s.Y = binary.LittleEndian.Uint32(data[12:16])
	s.Z = binary.LittleEndian.Uint32(data[16:20])
	s.Id = binary.LittleEndian.Uint64(data[20:28])
	s.StarClass = data[28]
	s.nameOffset = binary.LittleEndian.Uint32(data[29:33])

	return nil
}

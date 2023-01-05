package wav

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type WAV struct {
	FormatTag      uint16
	Channels       uint16
	SamplesPerSec  uint32
	AvgBytesPerSec uint32
	BlockAlign     uint16
	BitsPerSample  uint16
	Length         uint32
	Data           []byte
}

func (w *WAV) Append(wavFile *WAV) (err error) {
	if w.Channels != wavFile.Channels ||
		w.BitsPerSample != wavFile.BitsPerSample ||
		w.SamplesPerSec != wavFile.SamplesPerSec {
		err = fmt.Errorf("files are not compatible")
		return
	}
	w.Data = append(w.Data, wavFile.Data...)
	w.Length += wavFile.Length
	return
}

func (w *WAV) AppendBytes(fileBytes []byte) (err error) {
	wavFile, err := FromBytes(fileBytes)
	if err != nil {
		return
	}
	if w.Channels != wavFile.Channels ||
		w.BitsPerSample != wavFile.BitsPerSample ||
		w.SamplesPerSec != wavFile.SamplesPerSec {
		err = fmt.Errorf("files are not compatible")
		return
	}
	w.Data = append(w.Data, wavFile.Data...)
	w.Length += wavFile.Length
	return
}

const (
	WAVE_FORMAT_PCM        = 0x1
	WAVE_FORMAT_EXTENSIBLE = 0xFFFE
)

/* TODO: Replace below with this function */
func formatBytes(reader *bytes.Reader, offset int64, length int64, data any) {
	binary.Read(io.NewSectionReader(reader, offset, length), binary.LittleEndian, data)
}

func FromBytes(stream []byte) (audio WAV, err error) {

	reader := bytes.NewReader(stream)

	/* Check RIFF Header */
	var riff [4]byte
	binary.Read(io.NewSectionReader(reader, 0, 4), binary.LittleEndian, &riff)
	if string(riff[:]) != "RIFF" {
		err = fmt.Errorf("error: no RIFF header")
		return
	}

	/* Check WAVE Header */
	var wave [4]byte
	binary.Read(io.NewSectionReader(reader, 8, 4), binary.LittleEndian, &wave)
	if string(wave[:]) != "WAVE" {
		err = fmt.Errorf("error: no RIFF header")
		return
	}

	binary.Read(io.NewSectionReader(reader, 20, 2), binary.LittleEndian, &audio.FormatTag)

	if !(audio.FormatTag == WAVE_FORMAT_PCM || audio.FormatTag == WAVE_FORMAT_EXTENSIBLE) {
		err = fmt.Errorf("error: invalid format tag '%v'", audio.FormatTag)
		return
	}

	binary.Read(io.NewSectionReader(reader, 22, 2), binary.LittleEndian, &audio.Channels)
	binary.Read(io.NewSectionReader(reader, 24, 4), binary.LittleEndian, &audio.SamplesPerSec)
	binary.Read(io.NewSectionReader(reader, 28, 4), binary.LittleEndian, &audio.AvgBytesPerSec)
	binary.Read(io.NewSectionReader(reader, 32, 2), binary.LittleEndian, &audio.BlockAlign)
	binary.Read(io.NewSectionReader(reader, 34, 2), binary.LittleEndian, &audio.BitsPerSample)

	if audio.FormatTag == WAVE_FORMAT_PCM {
		binary.Read(io.NewSectionReader(reader, 40, 4), binary.LittleEndian, &audio.Length)
	} else if audio.FormatTag == WAVE_FORMAT_EXTENSIBLE {
		binary.Read(io.NewSectionReader(reader, 76, 4), binary.LittleEndian, &audio.Length)
	}

	buf := new(bytes.Buffer)
	if audio.FormatTag == WAVE_FORMAT_PCM {
		io.Copy(buf, io.NewSectionReader(reader, 44, int64(audio.Length)))
	} else if audio.FormatTag == WAVE_FORMAT_EXTENSIBLE {
		io.Copy(buf, io.NewSectionReader(reader, 80, int64(audio.Length)))
	}
	audio.Data = buf.Bytes()

	return
}

func getChannelMask(c uint16) (mask uint32) {
	if c == 1 {
		mask = 0x4
	} else if c == 2 {
		mask = 0x3 //
	} else if c == 4 {
		mask = 0x33
	} else if c == 6 {
		mask = 0x3f
	} else if c == 8 {
		mask = 0x63f
	}
	return
}

// Marshal returns audio data as WAV formatted data.
func (w *WAV) Bytes() (stream []byte, err error) {

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, []byte("RIFF"))

	if w.FormatTag == WAVE_FORMAT_PCM {
		binary.Write(buf, binary.LittleEndian, uint32(w.Length+36))
	} else if w.FormatTag == WAVE_FORMAT_EXTENSIBLE {
		binary.Write(buf, binary.LittleEndian, uint32(w.Length+72))
	} else {
		err = fmt.Errorf("error: invalid format tag")
		return
	}

	binary.Write(buf, binary.BigEndian, []byte("WAVEfmt "))

	if w.FormatTag == WAVE_FORMAT_PCM {
		binary.Write(buf, binary.LittleEndian, uint32(16))
	} else {
		binary.Write(buf, binary.LittleEndian, uint32(40))
	}

	binary.Write(buf, binary.LittleEndian, w.FormatTag)
	binary.Write(buf, binary.LittleEndian, w.Channels)
	binary.Write(buf, binary.LittleEndian, w.SamplesPerSec)
	binary.Write(buf, binary.LittleEndian, w.AvgBytesPerSec)
	binary.Write(buf, binary.LittleEndian, w.BlockAlign)
	binary.Write(buf, binary.LittleEndian, w.BitsPerSample)

	if w.FormatTag == WAVE_FORMAT_EXTENSIBLE {
		binary.Write(buf, binary.LittleEndian, uint16(22)) // cbSize
		// validBitsPerSample
		binary.Write(buf, binary.LittleEndian, w.BitsPerSample)
		// channelMask
		binary.Write(buf, binary.LittleEndian, uint32(getChannelMask(w.Channels)))
		//binary.Write(buf, binary.LittleEndian, uint16(0))            // reserved
		guid := [16]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, 0x00, 0x80, 0x00, 0x00, 0xaa, 0x00, 0x38, 0x9b, 0x71}
		binary.Write(buf, binary.BigEndian, guid)
		binary.Write(buf, binary.BigEndian, []byte("fact"))                           // fact chunk is an optional chunk
		binary.Write(buf, binary.LittleEndian, uint32(4))                             // 4 bytes
		binary.Write(buf, binary.LittleEndian, uint32(w.Length/uint32(w.BlockAlign))) // zero padding
	}

	binary.Write(buf, binary.BigEndian, []byte("data"))
	binary.Write(buf, binary.LittleEndian, w.Length)
	binary.Write(buf, binary.LittleEndian, w.Data)
	stream = buf.Bytes()

	return
}

package proto

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/golang/snappy"
)

// Snappy-encoded messages from the official Java client are encoded using
// snappy-java: see github.com/xerial/snappy-java.
// This does its own non-standard framing. We can detect this encoding
// by sniffing its special header.
//
// That library will still read plain (unframed) snappy-encoded messages,
// so we don't need to implement that codec on the compression side.
//
// (This is the same behavior as several of the other popular Kafka clients.)

var snappyJavaMagic = []byte("\x82SNAPPY\x00")

func snappyDecode(b []byte) ([]byte, error) {
	if !bytes.HasPrefix(b, snappyJavaMagic) {
		return snappy.Decode(nil, b)
	}

	// See https://github.com/xerial/snappy-java/blob/develop/src/main/java/org/xerial/snappy/SnappyInputStream.java
	version := binary.BigEndian.Uint32(b[8:12])
	if version != 1 {
		return nil, fmt.Errorf("cannot handle snappy-java codec version other than 1 (got %d)", version)
	}
	// b[12:16] is the "compatible version"; ignore for now
	var (
		decoded = make([]byte, 0, len(b))
		chunk   []byte
		err     error
	)
	for i := 16; i < len(b); {
		n := int(binary.BigEndian.Uint32(b[i : i+4]))
		i += 4
		chunk, err = snappy.Decode(chunk, b[i:i+n])
		if err != nil {
			return nil, err
		}
		i += n
		decoded = append(decoded, chunk...)
	}
	return decoded, nil
}

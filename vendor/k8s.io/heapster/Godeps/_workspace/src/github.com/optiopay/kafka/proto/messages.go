package proto

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"
	"time"

	"github.com/golang/snappy"
)

/*

Kafka wire protocol implemented as described in
https://cwiki.apache.org/confluence/display/KAFKA/A+Guide+To+The+Kafka+Protocol#AGuideToTheKafkaProtocol-Messagesets

*/

const (
	ProduceReqKind          = 0
	FetchReqKind            = 1
	OffsetReqKind           = 2
	MetadataReqKind         = 3
	OffsetCommitReqKind     = 8
	OffsetFetchReqKind      = 9
	ConsumerMetadataReqKind = 10

	// receive the latest offset (i.e. the offset of the next coming message)
	OffsetReqTimeLatest = -1

	// receive the earliest available offset. Note that because offsets are
	// pulled in descending order, asking for the earliest offset will always
	// return you a single element.
	OffsetReqTimeEarliest = -2

	// Server will not send any response.
	RequiredAcksNone = 0

	// Server will block until the message is committed by all in sync replicas
	// before sending a response.
	RequiredAcksAll = -1

	// Server will wait the data is written to the local log before sending a
	// response.
	RequiredAcksLocal = 1
)

type Compression int8

const (
	CompressionNone   Compression = 0
	CompressionGzip   Compression = 1
	CompressionSnappy Compression = 2
)

// ReadReq returns request kind ID and byte representation of the whole message
// in wire protocol format.
func ReadReq(r io.Reader) (requestKind int16, b []byte, err error) {
	dec := NewDecoder(r)
	msgSize := dec.DecodeInt32()
	requestKind = dec.DecodeInt16()
	if err := dec.Err(); err != nil {
		return 0, nil, err
	}
	// size of the message + size of the message itself
	b = make([]byte, msgSize+4)
	binary.BigEndian.PutUint32(b, uint32(msgSize))
	binary.BigEndian.PutUint16(b[4:], uint16(requestKind))
	if _, err := io.ReadFull(r, b[6:]); err != nil {
		return 0, nil, err
	}
	return requestKind, b, err
}

// ReadResp returns message correlation ID and byte representation of the whole
// message in wire protocol that is returned when reading from given stream,
// including 4 bytes of message size itself.
// Byte representation returned by ReadResp can be parsed by all response
// reeaders to transform it into specialized response structure.
func ReadResp(r io.Reader) (correlationID int32, b []byte, err error) {
	dec := NewDecoder(r)
	msgSize := dec.DecodeInt32()
	correlationID = dec.DecodeInt32()
	if err := dec.Err(); err != nil {
		return 0, nil, err
	}
	// size of the message + size of the message itself
	b = make([]byte, msgSize+4)
	binary.BigEndian.PutUint32(b, uint32(msgSize))
	binary.BigEndian.PutUint32(b[4:], uint32(correlationID))
	_, err = io.ReadFull(r, b[8:])
	return correlationID, b, err
}

// Message represents single entity of message set.
type Message struct {
	Key       []byte
	Value     []byte
	Offset    int64  // set when fetching and after successful producing
	Crc       uint32 // set when fetching, ignored when producing
	Topic     string // set when fetching, ignored when producing
	Partition int32  // set when fetching, ignored when producing
	TipOffset int64  // set when fetching, ignored when processing
}

// ComputeCrc returns crc32 hash for given message content.
func ComputeCrc(m *Message, compression Compression) uint32 {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	enc.EncodeInt8(0) // magic byte is always 0
	enc.EncodeInt8(int8(compression))
	enc.EncodeBytes(m.Key)
	enc.EncodeBytes(m.Value)
	return crc32.ChecksumIEEE(buf.Bytes())
}

// writeMessageSet writes a Message Set into w.
// It returns the number of bytes written and any error.
func writeMessageSet(w io.Writer, messages []*Message, compression Compression) (int, error) {
	if len(messages) == 0 {
		return 0, nil
	}
	// NOTE(caleb): it doesn't appear to be documented, but I observed that the
	// Java client sets the offset of the synthesized message set for a group of
	// compressed messages to be the offset of the last message in the set.
	compressOffset := messages[len(messages)-1].Offset
	switch compression {
	case CompressionGzip:
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		if _, err := writeMessageSet(gz, messages, CompressionNone); err != nil {
			return 0, err
		}
		if err := gz.Close(); err != nil {
			return 0, err
		}
		messages = []*Message{
			{
				Value:  buf.Bytes(),
				Offset: compressOffset,
			},
		}
	case CompressionSnappy:
		var buf bytes.Buffer
		if _, err := writeMessageSet(&buf, messages, CompressionNone); err != nil {
			return 0, err
		}
		messages = []*Message{
			{
				Value:  snappy.Encode(nil, buf.Bytes()),
				Offset: compressOffset,
			},
		}
	}

	totalSize := 0
	b := newSliceWriter(0)
	for _, message := range messages {
		bsize := 26 + len(message.Key) + len(message.Value)
		b.Reset(bsize)

		enc := NewEncoder(b)
		enc.EncodeInt64(message.Offset)
		msize := int32(14 + len(message.Key) + len(message.Value))
		enc.EncodeInt32(msize)
		enc.EncodeUint32(0) // crc32 placeholder
		enc.EncodeInt8(0)   // magic byte
		enc.EncodeInt8(int8(compression))
		enc.EncodeBytes(message.Key)
		enc.EncodeBytes(message.Value)

		if err := enc.Err(); err != nil {
			return totalSize, err
		}

		const hsize = 8 + 4 + 4 // offset + message size + crc32
		const crcoff = 8 + 4    // offset + message size
		binary.BigEndian.PutUint32(b.buf[crcoff:crcoff+4], crc32.ChecksumIEEE(b.buf[hsize:bsize]))

		if n, err := w.Write(b.Slice()); err != nil {
			return totalSize, err
		} else {
			totalSize += n
		}

	}
	return totalSize, nil
}

type slicewriter struct {
	buf  []byte
	pos  int
	size int
}

func newSliceWriter(bufsize int) *slicewriter {
	return &slicewriter{
		buf: make([]byte, bufsize),
		pos: 0,
	}
}

func (w *slicewriter) Write(p []byte) (int, error) {
	if len(w.buf) < w.pos+len(p) {
		return 0, errors.New("buffer too small")
	}
	copy(w.buf[w.pos:], p)
	w.pos += len(p)
	return len(p), nil
}

func (w *slicewriter) Reset(size int) {
	if size > len(w.buf) {
		w.buf = make([]byte, size+1000) // allocate a bit more than required
	}
	w.size = size
	w.pos = 0
}

func (w *slicewriter) Slice() []byte {
	return w.buf[:w.pos]
}

// readMessageSet reads and return messages from the stream.
// The size is known before a message set is decoded.
// Because kafka is sending message set directly from the drive, it might cut
// off part of the last message. This also means that the last message can be
// shorter than the header is saying. In such case just ignore the last
// malformed message from the set and returned earlier data.
func readMessageSet(r io.Reader, size int32) ([]*Message, error) {
	rd := io.LimitReader(r, int64(size))
	dec := NewDecoder(rd)
	set := make([]*Message, 0, 256)

	var buf []byte
	for {
		offset := dec.DecodeInt64()
		if err := dec.Err(); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return set, nil
			}
			return nil, err
		}
		// single message size
		size := dec.DecodeInt32()
		if err := dec.Err(); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return set, nil
			}
			return nil, err
		}

		// read message to buffer to compute its content crc
		if int(size) > len(buf) {
			// allocate a bit more than needed
			buf = make([]byte, size+10240)
		}
		msgbuf := buf[:size]

		if _, err := io.ReadFull(rd, msgbuf); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return set, nil
			}
			return nil, err
		}
		msgdec := NewDecoder(bytes.NewBuffer(msgbuf))

		msg := &Message{
			Offset: offset,
			Crc:    msgdec.DecodeUint32(),
		}

		if msg.Crc != crc32.ChecksumIEEE(msgbuf[4:]) {
			// ignore this message and because we want to have constant
			// history, do not process anything more
			return set, nil
		}

		// magic byte
		_ = msgdec.DecodeInt8()

		attributes := msgdec.DecodeInt8()
		switch compression := Compression(attributes & 3); compression {
		case CompressionNone:
			msg.Key = msgdec.DecodeBytes()
			msg.Value = msgdec.DecodeBytes()
			if err := msgdec.Err(); err != nil {
				return nil, fmt.Errorf("cannot decode message: %s", err)
			}
			set = append(set, msg)
		case CompressionGzip, CompressionSnappy:
			_ = msgdec.DecodeBytes() // ignore key
			val := msgdec.DecodeBytes()
			if err := msgdec.Err(); err != nil {
				return nil, fmt.Errorf("cannot decode message: %s", err)
			}
			var decoded []byte
			switch compression {
			case CompressionGzip:
				cr, err := gzip.NewReader(bytes.NewReader(val))
				if err != nil {
					return nil, fmt.Errorf("error decoding gzip message: %s", err)
				}
				decoded, err = ioutil.ReadAll(cr)
				if err != nil {
					return nil, fmt.Errorf("error decoding gzip message: %s", err)
				}
				_ = cr.Close()
			case CompressionSnappy:
				var err error
				decoded, err = snappyDecode(val)
				if err != nil {
					return nil, fmt.Errorf("error decoding snappy message: %s", err)
				}
			}
			msgs, err := readMessageSet(bytes.NewReader(decoded), int32(len(decoded)))
			if err != nil {
				return nil, err
			}
			set = append(set, msgs...)
		default:
			return nil, fmt.Errorf("cannot handle compression method: %d", compression)
		}
	}
}

type MetadataReq struct {
	CorrelationID int32
	ClientID      string
	Topics        []string
}

func ReadMetadataReq(r io.Reader) (*MetadataReq, error) {
	var req MetadataReq
	dec := NewDecoder(r)

	// total message size
	_ = dec.DecodeInt32()
	// api key + api version
	_ = dec.DecodeInt32()
	req.CorrelationID = dec.DecodeInt32()
	req.ClientID = dec.DecodeString()
	req.Topics = make([]string, dec.DecodeArrayLen())
	for i := range req.Topics {
		req.Topics[i] = dec.DecodeString()
	}

	if dec.Err() != nil {
		return nil, dec.Err()
	}
	return &req, nil
}

func (r *MetadataReq) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	// message size - for now just placeholder
	enc.Encode(int32(0))
	enc.Encode(int16(MetadataReqKind))
	enc.Encode(int16(0))
	enc.Encode(r.CorrelationID)
	enc.Encode(r.ClientID)

	enc.EncodeArrayLen(len(r.Topics))
	for _, name := range r.Topics {
		enc.Encode(name)
	}

	if enc.Err() != nil {
		return nil, enc.Err()
	}

	// update the message size information
	b := buf.Bytes()
	binary.BigEndian.PutUint32(b, uint32(len(b)-4))

	return b, nil
}

func (r *MetadataReq) WriteTo(w io.Writer) (int64, error) {
	b, err := r.Bytes()
	if err != nil {
		return 0, err
	}
	n, err := w.Write(b)
	return int64(n), err
}

type MetadataResp struct {
	CorrelationID int32
	Brokers       []MetadataRespBroker
	Topics        []MetadataRespTopic
}

type MetadataRespBroker struct {
	NodeID int32
	Host   string
	Port   int32
}

type MetadataRespTopic struct {
	Name       string
	Err        error
	Partitions []MetadataRespPartition
}

type MetadataRespPartition struct {
	ID       int32
	Err      error
	Leader   int32
	Replicas []int32
	Isrs     []int32
}

func (r *MetadataResp) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	// message size - for now just placeholder
	enc.Encode(int32(0))
	enc.Encode(r.CorrelationID)
	enc.EncodeArrayLen(len(r.Brokers))
	for _, broker := range r.Brokers {
		enc.Encode(broker.NodeID)
		enc.Encode(broker.Host)
		enc.Encode(broker.Port)
	}
	enc.EncodeArrayLen(len(r.Topics))
	for _, topic := range r.Topics {
		enc.EncodeError(topic.Err)
		enc.Encode(topic.Name)
		enc.EncodeArrayLen(len(topic.Partitions))
		for _, part := range topic.Partitions {
			enc.EncodeError(part.Err)
			enc.Encode(part.ID)
			enc.Encode(part.Leader)
			enc.Encode(part.Replicas)
			enc.Encode(part.Isrs)
		}
	}

	if enc.Err() != nil {
		return nil, enc.Err()
	}

	// update the message size information
	b := buf.Bytes()
	binary.BigEndian.PutUint32(b, uint32(len(b)-4))

	return b, nil
}

func ReadMetadataResp(r io.Reader) (*MetadataResp, error) {
	var resp MetadataResp
	dec := NewDecoder(r)

	// total message size
	_ = dec.DecodeInt32()
	resp.CorrelationID = dec.DecodeInt32()

	resp.Brokers = make([]MetadataRespBroker, dec.DecodeArrayLen())
	for i := range resp.Brokers {
		var b = &resp.Brokers[i]
		b.NodeID = dec.DecodeInt32()
		b.Host = dec.DecodeString()
		b.Port = dec.DecodeInt32()
	}

	resp.Topics = make([]MetadataRespTopic, dec.DecodeArrayLen())
	for ti := range resp.Topics {
		var t = &resp.Topics[ti]
		t.Err = errFromNo(dec.DecodeInt16())
		t.Name = dec.DecodeString()
		t.Partitions = make([]MetadataRespPartition, dec.DecodeArrayLen())
		for pi := range t.Partitions {
			var p = &t.Partitions[pi]
			p.Err = errFromNo(dec.DecodeInt16())
			p.ID = dec.DecodeInt32()
			p.Leader = dec.DecodeInt32()

			p.Replicas = make([]int32, dec.DecodeArrayLen())
			for ri := range p.Replicas {
				p.Replicas[ri] = dec.DecodeInt32()
			}

			p.Isrs = make([]int32, dec.DecodeArrayLen())
			for ii := range p.Isrs {
				p.Isrs[ii] = dec.DecodeInt32()
			}
		}
	}

	if dec.Err() != nil {
		return nil, dec.Err()
	}
	return &resp, nil
}

type FetchReq struct {
	CorrelationID int32
	ClientID      string
	MaxWaitTime   time.Duration
	MinBytes      int32

	Topics []FetchReqTopic
}

type FetchReqTopic struct {
	Name       string
	Partitions []FetchReqPartition
}

type FetchReqPartition struct {
	ID          int32
	FetchOffset int64
	MaxBytes    int32
}

func ReadFetchReq(r io.Reader) (*FetchReq, error) {
	var req FetchReq
	dec := NewDecoder(r)

	// total message size
	_ = dec.DecodeInt32()
	// api key + api version
	_ = dec.DecodeInt32()
	req.CorrelationID = dec.DecodeInt32()
	req.ClientID = dec.DecodeString()
	// replica id
	_ = dec.DecodeInt32()
	req.MaxWaitTime = time.Duration(dec.DecodeInt32()) * time.Millisecond
	req.MinBytes = dec.DecodeInt32()
	req.Topics = make([]FetchReqTopic, dec.DecodeArrayLen())
	for ti := range req.Topics {
		var topic = &req.Topics[ti]
		topic.Name = dec.DecodeString()
		topic.Partitions = make([]FetchReqPartition, dec.DecodeArrayLen())
		for pi := range topic.Partitions {
			var part = &topic.Partitions[pi]
			part.ID = dec.DecodeInt32()
			part.FetchOffset = dec.DecodeInt64()
			part.MaxBytes = dec.DecodeInt32()
		}
	}

	if dec.Err() != nil {
		return nil, dec.Err()
	}
	return &req, nil
}

func (r *FetchReq) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	// message size - for now just placeholder
	enc.Encode(int32(0))
	enc.Encode(int16(FetchReqKind))
	enc.Encode(int16(0))
	enc.Encode(r.CorrelationID)
	enc.Encode(r.ClientID)

	// replica id
	enc.Encode(int32(-1))
	enc.Encode(int32(r.MaxWaitTime / time.Millisecond))
	enc.Encode(r.MinBytes)

	enc.EncodeArrayLen(len(r.Topics))
	for _, topic := range r.Topics {
		enc.Encode(topic.Name)
		enc.EncodeArrayLen(len(topic.Partitions))
		for _, part := range topic.Partitions {
			enc.Encode(part.ID)
			enc.Encode(part.FetchOffset)
			enc.Encode(part.MaxBytes)
		}
	}

	if enc.Err() != nil {
		return nil, enc.Err()
	}

	// update the message size information
	b := buf.Bytes()
	binary.BigEndian.PutUint32(b, uint32(len(b)-4))

	return b, nil
}

func (r *FetchReq) WriteTo(w io.Writer) (int64, error) {
	b, err := r.Bytes()
	if err != nil {
		return 0, err
	}
	n, err := w.Write(b)
	return int64(n), err
}

type FetchResp struct {
	CorrelationID int32
	Topics        []FetchRespTopic
}

type FetchRespTopic struct {
	Name       string
	Partitions []FetchRespPartition
}

type FetchRespPartition struct {
	ID        int32
	Err       error
	TipOffset int64
	Messages  []*Message
}

func (r *FetchResp) Bytes() ([]byte, error) {
	var buf buffer
	enc := NewEncoder(&buf)

	enc.Encode(int32(0)) // placeholder
	enc.Encode(r.CorrelationID)
	enc.EncodeArrayLen(len(r.Topics))
	for _, topic := range r.Topics {
		enc.Encode(topic.Name)
		enc.EncodeArrayLen(len(topic.Partitions))
		for _, part := range topic.Partitions {
			enc.Encode(part.ID)
			enc.EncodeError(part.Err)
			enc.Encode(part.TipOffset)
			i := len(buf)
			enc.Encode(int32(0)) // placeholder
			// NOTE(caleb): writing compressed fetch response isn't implemented
			// for now, since that's not needed for clients.
			n, err := writeMessageSet(&buf, part.Messages, CompressionNone)
			if err != nil {
				return nil, err
			}
			binary.BigEndian.PutUint32(buf[i:i+4], uint32(n))
		}
	}

	if enc.Err() != nil {
		return nil, enc.Err()
	}

	binary.BigEndian.PutUint32(buf[:4], uint32(len(buf)-4))
	return []byte(buf), nil
}

func ReadFetchResp(r io.Reader) (*FetchResp, error) {
	var err error
	var resp FetchResp

	dec := NewDecoder(r)

	// total message size
	_ = dec.DecodeInt32()
	resp.CorrelationID = dec.DecodeInt32()

	resp.Topics = make([]FetchRespTopic, dec.DecodeArrayLen())
	for ti := range resp.Topics {
		var topic = &resp.Topics[ti]
		topic.Name = dec.DecodeString()
		topic.Partitions = make([]FetchRespPartition, dec.DecodeArrayLen())
		for pi := range topic.Partitions {
			var part = &topic.Partitions[pi]
			part.ID = dec.DecodeInt32()
			part.Err = errFromNo(dec.DecodeInt16())
			part.TipOffset = dec.DecodeInt64()
			if dec.Err() != nil {
				return nil, dec.Err()
			}
			msgSetSize := dec.DecodeInt32()
			if dec.Err() != nil {
				return nil, dec.Err()
			}
			if part.Messages, err = readMessageSet(r, msgSetSize); err != nil {
				return nil, err
			}
			for _, msg := range part.Messages {
				msg.Topic = topic.Name
				msg.Partition = part.ID
				msg.TipOffset = part.TipOffset
			}
		}
	}

	if dec.Err() != nil {
		return nil, dec.Err()
	}
	return &resp, nil
}

type ConsumerMetadataReq struct {
	CorrelationID int32
	ClientID      string
	ConsumerGroup string
}

func ReadConsumerMetadataReq(r io.Reader) (*ConsumerMetadataReq, error) {
	var req ConsumerMetadataReq
	dec := NewDecoder(r)

	// total message size
	_ = dec.DecodeInt32()
	// api key + api version
	_ = dec.DecodeInt32()
	req.CorrelationID = dec.DecodeInt32()
	req.ClientID = dec.DecodeString()
	req.ConsumerGroup = dec.DecodeString()

	if dec.Err() != nil {
		return nil, dec.Err()
	}
	return &req, nil
}

func (r *ConsumerMetadataReq) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	// message size - for now just placeholder
	enc.Encode(int32(0))
	enc.Encode(int16(ConsumerMetadataReqKind))
	enc.Encode(int16(0))
	enc.Encode(r.CorrelationID)
	enc.Encode(r.ClientID)

	enc.Encode(r.ConsumerGroup)

	if enc.Err() != nil {
		return nil, enc.Err()
	}

	// update the message size information
	b := buf.Bytes()
	binary.BigEndian.PutUint32(b, uint32(len(b)-4))

	return b, nil
}

func (r *ConsumerMetadataReq) WriteTo(w io.Writer) (int64, error) {
	b, err := r.Bytes()
	if err != nil {
		return 0, err
	}
	n, err := w.Write(b)
	return int64(n), err
}

type ConsumerMetadataResp struct {
	CorrelationID   int32
	Err             error
	CoordinatorID   int32
	CoordinatorHost string
	CoordinatorPort int32
}

func ReadConsumerMetadataResp(r io.Reader) (*ConsumerMetadataResp, error) {
	var resp ConsumerMetadataResp
	dec := NewDecoder(r)

	// total message size
	_ = dec.DecodeInt32()
	resp.CorrelationID = dec.DecodeInt32()
	resp.Err = errFromNo(dec.DecodeInt16())
	resp.CoordinatorID = dec.DecodeInt32()
	resp.CoordinatorHost = dec.DecodeString()
	resp.CoordinatorPort = dec.DecodeInt32()

	if err := dec.Err(); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *ConsumerMetadataResp) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	// message size - for now just placeholder
	enc.Encode(int32(0))
	enc.Encode(r.CorrelationID)
	enc.EncodeError(r.Err)
	enc.Encode(r.CoordinatorID)
	enc.Encode(r.CoordinatorHost)
	enc.Encode(r.CoordinatorPort)

	if enc.Err() != nil {
		return nil, enc.Err()
	}

	// update the message size information
	b := buf.Bytes()
	binary.BigEndian.PutUint32(b, uint32(len(b)-4))

	return b, nil
}

type OffsetCommitReq struct {
	CorrelationID int32
	ClientID      string
	ConsumerGroup string
	Topics        []OffsetCommitReqTopic
}

type OffsetCommitReqTopic struct {
	Name       string
	Partitions []OffsetCommitReqPartition
}

type OffsetCommitReqPartition struct {
	ID        int32
	Offset    int64
	TimeStamp time.Time
	Metadata  string
}

func ReadOffsetCommitReq(r io.Reader) (*OffsetCommitReq, error) {
	var req OffsetCommitReq
	dec := NewDecoder(r)

	// total message size
	_ = dec.DecodeInt32()
	// api key + api version
	_ = dec.DecodeInt32()
	req.CorrelationID = dec.DecodeInt32()
	req.ClientID = dec.DecodeString()
	req.ConsumerGroup = dec.DecodeString()
	req.Topics = make([]OffsetCommitReqTopic, dec.DecodeArrayLen())
	for ti := range req.Topics {
		var topic = &req.Topics[ti]
		topic.Name = dec.DecodeString()
		topic.Partitions = make([]OffsetCommitReqPartition, dec.DecodeArrayLen())
		for pi := range topic.Partitions {
			var part = &topic.Partitions[pi]
			part.ID = dec.DecodeInt32()
			part.Offset = dec.DecodeInt64()
			part.TimeStamp = time.Unix(0, dec.DecodeInt64()*int64(time.Millisecond))
			part.Metadata = dec.DecodeString()
		}
	}

	if dec.Err() != nil {
		return nil, dec.Err()
	}
	return &req, nil
}

func (r *OffsetCommitReq) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	// message size - for now just placeholder
	enc.Encode(int32(0))
	enc.Encode(int16(OffsetCommitReqKind))
	enc.Encode(int16(0))
	enc.Encode(r.CorrelationID)
	enc.Encode(r.ClientID)

	enc.Encode(r.ConsumerGroup)

	enc.EncodeArrayLen(len(r.Topics))
	for _, topic := range r.Topics {
		enc.Encode(topic.Name)
		enc.EncodeArrayLen(len(topic.Partitions))
		for _, part := range topic.Partitions {
			enc.Encode(part.ID)
			enc.Encode(part.Offset)
			// TODO(husio) is this really in milliseconds?
			enc.Encode(part.TimeStamp.UnixNano() / int64(time.Millisecond))
			enc.Encode(part.Metadata)
		}
	}

	if enc.Err() != nil {
		return nil, enc.Err()
	}

	// update the message size information
	b := buf.Bytes()
	binary.BigEndian.PutUint32(b, uint32(len(b)-4))

	return b, nil
}

func (r *OffsetCommitReq) WriteTo(w io.Writer) (int64, error) {
	b, err := r.Bytes()
	if err != nil {
		return 0, err
	}
	n, err := w.Write(b)
	return int64(n), err
}

type OffsetCommitResp struct {
	CorrelationID int32
	Topics        []OffsetCommitRespTopic
}

type OffsetCommitRespTopic struct {
	Name       string
	Partitions []OffsetCommitRespPartition
}

type OffsetCommitRespPartition struct {
	ID  int32
	Err error
}

func ReadOffsetCommitResp(r io.Reader) (*OffsetCommitResp, error) {
	var resp OffsetCommitResp
	dec := NewDecoder(r)

	// total message size
	_ = dec.DecodeInt32()
	resp.CorrelationID = dec.DecodeInt32()
	resp.Topics = make([]OffsetCommitRespTopic, dec.DecodeArrayLen())
	for ti := range resp.Topics {
		var t = &resp.Topics[ti]
		t.Name = dec.DecodeString()
		t.Partitions = make([]OffsetCommitRespPartition, dec.DecodeArrayLen())
		for pi := range t.Partitions {
			var p = &t.Partitions[pi]
			p.ID = dec.DecodeInt32()
			p.Err = errFromNo(dec.DecodeInt16())
		}
	}

	if err := dec.Err(); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *OffsetCommitResp) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	// message size - for now just placeholder
	enc.Encode(int32(0))
	enc.Encode(r.CorrelationID)
	enc.EncodeArrayLen(len(r.Topics))
	for _, t := range r.Topics {
		enc.Encode(t.Name)
		enc.EncodeArrayLen(len(t.Partitions))
		for _, p := range t.Partitions {
			enc.Encode(p.ID)
			enc.EncodeError(p.Err)
		}
	}

	if enc.Err() != nil {
		return nil, enc.Err()
	}

	// update the message size information
	b := buf.Bytes()
	binary.BigEndian.PutUint32(b, uint32(len(b)-4))

	return b, nil

}

type OffsetFetchReq struct {
	CorrelationID int32
	ClientID      string
	ConsumerGroup string
	Topics        []OffsetFetchReqTopic
}

type OffsetFetchReqTopic struct {
	Name       string
	Partitions []int32
}

func ReadOffsetFetchReq(r io.Reader) (*OffsetFetchReq, error) {
	var req OffsetFetchReq
	dec := NewDecoder(r)

	// total message size
	_ = dec.DecodeInt32()
	// api key + api version
	_ = dec.DecodeInt32()
	req.CorrelationID = dec.DecodeInt32()
	req.ClientID = dec.DecodeString()
	req.ConsumerGroup = dec.DecodeString()
	req.Topics = make([]OffsetFetchReqTopic, dec.DecodeArrayLen())
	for ti := range req.Topics {
		var topic = &req.Topics[ti]
		topic.Name = dec.DecodeString()
		topic.Partitions = make([]int32, dec.DecodeArrayLen())
		for pi := range topic.Partitions {
			topic.Partitions[pi] = dec.DecodeInt32()
		}
	}

	if dec.Err() != nil {
		return nil, dec.Err()
	}
	return &req, nil
}

func (r *OffsetFetchReq) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	// message size - for now just placeholder
	enc.Encode(int32(0))
	enc.Encode(int16(OffsetFetchReqKind))
	enc.Encode(int16(0))
	enc.Encode(r.CorrelationID)
	enc.Encode(r.ClientID)

	enc.Encode(r.ConsumerGroup)
	enc.EncodeArrayLen(len(r.Topics))
	for _, t := range r.Topics {
		enc.Encode(t.Name)
		enc.EncodeArrayLen(len(t.Partitions))
		for _, p := range t.Partitions {
			enc.Encode(p)
		}
	}

	if enc.Err() != nil {
		return nil, enc.Err()
	}

	// update the message size information
	b := buf.Bytes()
	binary.BigEndian.PutUint32(b, uint32(len(b)-4))

	return b, nil
}

func (r *OffsetFetchReq) WriteTo(w io.Writer) (int64, error) {
	b, err := r.Bytes()
	if err != nil {
		return 0, err
	}
	n, err := w.Write(b)
	return int64(n), err
}

type OffsetFetchResp struct {
	CorrelationID int32
	Topics        []OffsetFetchRespTopic
}

type OffsetFetchRespTopic struct {
	Name       string
	Partitions []OffsetFetchRespPartition
}

type OffsetFetchRespPartition struct {
	ID       int32
	Offset   int64
	Metadata string
	Err      error
}

func ReadOffsetFetchResp(r io.Reader) (*OffsetFetchResp, error) {
	var resp OffsetFetchResp
	dec := NewDecoder(r)

	// total message size
	_ = dec.DecodeInt32()
	resp.CorrelationID = dec.DecodeInt32()
	resp.Topics = make([]OffsetFetchRespTopic, dec.DecodeArrayLen())
	for ti := range resp.Topics {
		var t = &resp.Topics[ti]
		t.Name = dec.DecodeString()
		t.Partitions = make([]OffsetFetchRespPartition, dec.DecodeArrayLen())
		for pi := range t.Partitions {
			var p = &t.Partitions[pi]
			p.ID = dec.DecodeInt32()
			p.Offset = dec.DecodeInt64()
			p.Metadata = dec.DecodeString()
			p.Err = errFromNo(dec.DecodeInt16())
		}
	}

	if err := dec.Err(); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *OffsetFetchResp) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	// message size - for now just placeholder
	enc.Encode(int32(0))
	enc.Encode(r.CorrelationID)
	enc.EncodeArrayLen(len(r.Topics))
	for _, topic := range r.Topics {
		enc.Encode(topic.Name)
		enc.EncodeArrayLen(len(topic.Partitions))
		for _, part := range topic.Partitions {
			enc.Encode(part.ID)
			enc.Encode(part.Offset)
			enc.Encode(part.Metadata)
			enc.EncodeError(part.Err)
		}
	}

	if enc.Err() != nil {
		return nil, enc.Err()
	}

	// update the message size information
	b := buf.Bytes()
	binary.BigEndian.PutUint32(b, uint32(len(b)-4))

	return b, nil
}

type ProduceReq struct {
	CorrelationID int32
	ClientID      string
	Compression   Compression // only used when sending ProduceReqs
	RequiredAcks  int16
	Timeout       time.Duration
	Topics        []ProduceReqTopic
}

type ProduceReqTopic struct {
	Name       string
	Partitions []ProduceReqPartition
}

type ProduceReqPartition struct {
	ID       int32
	Messages []*Message
}

func ReadProduceReq(r io.Reader) (*ProduceReq, error) {
	var req ProduceReq
	dec := NewDecoder(r)

	// total message size
	_ = dec.DecodeInt32()
	// api key + api version
	_ = dec.DecodeInt32()
	req.CorrelationID = dec.DecodeInt32()
	req.ClientID = dec.DecodeString()
	req.RequiredAcks = dec.DecodeInt16()
	req.Timeout = time.Duration(dec.DecodeInt32()) * time.Millisecond
	req.Topics = make([]ProduceReqTopic, dec.DecodeArrayLen())
	for ti := range req.Topics {
		var topic = &req.Topics[ti]
		topic.Name = dec.DecodeString()
		topic.Partitions = make([]ProduceReqPartition, dec.DecodeArrayLen())
		for pi := range topic.Partitions {
			var part = &topic.Partitions[pi]
			part.ID = dec.DecodeInt32()
			if dec.Err() != nil {
				return nil, dec.Err()
			}
			msgSetSize := dec.DecodeInt32()
			if dec.Err() != nil {
				return nil, dec.Err()
			}
			var err error
			if part.Messages, err = readMessageSet(r, msgSetSize); err != nil {
				return nil, err
			}
		}
	}

	if dec.Err() != nil {
		return nil, dec.Err()
	}
	return &req, nil
}

func (r *ProduceReq) Bytes() ([]byte, error) {
	var buf buffer
	enc := NewEncoder(&buf)

	enc.EncodeInt32(0) // placeholder
	enc.EncodeInt16(ProduceReqKind)
	enc.EncodeInt16(0)
	enc.EncodeInt32(r.CorrelationID)
	enc.EncodeString(r.ClientID)

	enc.EncodeInt16(r.RequiredAcks)
	enc.EncodeInt32(int32(r.Timeout / time.Millisecond))
	enc.EncodeArrayLen(len(r.Topics))
	for _, t := range r.Topics {
		enc.EncodeString(t.Name)
		enc.EncodeArrayLen(len(t.Partitions))
		for _, p := range t.Partitions {
			enc.EncodeInt32(p.ID)
			i := len(buf)
			enc.EncodeInt32(0) // placeholder
			n, err := writeMessageSet(&buf, p.Messages, r.Compression)
			if err != nil {
				return nil, err
			}
			binary.BigEndian.PutUint32(buf[i:i+4], uint32(n))
		}
	}

	if enc.Err() != nil {
		return nil, enc.Err()
	}

	binary.BigEndian.PutUint32(buf[0:4], uint32(len(buf)-4))
	return []byte(buf), nil
}

func (r *ProduceReq) WriteTo(w io.Writer) (int64, error) {
	b, err := r.Bytes()
	if err != nil {
		return 0, err
	}
	n, err := w.Write(b)
	return int64(n), err
}

type ProduceResp struct {
	CorrelationID int32
	Topics        []ProduceRespTopic
}

type ProduceRespTopic struct {
	Name       string
	Partitions []ProduceRespPartition
}

type ProduceRespPartition struct {
	ID     int32
	Err    error
	Offset int64
}

func (r *ProduceResp) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	// message size - for now just placeholder
	enc.Encode(int32(0))
	enc.Encode(r.CorrelationID)
	enc.EncodeArrayLen(len(r.Topics))
	for _, topic := range r.Topics {
		enc.Encode(topic.Name)
		enc.EncodeArrayLen(len(topic.Partitions))
		for _, part := range topic.Partitions {
			enc.Encode(part.ID)
			enc.EncodeError(part.Err)
			enc.Encode(part.Offset)
		}
	}

	if enc.Err() != nil {
		return nil, enc.Err()
	}

	// update the message size information
	b := buf.Bytes()
	binary.BigEndian.PutUint32(b, uint32(len(b)-4))

	return b, nil
}

func ReadProduceResp(r io.Reader) (*ProduceResp, error) {
	var resp ProduceResp
	dec := NewDecoder(r)

	// total message size
	_ = dec.DecodeInt32()
	resp.CorrelationID = dec.DecodeInt32()
	resp.Topics = make([]ProduceRespTopic, dec.DecodeArrayLen())
	for ti := range resp.Topics {
		var t = &resp.Topics[ti]
		t.Name = dec.DecodeString()
		t.Partitions = make([]ProduceRespPartition, dec.DecodeArrayLen())
		for pi := range t.Partitions {
			var p = &t.Partitions[pi]
			p.ID = dec.DecodeInt32()
			p.Err = errFromNo(dec.DecodeInt16())
			p.Offset = dec.DecodeInt64()
		}
	}

	if err := dec.Err(); err != nil {
		return nil, err
	}
	return &resp, nil
}

type OffsetReq struct {
	CorrelationID int32
	ClientID      string
	ReplicaID     int32
	Topics        []OffsetReqTopic
}

type OffsetReqTopic struct {
	Name       string
	Partitions []OffsetReqPartition
}

type OffsetReqPartition struct {
	ID         int32
	TimeMs     int64 // cannot be time.Time because of negative values
	MaxOffsets int32
}

func ReadOffsetReq(r io.Reader) (*OffsetReq, error) {
	var req OffsetReq
	dec := NewDecoder(r)

	// total message size
	_ = dec.DecodeInt32()
	// api key + api version
	_ = dec.DecodeInt32()
	req.CorrelationID = dec.DecodeInt32()
	req.ClientID = dec.DecodeString()
	req.ReplicaID = dec.DecodeInt32()
	req.Topics = make([]OffsetReqTopic, dec.DecodeArrayLen())
	for ti := range req.Topics {
		var topic = &req.Topics[ti]
		topic.Name = dec.DecodeString()
		topic.Partitions = make([]OffsetReqPartition, dec.DecodeArrayLen())
		for pi := range topic.Partitions {
			var part = &topic.Partitions[pi]
			part.ID = dec.DecodeInt32()
			part.TimeMs = dec.DecodeInt64()
			part.MaxOffsets = dec.DecodeInt32()
		}
	}

	if dec.Err() != nil {
		return nil, dec.Err()
	}
	return &req, nil
}

func (r *OffsetReq) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	// message size - for now just placeholder
	enc.Encode(int32(0))
	enc.Encode(int16(OffsetReqKind))
	enc.Encode(int16(0))
	enc.Encode(r.CorrelationID)
	enc.Encode(r.ClientID)

	enc.Encode(r.ReplicaID)
	enc.EncodeArrayLen(len(r.Topics))
	for _, topic := range r.Topics {
		enc.Encode(topic.Name)
		enc.EncodeArrayLen(len(topic.Partitions))
		for _, part := range topic.Partitions {
			enc.Encode(part.ID)
			enc.Encode(part.TimeMs)
			enc.Encode(part.MaxOffsets)
		}
	}

	if enc.Err() != nil {
		return nil, enc.Err()
	}

	// update the message size information
	b := buf.Bytes()
	binary.BigEndian.PutUint32(b, uint32(len(b)-4))

	return b, nil
}

func (r *OffsetReq) WriteTo(w io.Writer) (int64, error) {
	b, err := r.Bytes()
	if err != nil {
		return 0, err
	}
	n, err := w.Write(b)
	return int64(n), err
}

type OffsetResp struct {
	CorrelationID int32
	Topics        []OffsetRespTopic
}

type OffsetRespTopic struct {
	Name       string
	Partitions []OffsetRespPartition
}

type OffsetRespPartition struct {
	ID      int32
	Err     error
	Offsets []int64
}

func ReadOffsetResp(r io.Reader) (*OffsetResp, error) {
	var resp OffsetResp
	dec := NewDecoder(r)

	// total message size
	_ = dec.DecodeInt32()
	resp.CorrelationID = dec.DecodeInt32()
	resp.Topics = make([]OffsetRespTopic, dec.DecodeArrayLen())
	for ti := range resp.Topics {
		var t = &resp.Topics[ti]
		t.Name = dec.DecodeString()
		t.Partitions = make([]OffsetRespPartition, dec.DecodeArrayLen())
		for pi := range t.Partitions {
			var p = &t.Partitions[pi]
			p.ID = dec.DecodeInt32()
			p.Err = errFromNo(dec.DecodeInt16())
			p.Offsets = make([]int64, dec.DecodeArrayLen())
			for oi := range p.Offsets {
				p.Offsets[oi] = dec.DecodeInt64()
			}
		}
	}

	if err := dec.Err(); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (r *OffsetResp) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	// message size - for now just placeholder
	enc.Encode(int32(0))
	enc.Encode(r.CorrelationID)
	enc.EncodeArrayLen(len(r.Topics))
	for _, topic := range r.Topics {
		enc.Encode(topic.Name)
		enc.EncodeArrayLen(len(topic.Partitions))
		for _, part := range topic.Partitions {
			enc.Encode(part.ID)
			enc.EncodeError(part.Err)
			enc.EncodeArrayLen(len(part.Offsets))
			for _, off := range part.Offsets {
				enc.Encode(off)
			}
		}
	}

	if enc.Err() != nil {
		return nil, enc.Err()
	}

	// update the message size information
	b := buf.Bytes()
	binary.BigEndian.PutUint32(b, uint32(len(b)-4))

	return b, nil
}

type buffer []byte

func (b *buffer) Write(p []byte) (int, error) {
	*b = append(*b, p...)
	return len(p), nil
}

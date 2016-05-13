package kafka

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"sync"
	"syscall"
	"time"

	"github.com/optiopay/kafka/proto"
)

const (
	// StartOffsetNewest configures the consumer to fetch messages produced
	// after creating the consumer.
	StartOffsetNewest = -1

	// StartOffsetOldest configures the consumer to fetch starting from the
	// oldest message available.
	StartOffsetOldest = -2
)

var (
	// Returned by consumers on Fetch when the retry limit is set and exceeded.
	ErrNoData = errors.New("no data")

	// Make sure interfaces are implemented
	_ Client            = &Broker{}
	_ Consumer          = &consumer{}
	_ Producer          = &producer{}
	_ OffsetCoordinator = &offsetCoordinator{}
)

// Client is the interface implemented by Broker.
type Client interface {
	Producer(conf ProducerConf) Producer
	Consumer(conf ConsumerConf) (Consumer, error)
	OffsetCoordinator(conf OffsetCoordinatorConf) (OffsetCoordinator, error)
	OffsetEarliest(topic string, partition int32) (offset int64, err error)
	OffsetLatest(topic string, partition int32) (offset int64, err error)
	Close()
}

// Consumer is the interface that wraps the Consume method.
//
// Consume reads a message from a consumer, returning an error when
// encountered.
type Consumer interface {
	Consume() (*proto.Message, error)
}

// Producer is the interface that wraps the Produce method.
//
// Produce writes the messages to the given topic and partition.
// It returns the offset of the first message and any error encountered.
// The offset of each message is also updated accordingly.
type Producer interface {
	Produce(topic string, partition int32, messages ...*proto.Message) (offset int64, err error)
}

// OffsetCoordinator is the interface which wraps the Commit and Offset methods.
type OffsetCoordinator interface {
	Commit(topic string, partition int32, offset int64) error
	Offset(topic string, partition int32) (offset int64, metadata string, err error)
}

type topicPartition struct {
	topic     string
	partition int32
}

func (tp topicPartition) String() string {
	return fmt.Sprintf("%s:%d", tp.topic, tp.partition)
}

type clusterMetadata struct {
	created    time.Time
	nodes      map[int32]string         // node ID to address
	endpoints  map[topicPartition]int32 // partition to leader node ID
	partitions map[string]int32         // topic to number of partitions
}

type BrokerConf struct {
	// Kafka client ID.
	ClientID string

	// LeaderRetryLimit limits the number of connection attempts to a single
	// node before failing. Use LeaderRetryWait to control the wait time
	// between retries.
	//
	// Defaults to 10.
	LeaderRetryLimit int

	// LeaderRetryWait sets a limit to the waiting time when trying to connect
	// to a single node after failure.
	//
	// Defaults to 500ms.
	//
	// Timeout on a connection is controlled by the DialTimeout setting.
	LeaderRetryWait time.Duration

	// AllowTopicCreation enables a last-ditch "send produce request" which
	// happens if we do not know about a topic. This enables topic creation
	// if your Kafka cluster is configured to allow it.
	//
	// Defaults to False.
	AllowTopicCreation bool

	// Any new connection dial timeout.
	//
	// Default is 10 seconds.
	DialTimeout time.Duration

	// DialRetryLimit limits the number of connection attempts to every node in
	// cluster before failing. Use DialRetryWait to control the wait time
	// between retries.
	//
	// Defaults to 10.
	DialRetryLimit int

	// DialRetryWait sets a limit to the waiting time when trying to establish
	// broker connection to single node to fetch cluster metadata.
	//
	// Defaults to 500ms.
	DialRetryWait time.Duration

	// DEPRECATED 2015-07-10 - use Logger instead
	//
	// TODO(husio) remove
	//
	// Logger used by the broker.
	Log interface {
		Print(...interface{})
		Printf(string, ...interface{})
	}

	// Logger is general logging interface that can be provided by popular
	// logging frameworks. Used to notify and as replacement for stdlib `log`
	// package.
	Logger Logger
}

func NewBrokerConf(clientID string) BrokerConf {
	return BrokerConf{
		ClientID:           clientID,
		DialTimeout:        10 * time.Second,
		DialRetryLimit:     10,
		DialRetryWait:      500 * time.Millisecond,
		AllowTopicCreation: false,
		LeaderRetryLimit:   10,
		LeaderRetryWait:    500 * time.Millisecond,
		Logger:             &nullLogger{},
	}
}

// Broker is an abstract connection to kafka cluster, managing connections to
// all kafka nodes.
type Broker struct {
	conf BrokerConf

	mu       sync.Mutex
	metadata clusterMetadata
	conns    map[int32]*connection
}

// Dial connects to any node from a given list of kafka addresses and after
// successful metadata fetch, returns broker.
//
// The returned broker is not initially connected to any kafka node.
func Dial(nodeAddresses []string, conf BrokerConf) (*Broker, error) {
	broker := &Broker{
		conf:  conf,
		conns: make(map[int32]*connection),
	}

	if len(nodeAddresses) == 0 {
		return nil, errors.New("no addresses provided")
	}
	numAddresses := len(nodeAddresses)

	for i := 0; i < conf.DialRetryLimit; i++ {
		if i > 0 {
			conf.Logger.Debug("cannot fetch metadata from any connection",
				"retry", i,
				"sleep", conf.DialRetryWait)
			time.Sleep(conf.DialRetryWait)
		}

		// This iterates starting at a random location in the slice, to prevent
		// hitting the first server repeatedly
		offset := rand.Intn(numAddresses)
		for idx := 0; idx < numAddresses; idx++ {
			addr := nodeAddresses[(idx+offset)%numAddresses]

			conn, err := newTCPConnection(addr, conf.DialTimeout)
			if err != nil {
				conf.Logger.Debug("cannot connect",
					"address", addr,
					"err", err)
				continue
			}
			defer func(c *connection) {
				_ = c.Close()
			}(conn)
			resp, err := conn.Metadata(&proto.MetadataReq{
				ClientID: broker.conf.ClientID,
				Topics:   nil,
			})
			if err != nil {
				conf.Logger.Debug("cannot fetch metadata",
					"address", addr,
					"err", err)
				continue
			}
			if len(resp.Brokers) == 0 {
				conf.Logger.Debug("response with no broker data",
					"address", addr)
				continue
			}
			broker.cacheMetadata(resp)
			return broker, nil
		}
	}
	return nil, errors.New("cannot connect")
}

// Close closes the broker and all active kafka nodes connections.
func (b *Broker) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for nodeID, conn := range b.conns {
		if err := conn.Close(); err != nil {
			b.conf.Logger.Info("cannot close node connection",
				"nodeID", nodeID,
				"err", err)
		}
	}
}

func (b *Broker) Metadata() (*proto.MetadataResp, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.fetchMetadata()
}

// refreshMetadata is requesting metadata information from any node and refresh
// internal cached representation.
// Because it's changing internal state, this method requires lock protection,
// but it does not acquire nor release lock itself.
func (b *Broker) refreshMetadata() error {
	meta, err := b.fetchMetadata()
	if err == nil {
		b.cacheMetadata(meta)
	}
	return err
}

// muRefreshMetadata calls refreshMetadata, but protects it with broker's lock.
func (b *Broker) muRefreshMetadata() error {
	b.mu.Lock()
	err := b.refreshMetadata()
	b.mu.Unlock()
	return err
}

// fetchMetadata is requesting metadata information from any node and return
// protocol response if successful
//
// If "topics" are specified, only fetch metadata for those topics (can be
// used to create a topic)
//
// Because it's using metadata information to find node connections it's not
// thread safe and using it require locking.
func (b *Broker) fetchMetadata(topics ...string) (*proto.MetadataResp, error) {
	checkednodes := make(map[int32]bool)

	// try all existing connections first
	for nodeID, conn := range b.conns {
		checkednodes[nodeID] = true
		resp, err := conn.Metadata(&proto.MetadataReq{
			ClientID: b.conf.ClientID,
			Topics:   topics,
		})
		if err != nil {
			b.conf.Logger.Debug("cannot fetch metadata from node",
				"nodeID", nodeID,
				"err", err)
			continue
		}
		return resp, nil
	}

	// try all nodes that we know of that we're not connected to
	for nodeID, addr := range b.metadata.nodes {
		if _, ok := checkednodes[nodeID]; ok {
			continue
		}
		conn, err := newTCPConnection(addr, b.conf.DialTimeout)
		if err != nil {
			b.conf.Logger.Debug("cannot connect",
				"address", addr,
				"err", err)
			continue
		}
		resp, err := conn.Metadata(&proto.MetadataReq{
			ClientID: b.conf.ClientID,
			Topics:   topics,
		})

		// we had no active connection to this node, so most likely we don't need it
		_ = conn.Close()

		if err != nil {
			b.conf.Logger.Debug("cannot fetch metadata from node",
				"nodeID", nodeID,
				"err", err)
			continue
		}
		return resp, nil
	}

	return nil, errors.New("cannot fetch metadata. No topics created?")
}

// cacheMetadata creates new internal metadata representation using data from
// given response. It's call has to be protected with lock.
//
// Do not call with partial metadata response, this assumes we have the full
// set of metadata in the response
func (b *Broker) cacheMetadata(resp *proto.MetadataResp) {
	if !b.metadata.created.IsZero() {
		b.conf.Logger.Debug("rewriting old metadata",
			"age", time.Now().Sub(b.metadata.created))
	}
	b.metadata = clusterMetadata{
		created:    time.Now(),
		nodes:      make(map[int32]string),
		endpoints:  make(map[topicPartition]int32),
		partitions: make(map[string]int32),
	}
	debugmsg := make([]interface{}, 0)
	for _, node := range resp.Brokers {
		addr := fmt.Sprintf("%s:%d", node.Host, node.Port)
		b.metadata.nodes[node.NodeID] = addr
		debugmsg = append(debugmsg, node.NodeID, addr)
	}
	for _, topic := range resp.Topics {
		for _, part := range topic.Partitions {
			dest := topicPartition{topic.Name, part.ID}
			b.metadata.endpoints[dest] = part.Leader
			debugmsg = append(debugmsg, dest, part.Leader)
		}
		b.metadata.partitions[topic.Name] = int32(len(topic.Partitions))
	}
	b.conf.Logger.Debug("new metadata cached", debugmsg...)
}

// PartitionCount returns how many partitions a given topic has. If a topic
// is not known, 0 and an error are returned.
func (b *Broker) PartitionCount(topic string) (int32, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	count, ok := b.metadata.partitions[topic]
	if ok {
		return count, nil
	}

	return 0, fmt.Errorf("topic %s not found in metadata", topic)
}

// muLeaderConnection returns connection to leader for given partition. If
// connection does not exist, broker will try to connect first and add store
// connection for any further use.
//
// Failed connection retry is controlled by broker configuration.
//
// If broker is configured to allow topic creation, then if we don't find
// the leader we will return a random broker. The broker will error if we end
// up producing to it incorrectly (i.e., our metadata happened to be out of
// date).
func (b *Broker) muLeaderConnection(topic string, partition int32) (conn *connection, err error) {
	tp := topicPartition{topic, partition}

	b.mu.Lock()
	defer b.mu.Unlock()

	for retry := 0; retry < b.conf.LeaderRetryLimit; retry++ {
		if retry != 0 {
			b.mu.Unlock()
			b.conf.Logger.Debug("cannot get leader connection",
				"topic", topic,
				"partition", partition,
				"retry", retry,
				"sleep", b.conf.LeaderRetryWait.String())
			time.Sleep(b.conf.LeaderRetryWait)
			b.mu.Lock()
		}

		nodeID, ok := b.metadata.endpoints[tp]
		if !ok {
			err = b.refreshMetadata()
			if err != nil {
				b.conf.Logger.Info("cannot get leader connection: cannot refresh metadata",
					"err", err)
				continue
			}
			nodeID, ok = b.metadata.endpoints[tp]
			if !ok {
				err = proto.ErrUnknownTopicOrPartition
				// If we allow topic creation, now is the point where it is likely that this
				// is a brand new topic, so try to get metadata on it (which will trigger
				// the creation process)
				if b.conf.AllowTopicCreation {
					_, err := b.fetchMetadata(topic)
					if err != nil {
						b.conf.Logger.Info("failed to fetch metadata for new topic",
							"topic", topic,
							"err", err)
					}
				} else {
					b.conf.Logger.Info("cannot get leader connection: unknown topic or partition",
						"topic", topic,
						"partition", partition,
						"endpoint", tp)
				}
				continue
			}
		}

		conn, ok = b.conns[nodeID]
		if !ok {
			addr, ok := b.metadata.nodes[nodeID]
			if !ok {
				b.conf.Logger.Info("cannot get leader connection: no information about node",
					"nodeID", nodeID)
				err = proto.ErrBrokerNotAvailable
				continue
			}
			conn, err = newTCPConnection(addr, b.conf.DialTimeout)
			if err != nil {
				b.conf.Logger.Info("cannot get leader connection: cannot connect to node",
					"address", addr,
					"err", err)
				continue
			}
			b.conns[nodeID] = conn
		}
		return conn, nil
	}
	return nil, err
}

// coordinatorConnection returns connection to offset coordinator for given group.
//
// Failed connection retry is controlled by broker configuration.
func (b *Broker) muCoordinatorConnection(consumerGroup string) (conn *connection, resErr error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for retry := 0; retry < b.conf.LeaderRetryLimit; retry++ {
		if retry != 0 {
			b.mu.Unlock()
			time.Sleep(b.conf.LeaderRetryWait)
			b.mu.Lock()
		}

		// first try all already existing connections
		for _, conn := range b.conns {
			resp, err := conn.ConsumerMetadata(&proto.ConsumerMetadataReq{
				ClientID:      b.conf.ClientID,
				ConsumerGroup: consumerGroup,
			})
			if err != nil {
				b.conf.Logger.Debug("cannot fetch coordinator metadata",
					"consumGrp", consumerGroup,
					"err", err)
				resErr = err
				continue
			}
			if resp.Err != nil {
				b.conf.Logger.Debug("coordinator metadata response error",
					"consumGrp", consumerGroup,
					"err", resp.Err)
				resErr = err
				continue
			}

			addr := fmt.Sprintf("%s:%d", resp.CoordinatorHost, resp.CoordinatorPort)
			conn, err := newTCPConnection(addr, b.conf.DialTimeout)
			if err != nil {
				b.conf.Logger.Debug("cannot connect to node",
					"coordinatorID", resp.CoordinatorID,
					"address", addr,
					"err", err)
				resErr = err
				continue
			}
			b.conns[resp.CoordinatorID] = conn
			return conn, nil
		}

		// if none of the connections worked out, try with fresh data
		if err := b.refreshMetadata(); err != nil {
			b.conf.Logger.Debug("cannot refresh metadata",
				"err", err)
			resErr = err
			continue
		}

		for nodeID, addr := range b.metadata.nodes {
			if _, ok := b.conns[nodeID]; ok {
				// connection to node is cached so it was already checked
				continue
			}
			conn, err := newTCPConnection(addr, b.conf.DialTimeout)
			if err != nil {
				b.conf.Logger.Debug("cannot connect to node",
					"nodeID", nodeID,
					"address", addr,
					"err", err)
				resErr = err
				continue
			}
			b.conns[nodeID] = conn

			resp, err := conn.ConsumerMetadata(&proto.ConsumerMetadataReq{
				ClientID:      b.conf.ClientID,
				ConsumerGroup: consumerGroup,
			})
			if err != nil {
				b.conf.Logger.Debug("cannot fetch metadata",
					"consumGrp", consumerGroup,
					"err", err)
				resErr = err
				continue
			}
			if resp.Err != nil {
				b.conf.Logger.Debug("metadata response error",
					"consumGrp", consumerGroup,
					"err", resp.Err)
				resErr = err
				continue
			}

			addr := fmt.Sprintf("%s:%d", resp.CoordinatorHost, resp.CoordinatorPort)
			conn, err = newTCPConnection(addr, b.conf.DialTimeout)
			if err != nil {
				b.conf.Logger.Debug("cannot connect to node",
					"coordinatorID", resp.CoordinatorID,
					"address", addr,
					"err", err)
				resErr = err
				continue
			}
			b.conns[resp.CoordinatorID] = conn
			return conn, nil
		}
		resErr = proto.ErrNoCoordinator
	}
	return nil, resErr
}

// muCloseDeadConnection is closing and removing any reference to given
// connection. Because we remove dead connection, additional request to refresh
// metadata is made
//
// muCloseDeadConnection call it protected with broker's lock.
func (b *Broker) muCloseDeadConnection(conn *connection) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for nid, c := range b.conns {
		if c == conn {
			b.conf.Logger.Debug("closing dead connection",
				"nodeID", nid)
			delete(b.conns, nid)
			_ = c.Close()
			if err := b.refreshMetadata(); err != nil {
				b.conf.Logger.Debug("cannot refresh metadata",
					"err", err)
			}
			return
		}
	}
}

// offset will return offset value for given partition. Use timems to specify
// which offset value should be returned.
func (b *Broker) offset(topic string, partition int32, timems int64) (offset int64, err error) {
	conn, err := b.muLeaderConnection(topic, partition)
	if err != nil {
		return 0, err
	}
	resp, err := conn.Offset(&proto.OffsetReq{
		ClientID:  b.conf.ClientID,
		ReplicaID: -1, // any client
		Topics: []proto.OffsetReqTopic{
			{
				Name: topic,
				Partitions: []proto.OffsetReqPartition{
					{
						ID:         partition,
						TimeMs:     timems,
						MaxOffsets: 2,
					},
				},
			},
		},
	})
	if err != nil {
		if _, ok := err.(*net.OpError); ok || err == io.EOF || err == syscall.EPIPE {
			// Connection is broken, so should be closed, but the error is
			// still valid and should be returned so that retry mechanism have
			// chance to react.
			b.conf.Logger.Debug("connection died while sending message",
				"topic", topic,
				"partition", partition,
				"err", err)
			b.muCloseDeadConnection(conn)
		}
		return 0, err
	}
	found := false
	for _, t := range resp.Topics {
		if t.Name != topic {
			b.conf.Logger.Debug("unexpected topic information",
				"expected", topic,
				"got", t.Name)
			continue
		}
		for _, part := range t.Partitions {
			if part.ID != partition {
				b.conf.Logger.Debug("unexpected partition information",
					"topic", t.Name,
					"expected", partition,
					"got", part.ID)
				continue
			}
			found = true
			// happens when there are no messages
			if len(part.Offsets) == 0 {
				offset = 0
			} else {
				offset = part.Offsets[0]
			}
			err = part.Err
		}
	}
	if !found {
		return 0, errors.New("incomplete fetch response")
	}
	return offset, err
}

// OffsetEarliest returns the oldest offset available on the given partition.
func (b *Broker) OffsetEarliest(topic string, partition int32) (offset int64, err error) {
	return b.offset(topic, partition, -2)
}

// OffsetLatest return the offset of the next message produced in given partition
func (b *Broker) OffsetLatest(topic string, partition int32) (offset int64, err error) {
	return b.offset(topic, partition, -1)
}

type ProducerConf struct {
	// Compression method to use, defaulting to proto.CompressionNone.
	Compression proto.Compression

	// Timeout of single produce request. By default, 5 seconds.
	RequestTimeout time.Duration

	// Message ACK configuration. Use proto.RequiredAcksAll to require all
	// servers to write, proto.RequiredAcksLocal to wait only for leader node
	// answer or proto.RequiredAcksNone to not wait for any response.
	// Setting this to any other, greater than zero value will make producer to
	// wait for given number of servers to confirm write before returning.
	RequiredAcks int16

	// RetryLimit specify how many times message producing should be retried in
	// case of failure, before returning the error to the caller. By default
	// set to 10.
	RetryLimit int

	// RetryWait specify wait duration before produce retry after failure. By
	// default set to 200ms.
	RetryWait time.Duration

	// Logger used by producer. By default, reuse logger assigned to broker.
	Logger Logger
}

// NewProducerConf returns a default producer configuration.
func NewProducerConf() ProducerConf {
	return ProducerConf{
		Compression:    proto.CompressionNone,
		RequestTimeout: 5 * time.Second,
		RequiredAcks:   proto.RequiredAcksAll,
		RetryLimit:     10,
		RetryWait:      200 * time.Millisecond,
		Logger:         nil,
	}
}

// producer is the link to the client with extra configuration.
type producer struct {
	conf   ProducerConf
	broker *Broker
}

// Producer returns new producer instance, bound to the broker.
func (b *Broker) Producer(conf ProducerConf) Producer {
	if conf.Logger == nil {
		conf.Logger = b.conf.Logger
	}
	return &producer{
		conf:   conf,
		broker: b,
	}
}

// Produce writes messages to the given destination. Writes within the call are
// atomic, meaning either all or none of them are written to kafka.  Produce
// has a configurable amount of retries which may be attempted when common
// errors are encountered.  This behaviour can be configured with the
// RetryLimit and RetryWait attributes.
//
// Upon a successful call, the message's Offset field is updated.
func (p *producer) Produce(topic string, partition int32, messages ...*proto.Message) (offset int64, err error) {

retryLoop:
	for retry := 0; retry < p.conf.RetryLimit; retry++ {
		if retry != 0 {
			time.Sleep(p.conf.RetryWait)
		}

		offset, err = p.produce(topic, partition, messages...)

		switch err {
		case nil:
			break retryLoop
		case io.EOF, syscall.EPIPE:
			// p.produce call is closing connection when this error shows up,
			// but it's also returning it so that retry loop can count this
			// case
			// we cannot handle this error here, because there is no direct
			// access to connection
		default:
			if err := p.broker.muRefreshMetadata(); err != nil {
				p.conf.Logger.Debug("cannot refresh metadata",
					"err", err)
			}
		}
		p.conf.Logger.Debug("cannot produce messages",
			"retry", retry,
			"err", err)
	}

	if err == nil {
		// offset is the offset value of first published messages
		for i, msg := range messages {
			msg.Offset = int64(i) + offset
		}
	}

	return offset, err
}

// produce send produce request to leader for given destination.
func (p *producer) produce(topic string, partition int32, messages ...*proto.Message) (offset int64, err error) {
	conn, err := p.broker.muLeaderConnection(topic, partition)
	if err != nil {
		return 0, err
	}

	req := proto.ProduceReq{
		ClientID:     p.broker.conf.ClientID,
		Compression:  p.conf.Compression,
		RequiredAcks: p.conf.RequiredAcks,
		Timeout:      p.conf.RequestTimeout,
		Topics: []proto.ProduceReqTopic{
			{
				Name: topic,
				Partitions: []proto.ProduceReqPartition{
					{
						ID:       partition,
						Messages: messages,
					},
				},
			},
		},
	}

	resp, err := conn.Produce(&req)
	if err != nil {
		if _, ok := err.(*net.OpError); ok || err == io.EOF || err == syscall.EPIPE {
			// Connection is broken, so should be closed, but the error is
			// still valid and should be returned so that retry mechanism have
			// chance to react.
			p.conf.Logger.Debug("connection died while sending message",
				"topic", topic,
				"partition", partition,
				"err", err)
			p.broker.muCloseDeadConnection(conn)
		}
		return 0, err
	}

	// we expect single partition response
	found := false
	for _, t := range resp.Topics {
		if t.Name != topic {
			p.conf.Logger.Debug("unexpected topic information received",
				"expected", topic,
				"got", t.Name)
			continue
		}
		for _, part := range t.Partitions {
			if part.ID != partition {
				p.conf.Logger.Debug("unexpected partition information received",
					"topic", t.Name,
					"expected", partition,
					"got", part.ID)
				continue
			}
			found = true
			offset = part.Offset
			err = part.Err
		}
	}

	if !found {
		return 0, errors.New("incomplete produce response")
	}
	return offset, err
}

type ConsumerConf struct {
	// Topic name that should be consumed
	Topic string

	// Partition ID that should be consumed.
	Partition int32

	// RequestTimeout controls fetch request timeout. This operation is
	// blocking the whole connection, so it should always be set to a small
	// value. By default it's set to 50ms.
	// To control fetch function timeout use RetryLimit and RetryWait.
	RequestTimeout time.Duration

	// RetryLimit limits fetching messages a given amount of times before
	// returning ErrNoData error.
	//
	// Default is -1, which turns this limit off.
	RetryLimit int

	// RetryWait controls the duration of wait between fetch request calls,
	// when no data was returned.
	//
	// Default is 50ms.
	RetryWait time.Duration

	// RetryErrLimit limits the number of retry attempts when an error is
	// encountered.
	//
	// Default is 10.
	RetryErrLimit int

	// RetryErrWait controls the wait duration between retries after failed
	// fetch request.
	//
	// Default is 500ms.
	RetryErrWait time.Duration

	// MinFetchSize is the minimum size of messages to fetch in bytes.
	//
	// Default is 1 to fetch any message available.
	MinFetchSize int32

	// MaxFetchSize is the maximum size of data which can be sent by kafka node
	// to consumer.
	//
	// Default is 2000000 bytes.
	MaxFetchSize int32

	// Consumer cursor starting point. Set to StartOffsetNewest to receive only
	// newly created messages or StartOffsetOldest to read everything. Assign
	// any offset value to manually set cursor -- consuming starts with the
	// message whose offset is equal to given value (including first message).
	//
	// Default is StartOffsetOldest.
	StartOffset int64

	// Logger used by consumer. By default, reuse logger assigned to broker.
	Logger Logger
}

// NewConsumerConf returns the default consumer configuration.
func NewConsumerConf(topic string, partition int32) ConsumerConf {
	return ConsumerConf{
		Topic:          topic,
		Partition:      partition,
		RequestTimeout: time.Millisecond * 50,
		RetryLimit:     -1,
		RetryWait:      time.Millisecond * 50,
		RetryErrLimit:  10,
		RetryErrWait:   time.Millisecond * 500,
		MinFetchSize:   1,
		MaxFetchSize:   2000000,
		StartOffset:    StartOffsetOldest,
		Logger:         nil,
	}
}

// Consumer represents a single partition reading buffer. Consumer is also
// providing limited failure handling and message filtering.
type consumer struct {
	broker *Broker
	conf   ConsumerConf

	mu     sync.Mutex
	offset int64 // offset of next NOT consumed message
	conn   *connection
	msgbuf []*proto.Message
}

// Consumer creates a new consumer instance, bound to the broker.
func (b *Broker) Consumer(conf ConsumerConf) (Consumer, error) {
	conn, err := b.muLeaderConnection(conf.Topic, conf.Partition)
	if err != nil {
		return nil, err
	}
	if conf.Logger == nil {
		conf.Logger = b.conf.Logger
	}
	offset := conf.StartOffset
	if offset < 0 {
		switch offset {
		case StartOffsetNewest:
			off, err := b.OffsetLatest(conf.Topic, conf.Partition)
			if err != nil {
				return nil, err
			}
			offset = off
		case StartOffsetOldest:
			off, err := b.OffsetEarliest(conf.Topic, conf.Partition)
			if err != nil {
				return nil, err
			}
			offset = off
		default:
			return nil, fmt.Errorf("invalid start offset: %d", conf.StartOffset)
		}
	}
	c := &consumer{
		broker: b,
		conn:   conn,
		conf:   conf,
		msgbuf: make([]*proto.Message, 0),
		offset: offset,
	}
	return c, nil
}

// Consume is returning single message from consumed partition. Consumer can
// retry fetching messages even if responses return no new data. Retry
// behaviour can be configured through RetryLimit and RetryWait consumer
// parameters.
//
// Consume can retry sending request on common errors. This behaviour can be
// configured with RetryErrLimit and RetryErrWait consumer configuration
// attributes.
func (c *consumer) Consume() (*proto.Message, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var retry int
	for len(c.msgbuf) == 0 {
		var err error
		c.msgbuf, err = c.fetch()
		if err != nil {
			return nil, err
		}
		if len(c.msgbuf) == 0 {
			if c.conf.RetryWait > 0 {
				time.Sleep(c.conf.RetryWait)
			}
			retry += 1
			if c.conf.RetryLimit != -1 && retry > c.conf.RetryLimit {
				return nil, ErrNoData
			}
		}
	}

	msg := c.msgbuf[0]
	c.msgbuf = c.msgbuf[1:]
	c.offset = msg.Offset + 1
	return msg, nil
}

// fetch and return next batch of messages. In case of certain set of errors,
// retry sending fetch request. Retry behaviour can be configured with
// RetryErrLimit and RetryErrWait consumer configuration attributes.
func (c *consumer) fetch() ([]*proto.Message, error) {
	req := proto.FetchReq{
		ClientID:    c.broker.conf.ClientID,
		MaxWaitTime: c.conf.RequestTimeout,
		MinBytes:    c.conf.MinFetchSize,
		Topics: []proto.FetchReqTopic{
			{
				Name: c.conf.Topic,
				Partitions: []proto.FetchReqPartition{
					{
						ID:          c.conf.Partition,
						FetchOffset: c.offset,
						MaxBytes:    c.conf.MaxFetchSize,
					},
				},
			},
		},
	}

	var resErr error
consumeRetryLoop:
	for retry := 0; retry < c.conf.RetryErrLimit; retry++ {
		if retry != 0 {
			time.Sleep(c.conf.RetryErrWait)
		}

		if c.conn == nil {
			conn, err := c.broker.muLeaderConnection(c.conf.Topic, c.conf.Partition)
			if err != nil {
				resErr = err
				continue
			}
			c.conn = conn
		}

		resp, err := c.conn.Fetch(&req)
		resErr = err

		if _, ok := err.(*net.OpError); ok || err == io.EOF || err == syscall.EPIPE {
			c.conf.Logger.Debug("connection died while fetching message",
				"topic", c.conf.Topic,
				"partition", c.conf.Partition,
				"err", err)
			c.broker.muCloseDeadConnection(c.conn)
			c.conn = nil
			continue
		}

		if err != nil {
			c.conf.Logger.Debug("cannot fetch messages: unknown error",
				"retry", retry,
				"err", err)
			c.broker.muCloseDeadConnection(c.conn)
			c.conn = nil
			continue
		}

		for _, topic := range resp.Topics {
			if topic.Name != c.conf.Topic {
				c.conf.Logger.Warn("unexpected topic information received",
					"got", topic.Name,
					"expected", c.conf.Topic)
				continue
			}
			for _, part := range topic.Partitions {
				if part.ID != c.conf.Partition {
					c.conf.Logger.Warn("unexpected partition information received",
						"topic", topic.Name,
						"expected", c.conf.Partition,
						"got", part.ID)
					continue
				}
				switch part.Err {
				case proto.ErrLeaderNotAvailable, proto.ErrNotLeaderForPartition, proto.ErrBrokerNotAvailable:
					c.conf.Logger.Debug("cannot fetch messages",
						"retry", retry,
						"err", part.Err)
					if err := c.broker.muRefreshMetadata(); err != nil {
						c.conf.Logger.Debug("cannot refresh metadata",
							"err", err)
					}
					// The connection is fine, so don't close it,
					// but we may very well need to talk to a different broker now.
					// Set the conn to nil so that next time around the loop
					// we'll check the metadata again to see who we're supposed to talk to.
					c.conn = nil
					continue consumeRetryLoop
				}
				return part.Messages, part.Err
			}
		}
		return nil, errors.New("incomplete fetch response")
	}

	return nil, resErr
}

type OffsetCoordinatorConf struct {
	ConsumerGroup string

	// RetryErrLimit limits messages fetch retry upon failure. By default 10.
	RetryErrLimit int

	// RetryErrWait controls wait duration between retries after failed fetch
	// request. By default 500ms.
	RetryErrWait time.Duration

	// Logger used by consumer. By default, reuse logger assigned to broker.
	Logger Logger
}

// NewOffsetCoordinatorConf returns default OffsetCoordinator configuration.
func NewOffsetCoordinatorConf(consumerGroup string) OffsetCoordinatorConf {
	return OffsetCoordinatorConf{
		ConsumerGroup: consumerGroup,
		RetryErrLimit: 10,
		RetryErrWait:  time.Millisecond * 500,
		Logger:        nil,
	}
}

type offsetCoordinator struct {
	conf   OffsetCoordinatorConf
	broker *Broker

	mu   sync.Mutex
	conn *connection
}

// OffsetCoordinator returns offset management coordinator for single consumer
// group, bound to broker.
func (b *Broker) OffsetCoordinator(conf OffsetCoordinatorConf) (OffsetCoordinator, error) {
	conn, err := b.muCoordinatorConnection(conf.ConsumerGroup)
	if err != nil {
		return nil, err
	}
	if conf.Logger == nil {
		conf.Logger = b.conf.Logger
	}
	c := &offsetCoordinator{
		broker: b,
		conf:   conf,
		conn:   conn,
	}
	return c, nil
}

// Commit is saving offset information for given topic and partition.
//
// Commit can retry saving offset information on common errors. This behaviour
// can be configured with with RetryErrLimit and RetryErrWait coordinator
// configuration attributes.
func (c *offsetCoordinator) Commit(topic string, partition int32, offset int64) error {
	return c.commit(topic, partition, offset, "")
}

// Commit works exactly like Commit method, but store extra metadata string
// together with offset information.
func (c *offsetCoordinator) CommitFull(topic string, partition int32, offset int64, metadata string) error {
	return c.commit(topic, partition, offset, metadata)
}

// commit is saving offset and metadata information. Provides limited error
// handling configurable through OffsetCoordinatorConf.
func (c *offsetCoordinator) commit(topic string, partition int32, offset int64, metadata string) (resErr error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for retry := 0; retry < c.conf.RetryErrLimit; retry++ {
		if retry != 0 {
			c.mu.Unlock()
			time.Sleep(c.conf.RetryErrWait)
			c.mu.Lock()
		}

		// connection can be set to nil if previously reference connection died
		if c.conn == nil {
			conn, err := c.broker.muCoordinatorConnection(c.conf.ConsumerGroup)
			if err != nil {
				resErr = err
				c.conf.Logger.Debug("cannot connect to coordinator",
					"consumGrp", c.conf.ConsumerGroup,
					"err", err)
				continue
			}
			c.conn = conn
		}

		resp, err := c.conn.OffsetCommit(&proto.OffsetCommitReq{
			ClientID:      c.broker.conf.ClientID,
			ConsumerGroup: c.conf.ConsumerGroup,
			Topics: []proto.OffsetCommitReqTopic{
				{
					Name: topic,
					Partitions: []proto.OffsetCommitReqPartition{
						{ID: partition, Offset: offset, TimeStamp: time.Now(), Metadata: metadata},
					},
				},
			},
		})
		resErr = err

		if _, ok := err.(*net.OpError); ok || err == io.EOF || err == syscall.EPIPE {
			c.conf.Logger.Debug("connection died while commiting",
				"topic", topic,
				"partition", partition,
				"consumGrp", c.conf.ConsumerGroup)
			c.broker.muCloseDeadConnection(c.conn)
			c.conn = nil
		} else if err == nil {
			for _, t := range resp.Topics {
				if t.Name != topic {
					c.conf.Logger.Debug("unexpected topic information received",
						"got", t.Name,
						"expected", topic)
					continue

				}
				for _, part := range t.Partitions {
					if part.ID != partition {
						c.conf.Logger.Debug("unexpected partition information received",
							"topic", topic,
							"got", part.ID,
							"expected", partition)
						continue
					}
					return part.Err
				}
			}
			return errors.New("response does not contain commit information")
		}
	}
	return resErr
}

// Offset is returning last offset and metadata information committed for given
// topic and partition.
// Offset can retry sending request on common errors. This behaviour can be
// configured with with RetryErrLimit and RetryErrWait coordinator
// configuration attributes.
func (c *offsetCoordinator) Offset(topic string, partition int32) (offset int64, metadata string, resErr error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for retry := 0; retry < c.conf.RetryErrLimit; retry++ {
		if retry != 0 {
			c.mu.Unlock()
			time.Sleep(c.conf.RetryErrWait)
			c.mu.Lock()
		}

		// connection can be set to nil if previously reference connection died
		if c.conn == nil {
			conn, err := c.broker.muCoordinatorConnection(c.conf.ConsumerGroup)
			if err != nil {
				c.conf.Logger.Debug("cannot connect to coordinator",
					"consumGrp", c.conf.ConsumerGroup,
					"err", err)
				resErr = err
				continue
			}
			c.conn = conn
		}
		resp, err := c.conn.OffsetFetch(&proto.OffsetFetchReq{
			ConsumerGroup: c.conf.ConsumerGroup,
			Topics: []proto.OffsetFetchReqTopic{
				{
					Name:       topic,
					Partitions: []int32{partition},
				},
			},
		})
		resErr = err

		switch err {
		case io.EOF, syscall.EPIPE:
			c.conf.Logger.Debug("connection died while fetching offset",
				"topic", topic,
				"partition", partition,
				"consumGrp", c.conf.ConsumerGroup)
			c.broker.muCloseDeadConnection(c.conn)
			c.conn = nil
		case nil:
			for _, t := range resp.Topics {
				if t.Name != topic {
					c.conf.Logger.Debug("unexpected topic information received",
						"got", t.Name,
						"expected", topic)
					continue
				}
				for _, part := range t.Partitions {
					if part.ID != partition {
						c.conf.Logger.Debug("unexpected partition information received",
							"topic", topic,
							"expected", partition,
							"get", part.ID)
						continue
					}
					if part.Err != nil {
						return 0, "", part.Err
					}
					return part.Offset, part.Metadata, nil
				}
			}
			return 0, "", errors.New("response does not contain offset information")
		}
	}

	return 0, "", resErr
}

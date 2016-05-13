package proto

import (
	"fmt"
)

var (
	ErrUnknown                 = &KafkaError{-1, "unknown"}
	ErrOffsetOutOfRange        = &KafkaError{1, "offset out of range"}
	ErrInvalidMessage          = &KafkaError{2, "invalid message"}
	ErrUnknownTopicOrPartition = &KafkaError{3, "unknown topic or partition"}
	ErrInvalidMessageSize      = &KafkaError{4, "invalid message size"}
	ErrLeaderNotAvailable      = &KafkaError{5, "leader not available"}
	ErrNotLeaderForPartition   = &KafkaError{6, "not leader for partition"}
	ErrRequestTimeout          = &KafkaError{7, "request timeed out"}
	ErrBrokerNotAvailable      = &KafkaError{8, "broker not available"}
	ErrReplicaNotAvailable     = &KafkaError{9, "replica not available"}
	ErrMessageSizeTooLarge     = &KafkaError{10, "message size too large"}
	ErrScaleControllerEpoch    = &KafkaError{11, "scale controller epoch"}
	ErrOffsetMetadataTooLarge  = &KafkaError{12, "offset metadata too large"}
	ErrOffsetLoadInProgress    = &KafkaError{14, "offsets load in progress"}
	ErrNoCoordinator           = &KafkaError{15, "consumer coordinator not available"}
	ErrNotCoordinator          = &KafkaError{16, "not coordinator for consumer"}

	errnoToErr = map[int16]error{
		-1: ErrUnknown,
		1:  ErrOffsetOutOfRange,
		2:  ErrInvalidMessage,
		3:  ErrUnknownTopicOrPartition,
		4:  ErrInvalidMessageSize,
		5:  ErrLeaderNotAvailable,
		6:  ErrNotLeaderForPartition,
		7:  ErrRequestTimeout,
		8:  ErrBrokerNotAvailable,
		9:  ErrReplicaNotAvailable,
		10: ErrMessageSizeTooLarge,
		11: ErrScaleControllerEpoch,
		12: ErrOffsetMetadataTooLarge,
		14: ErrOffsetLoadInProgress,
		15: ErrNoCoordinator,
		16: ErrNotCoordinator,
	}
)

type KafkaError struct {
	errno   int16
	message string
}

func (err *KafkaError) Error() string {
	return fmt.Sprintf("%s (%d)", err.message, err.errno)
}

func (err *KafkaError) Errno() int {
	return int(err.errno)
}

func errFromNo(errno int16) error {
	if errno == 0 {
		return nil
	}
	err, ok := errnoToErr[errno]
	if !ok {
		return fmt.Errorf("unknown kafka error %d", errno)
	}
	return err
}

package sarama

type ListGroupsRequest struct {
}

func (r *ListGroupsRequest) encode(pe packetEncoder) error {
	return nil
}

func (r *ListGroupsRequest) decode(pd packetDecoder) (err error) {
	return nil
}

func (r *ListGroupsRequest) key() int16 {
	return 16
}

func (r *ListGroupsRequest) version() int16 {
	return 0
}

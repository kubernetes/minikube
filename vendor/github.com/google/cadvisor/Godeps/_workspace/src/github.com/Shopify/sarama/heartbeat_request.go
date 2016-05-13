package sarama

type HeartbeatRequest struct {
	GroupId      string
	GenerationId int32
	MemberId     string
}

func (r *HeartbeatRequest) encode(pe packetEncoder) error {
	if err := pe.putString(r.GroupId); err != nil {
		return err
	}

	pe.putInt32(r.GenerationId)

	if err := pe.putString(r.MemberId); err != nil {
		return err
	}

	return nil
}

func (r *HeartbeatRequest) decode(pd packetDecoder) (err error) {
	if r.GroupId, err = pd.getString(); err != nil {
		return
	}
	if r.GenerationId, err = pd.getInt32(); err != nil {
		return
	}
	if r.MemberId, err = pd.getString(); err != nil {
		return
	}

	return nil
}

func (r *HeartbeatRequest) key() int16 {
	return 12
}

func (r *HeartbeatRequest) version() int16 {
	return 0
}

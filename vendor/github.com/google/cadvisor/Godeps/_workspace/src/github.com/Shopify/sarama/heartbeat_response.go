package sarama

type HeartbeatResponse struct {
	Err KError
}

func (r *HeartbeatResponse) encode(pe packetEncoder) error {
	pe.putInt16(int16(r.Err))
	return nil
}

func (r *HeartbeatResponse) decode(pd packetDecoder) error {
	if kerr, err := pd.getInt16(); err != nil {
		return err
	} else {
		r.Err = KError(kerr)
	}

	return nil
}

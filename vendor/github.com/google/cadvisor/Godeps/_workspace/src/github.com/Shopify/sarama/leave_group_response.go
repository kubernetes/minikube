package sarama

type LeaveGroupResponse struct {
	Err KError
}

func (r *LeaveGroupResponse) encode(pe packetEncoder) error {
	pe.putInt16(int16(r.Err))
	return nil
}

func (r *LeaveGroupResponse) decode(pd packetDecoder) (err error) {
	if kerr, err := pd.getInt16(); err != nil {
		return err
	} else {
		r.Err = KError(kerr)
	}

	return nil
}

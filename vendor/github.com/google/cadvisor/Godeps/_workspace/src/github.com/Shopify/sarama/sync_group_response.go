package sarama

type SyncGroupResponse struct {
	Err              KError
	MemberAssignment []byte
}

func (r *SyncGroupResponse) encode(pe packetEncoder) error {
	pe.putInt16(int16(r.Err))
	return pe.putBytes(r.MemberAssignment)
}

func (r *SyncGroupResponse) decode(pd packetDecoder) (err error) {
	if kerr, err := pd.getInt16(); err != nil {
		return err
	} else {
		r.Err = KError(kerr)
	}

	r.MemberAssignment, err = pd.getBytes()
	return
}

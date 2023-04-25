package egoscale

func (*ListResourceDetails) name() string {
	return "listResourceDetails"
}

func (*ListResourceDetails) response() interface{} {
	return new(ListResourceDetailsResponse)
}

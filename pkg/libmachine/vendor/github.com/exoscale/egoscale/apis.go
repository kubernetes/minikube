package egoscale

func (*ListAPIs) name() string {
	return "listApis"
}

func (*ListAPIs) response() interface{} {
	return new(ListAPIsResponse)
}

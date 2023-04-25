package egoscale

func (*ListNetworkOfferings) name() string {
	return "listNetworkOfferings"
}

func (*ListNetworkOfferings) response() interface{} {
	return new(ListNetworkOfferingsResponse)
}

func (*UpdateNetworkOffering) name() string {
	return "updateNetworkOffering"
}

func (*UpdateNetworkOffering) response() interface{} {
	return new(UpdateNetworkOfferingResponse)
}

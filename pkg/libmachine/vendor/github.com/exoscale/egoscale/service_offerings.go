package egoscale

// name returns the CloudStack API command name
func (*ListServiceOfferings) name() string {
	return "listServiceOfferings"
}

func (*ListServiceOfferings) response() interface{} {
	return new(ListServiceOfferingsResponse)
}

// ListServiceOfferingsResponse represents a list of service offerings
type ListServiceOfferingsResponse struct {
	Count           int               `json:"count"`
	ServiceOffering []ServiceOffering `json:"serviceoffering"`
}

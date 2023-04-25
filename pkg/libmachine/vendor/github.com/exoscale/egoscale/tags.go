package egoscale

// name returns the CloudStack API command name
func (*CreateTags) name() string {
	return "createTags"
}

func (*CreateTags) asyncResponse() interface{} {
	return new(booleanResponse)
}

// name returns the CloudStack API command name
func (*DeleteTags) name() string {
	return "deleteTags"
}

func (*DeleteTags) asyncResponse() interface{} {
	return new(booleanResponse)
}

// name returns the CloudStack API command name
func (*ListTags) name() string {
	return "listTags"
}

func (*ListTags) response() interface{} {
	return new(ListTagsResponse)
}

package egoscale

func (*RegisterUserKeys) name() string {
	return "registerUserKeys"
}

func (*RegisterUserKeys) response() interface{} {
	return new(RegisterUserKeysResponse)
}

func (*CreateUser) name() string {
	return "createUser"
}

func (*CreateUser) response() interface{} {
	return new(CreateUserResponse)
}

func (*UpdateUser) name() string {
	return "updateUser"
}

func (*UpdateUser) response() interface{} {
	return new(UpdateUserResponse)
}

func (*ListUsers) name() string {
	return "listUsers"
}

func (*ListUsers) response() interface{} {
	return new(ListUsersResponse)
}

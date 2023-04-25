package egoscale

// User represents a User
type User struct {
	APIKey              string `json:"apikey,omitempty" doc:"the api key of the user"`
	Account             string `json:"account,omitempty" doc:"the account name of the user"`
	AccountID           string `json:"accountid,omitempty" doc:"the account ID of the user"`
	AccountType         int16  `json:"accounttype,omitempty" doc:"the account type of the user"`
	Created             string `json:"created,omitempty" doc:"the date and time the user account was created"`
	Domain              string `json:"domain,omitempty" doc:"the domain name of the user"`
	DomainID            string `json:"domainid,omitempty" doc:"the domain ID of the user"`
	Email               string `json:"email,omitempty" doc:"the user email address"`
	FirstName           string `json:"firstname,omitempty" doc:"the user firstname"`
	ID                  string `json:"id,omitempty" doc:"the user ID"`
	IsCallerChildDomain bool   `json:"iscallerchilddomain,omitempty" doc:"the boolean value representing if the updating target is in caller's child domain"`
	IsDefault           bool   `json:"isdefault,omitempty" doc:"true if user is default, false otherwise"`
	LastName            string `json:"lastname,omitempty" doc:"the user lastname"`
	RoleID              string `json:"roleid,omitempty" doc:"the ID of the role"`
	RoleName            string `json:"rolename,omitempty" doc:"the name of the role"`
	RoleType            string `json:"roletype,omitempty" doc:"the type of the role"`
	SecretKey           string `json:"secretkey,omitempty" doc:"the secret key of the user"`
	State               string `json:"state,omitempty" doc:"the user state"`
	Timezone            string `json:"timezone,omitempty" doc:"the timezone user was created in"`
	UserName            string `json:"username,omitempty" doc:"the user name"`
}

// RegisterUserKeys registers a new set of key of the given user
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/registerUserKeys.html
type RegisterUserKeys struct {
	ID string `json:"id" doc:"User id"`
}

// RegisterUserKeysResponse represents a new set of UserKeys
//
// NB: only the APIKey and SecretKey will be filled, hence the different key name
type RegisterUserKeysResponse struct {
	UserKeys User `json:"userkeys"`
}

// CreateUser represents the creation of a User
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/createUser.html
type CreateUser struct {
	Account   string `json:"account" doc:"Creates the user under the specified account. If no account is specified, the username will be used as the account name."`
	Email     string `json:"email" doc:"email"`
	FirstName string `json:"firstname" doc:"firstname"`
	LastName  string `json:"lastname" doc:"lastname"`
	Password  string `json:"password" doc:"Clear text password (Default hashed to SHA256SALT). If you wish to use any other hashing algorithm, you would need to write a custom authentication adapter See Docs section."`
	UserName  string `json:"username" doc:"Unique username."`
	DomainID  string `json:"domainid,omitempty" doc:"Creates the user under the specified domain. Has to be accompanied with the account parameter"`
	Timezone  string `json:"timezone,omitempty" doc:"Specifies a timezone for this command. For more information on the timezone parameter, see Time Zone Format."`
	UserID    string `json:"userid,omitempty" doc:"User UUID, required for adding account from external provisioning system"`
}

// CreateUserResponse represents the freshly created User
type CreateUserResponse struct {
	User User `json:"user"`
}

// UpdateUser represents the modification of a User
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/updateUser.html
type UpdateUser struct {
	ID            string `json:"id" doc:"User uuid"`
	Email         string `json:"email,omitempty" doc:"email"`
	FirstName     string `json:"firstname,omitempty" doc:"first name"`
	LastName      string `json:"lastname,omitempty" doc:"last name"`
	Password      string `json:"password,omitempty" doc:"Clear text password (default hashed to SHA256SALT). If you wish to use any other hashing algorithm, you would need to write a custom authentication adapter. Can't be passed when command is executed via integration.api.port"`
	Timezone      string `json:"timezone,omitempty" doc:"Specifies a timezone for this command. For more information on the timezone parameter, see Time Zone Format."`
	UserAPIKey    string `json:"userapikey,omitempty" doc:"The API key for the user. Must be specified with userSecretKey"`
	UserName      string `json:"username,omitempty" doc:"Unique username"`
	UserSecretKey string `json:"usersecretkey,omitempty" doc:"The secret key for the user. Must be specified with userApiKey"`
}

// UpdateUserResponse represents the freshly modified User
type UpdateUserResponse CreateUserResponse

// ListUsers represents the search for Users
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/listUsers.html
type ListUsers struct {
	Account     string `json:"account,omitempty" doc:"list resources by account. Must be used with the domainId parameter."`
	AccountType int64  `json:"accounttype,omitempty" doc:"List users by account type. Valid types include admin, domain-admin, read-only-admin, or user."`
	DomainID    string `json:"domainid,omitempty" doc:"list only resources belonging to the domain specified"`
	ID          string `json:"id,omitempty" doc:"List user by ID."`
	IsRecursive bool   `json:"isrecursive,omitempty" doc:"defaults to false, but if true, lists all resources from the parent specified by the domainId till leaves."`
	Keyword     string `json:"keyword,omitempty" doc:"List by keyword"`
	ListAll     bool   `json:"listall,omitempty" doc:"If set to false, list only resources belonging to the command's caller; if set to true - list resources that the caller is authorized to see. Default value is false"`
	Page        int    `json:"page,omitempty"`
	PageSize    int    `json:"pagesize,omitempty"`
	State       string `json:"state,omitempty" doc:"List users by state of the user account."`
	Username    string `json:"username,omitempty" doc:"List user by the username"`
}

// ListUsersResponse represents a list of users
type ListUsersResponse struct {
	Count int    `json:"count"`
	User  []User `json:"user"`
}

package egoscale

// API represents an API service
type API struct {
	Description string     `json:"description,omitempty" doc:"description of the api"`
	IsAsync     bool       `json:"isasync,omitempty" doc:"true if api is asynchronous"`
	Name        string     `json:"name,omitempty" doc:"the name of the api command"`
	Related     string     `json:"related,omitempty" doc:"comma separated related apis"`
	Since       string     `json:"since,omitempty" doc:"version of CloudStack the api was introduced in"`
	Type        string     `json:"type,omitempty" doc:"response field type"`
	Params      []APIParam `json:"params,omitempty" doc:"the list params the api accepts"`
	Response    []APIParam `json:"response,omitempty" doc:"api response fields"`
}

// APIParam represents an API parameter field
type APIParam struct {
	Description string `json:"description"`
	Length      int64  `json:"length"`
	Name        string `json:"name"`
	Required    bool   `json:"required"`
	Since       string `json:"since"`
	Type        string `json:"type"`
}

// ListAPIs represents a query to list the api
//
// CloudStack API: https://cloudstack.apache.org/api/apidocs-4.10/apis/listApis.html
type ListAPIs struct {
	Name string `json:"name,omitempty" doc:"API name"`
}

// ListAPIsResponse represents a list of API
type ListAPIsResponse struct {
	Count int   `json:"count"`
	API   []API `json:"api"`
}

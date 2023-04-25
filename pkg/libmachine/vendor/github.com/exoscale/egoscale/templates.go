package egoscale

import "fmt"

// ListRequest builds the ListTemplates request
func (temp *Template) ListRequest() (ListCommand, error) {
	req := &ListTemplates{
		Name:       temp.Name,
		Account:    temp.Account,
		DomainID:   temp.DomainID,
		ID:         temp.ID,
		ProjectID:  temp.ProjectID,
		ZoneID:     temp.ZoneID,
		Hypervisor: temp.Hypervisor,
		//TODO Tags
	}
	if temp.IsFeatured {
		req.TemplateFilter = "featured"
	}
	if temp.Removed != "" {
		*req.ShowRemoved = true
	}

	return req, nil
}

func (*ListTemplates) each(resp interface{}, callback IterateItemFunc) {
	temps, ok := resp.(*ListTemplatesResponse)
	if !ok {
		callback(nil, fmt.Errorf("ListTemplatesResponse expected, got %t", resp))
		return
	}

	for i := range temps.Template {
		if !callback(&temps.Template[i], nil) {
			break
		}
	}
}

// SetPage sets the current page
func (ls *ListTemplates) SetPage(page int) {
	ls.Page = page
}

// SetPageSize sets the page size
func (ls *ListTemplates) SetPageSize(pageSize int) {
	ls.PageSize = pageSize
}

// ResourceType returns the type of the resource
func (*Template) ResourceType() string {
	return "Template"
}

// name returns the CloudStack API command name
func (*ListTemplates) name() string {
	return "listTemplates"
}

func (*ListTemplates) response() interface{} {
	return new(ListTemplatesResponse)
}

package request

type Common interface {
	GetRegion() string
	SetRegion(string)

	GetProjectId() string
	SetProjectId(string)
}

type CommonBase struct {
	Region string

	ProjectId string
}

// GetRegion will return region of request
func (c *CommonBase) GetRegion() string {
	return c.Region
}

// SetRegion will set region of request
func (c *CommonBase) SetRegion(region string) {
	c.Region = region
}

// GetProjectId will get project id of request
func (c *CommonBase) GetProjectId() string {
	return c.ProjectId
}

// SetProjectId will set project id of request
func (c *CommonBase) SetProjectId(projectId string) {
	c.ProjectId = projectId
}

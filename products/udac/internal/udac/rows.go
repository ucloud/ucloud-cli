package udac

// importedInstanceRow is the row format for list output
type importedInstanceRow struct {
	ResourceID string `json:"resourceId" header:"ResourceID"`
	InstanceID string `json:"instanceId" header:"InstanceID"`
	Name       string `json:"name" header:"Name"`
	Type       string `json:"type" header:"Type"`
	Status     string `json:"status" header:"Status"`
	ImportTime string `json:"importTime" header:"ImportTime"`
	Region     string `json:"region" header:"Region"`
	Zone       string `json:"zone" header:"Zone"`
}

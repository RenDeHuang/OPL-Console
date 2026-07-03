package console

type UserView struct {
	ID     string `json:"id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Status string `json:"status"`
}

type OrganizationView struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type Me struct {
	User         UserView         `json:"user"`
	Organization OrganizationView `json:"organization"`
}

type Package struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	CPU               int    `json:"cpu"`
	MemoryGB          int    `json:"memoryGb"`
	StorageGB         int    `json:"storageGb"`
	ComputeHourlyFen  int64  `json:"computeHourlyFen"`
	StorageGBMonthFen int64  `json:"storageGbMonthFen"`
	Available         bool   `json:"available"`
}

type ManagedWorkspace struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	State    string `json:"state"`
	Policy   string `json:"policy"`
	URL      string `json:"url,omitempty"`
	Provider string `json:"provider,omitempty"`
}

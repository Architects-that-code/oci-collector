package children

type TenancyCollector struct {
	TenancyId         string `json:"ocid"`
	TenancyName       string `json:"name"`
	TenancyConfigured bool   `json:"configured"`
	GovernanceStatus  string `json:"governanceStatus"`
	LifecycleState    string `json:"lifecycleState"`
}

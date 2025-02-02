package release

type Asset struct {
	Contents string
	URL      string
	Name     string
}

type Release struct {
	Project string   `json:"project" bson:"project"`
	Runs    []string `json:"runs" bson:"runs"`
	Status  string   `json:"status" bson:"status"`
	Assets  []Asset  `json:"assets" bson:"assets"`
}

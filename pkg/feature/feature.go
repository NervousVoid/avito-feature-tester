package feature

type Template struct {
	FeatureSlug      string        `json:"feature_slug"`
	Features         []string      `json:"features"`
	UserID           int           `json:"user_id"`
	AssignFeatures   []interface{} `json:"assign_features"`
	UnassignFeatures []interface{} `json:"unassign_features"`
	Fraction         int           `json:"fraction"`
}

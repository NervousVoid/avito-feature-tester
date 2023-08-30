package feature

type Template struct {
	FeatureSlug      string   `json:"feature_slug,omitempty"`
	Features         []string `json:"features,omitempty"`
	UserID           int      `json:"user_id,omitempty"`
	AssignFeatures   []string `json:"assign_features,omitempty"`
	UnassignFeatures []string `json:"unassign_features,omitempty"`
	Fraction         int      `json:"fraction,omitempty"`
}

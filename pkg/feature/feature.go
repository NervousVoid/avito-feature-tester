package feature

type Template struct {
	FeatureSlug         string        `json:"feature_slug"`
	UserID              int           `json:"user_id"`
	AddFeaturesSlugs    []interface{} `json:"add_features_slugs"`
	DeleteFeaturesSlugs []interface{} `json:"delete_features_slugs"`
}

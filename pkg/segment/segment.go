package segment

type Template struct {
	SegmentSlug      string   `json:"segment_slug,omitempty"`
	Segments         []string `json:"segments,omitempty"`
	UserID           int      `json:"user_id,omitempty"`
	AssignSegments   []string `json:"assign_segments,omitempty"`
	UnassignSegments []string `json:"unassign_segments,omitempty"`
	Fraction         int      `json:"fraction,omitempty"`
}

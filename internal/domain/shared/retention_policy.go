package shared

// Value Objects

type RetentionPolicy string

const (
	RetentionPolicyKeepLast RetentionPolicy = "keep_last"
	RetentionPolicyMaxAge   RetentionPolicy = "max_age"
)

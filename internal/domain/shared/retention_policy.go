package shared

type RetentionPolicy string

const (
	RetentionPolicyNone     RetentionPolicy = "none"
	RetentionPolicyKeepLast RetentionPolicy = "keep_last"
	RetentionPolicyMaxAge   RetentionPolicy = "max_age"
)

type RetentionConfig struct {
	Policy     RetentionPolicy
	KeepLastN  int
	MaxAgeDays int
}

func NoRetention() RetentionConfig {
	return RetentionConfig{Policy: RetentionPolicyNone}
}

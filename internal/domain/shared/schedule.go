package shared

type Schedule struct {
	CronExpression string
	Timezone       string
}

func NewCronSchedule(cronExpression string, timezone string) Schedule {
	return Schedule{
		CronExpression: cronExpression,
		Timezone:       timezone,
	}
}

func (s Schedule) IsZero() bool {
	return s.CronExpression == "" && s.Timezone == ""
}

func (s Schedule) String() string {
	return s.CronExpression
}

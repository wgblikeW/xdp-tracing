package options

// RedisOptions defines options for redis clutser.
type RedisOptions struct {
}

func NewRedisOptions() *RedisOptions {
	return &RedisOptions{}
}

func Validate() []error {
	errs := []error{}

	return errs
}

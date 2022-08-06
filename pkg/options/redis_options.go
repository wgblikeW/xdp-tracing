// Copyright 2022 p1nant0m <wgblike@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

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

package main

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type config struct {
	Bucket   string   `required:"true"`
	Prefix   string   `default:"postgresql-backups/"`
	Schedule schedule `default:"00:00"`
	Args     []string `default:"--clean,--if-exists"`
}

func parseConfig() *config {
	c := new(config)
	envconfig.MustProcess("pgez", c)
	return c
}

type schedule struct {
	Hour   int
	Minute int
}

func (s *schedule) Decode(value string) error {
	var h, m uint
	n, err := fmt.Sscanf(value, "%d:%d", &h, &m)
	if err != nil {
		return err
	}
	if n != 2 || h > 23 || m > 59 {
		return fmt.Errorf("invalid time: %s", value)
	}

	s.Hour, s.Minute = int(h), int(m)
	return nil
}

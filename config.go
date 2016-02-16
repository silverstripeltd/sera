package main

// Config holds configuration from `/etc/sera.json`
type Config struct {
	Server  string `json:"server"`
	Syslog  bool   `json:"syslog"`
	Verbose bool   `json:"verbose"`
}

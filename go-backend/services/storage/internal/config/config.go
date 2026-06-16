package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	BackupDir string `json:",default=./data/backups"`
}

package cmd

import "github.com/kiry163/image-cli/pkg/config"

func CurrentConfig() config.Config {
	return appConfig
}

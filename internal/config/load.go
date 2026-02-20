package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

//go:embed default.yaml
var embeddedDefault []byte

func Load(pathArg, cwd string) (*Config, string, error) {
	if cwd == "" {
		return nil, "", fmt.Errorf("当前目录为空，没法继续")
	}

	var (
		raw    []byte
		source string
		err    error
	)

	if pathArg != "" {
		target := pathArg
		if !filepath.IsAbs(target) {
			target = filepath.Join(cwd, target)
		}
		raw, err = os.ReadFile(target)
		if err != nil {
			return nil, "", fmt.Errorf("读取配置文件失败（%s）：%w", target, err)
		}
		source = target
	} else {
		project := filepath.Join(cwd, "syl-md2ppt.yaml")
		if st, statErr := os.Stat(project); statErr == nil && !st.IsDir() {
			raw, err = os.ReadFile(project)
			if err != nil {
				return nil, "", fmt.Errorf("读取配置文件失败（%s）：%w", project, err)
			}
			source = project
		} else {
			raw = embeddedDefault
			source = "embedded:default.yaml"
		}
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(raw, cfg); err != nil {
		return nil, "", fmt.Errorf("配置文件格式读不懂（%s）：%w", source, err)
	}
	cfg.applyDefaults()
	return cfg, source, nil
}

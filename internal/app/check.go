package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"syl-md2ppt/internal/config"
	"syl-md2ppt/internal/discovery"
)

type CheckResult struct {
	PairCount     int
	WarningCount  int
	ConflictCount int
	HasConflict   bool
	Warnings      []string
	Items         []CheckItem
	ConfigSource  string
}

type CheckItem struct {
	No     int
	ENPath string
	CNPath string
}

func Check(opts Options) (CheckResult, error) {
	if strings.TrimSpace(opts.SourceDir) == "" {
		return CheckResult{}, fmt.Errorf("还没给数据源目录")
	}

	cwd := opts.CWD
	if cwd == "" {
		wd, err := os.Getwd()
		if err != nil {
			return CheckResult{}, fmt.Errorf("读取当前目录失败：%w", err)
		}
		cwd = wd
	}

	cfg, cfgSrc, err := config.Load(opts.ConfigPath, cwd)
	if err != nil {
		return CheckResult{}, err
	}

	pairs, warnings, err := discovery.Discover(opts.SourceDir, cfg, discovery.DiscoverOptions{
		FailOnConflict: false,
	})
	if err != nil {
		return CheckResult{}, err
	}
	if len(pairs) == 0 {
		return CheckResult{}, fmt.Errorf("没找到可用的双语 Markdown 文件，请检查 EN/CN 目录和文件名中的数字")
	}

	warnings = dedupeStrings(warnings)
	conflictCount := countConflictWarnings(warnings)
	items := make([]CheckItem, 0, len(pairs))
	for i, p := range pairs {
		items = append(items, CheckItem{
			No:     i + 1,
			ENPath: toAbsPath(p.ENPath, cwd),
			CNPath: toAbsPath(p.CNPath, cwd),
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].No != items[j].No {
			return items[i].No < items[j].No
		}
		return items[i].ENPath < items[j].ENPath
	})
	return CheckResult{
		PairCount:     len(pairs),
		WarningCount:  len(warnings),
		ConflictCount: conflictCount,
		HasConflict:   conflictCount > 0,
		Warnings:      warnings,
		Items:         items,
		ConfigSource:  cfgSrc,
	}, nil
}

func countConflictWarnings(in []string) int {
	count := 0
	for _, s := range in {
		if strings.HasPrefix(s, "冲突组：") {
			count++
		}
	}
	return count
}

func toAbsPath(path, cwd string) string {
	if filepath.IsAbs(path) {
		return path
	}
	if cwd == "" {
		return path
	}
	return filepath.Join(cwd, path)
}

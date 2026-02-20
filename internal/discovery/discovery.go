package discovery

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"syl-md2ppt/internal/config"
)

type Pair struct {
	RelPath string
	ENPath  string
	CNPath  string
	Domain  int
	Card    int
	Side    string
}

type parsedFile struct {
	relPath string
	absPath string
	domain  int
	card    int
	side    string
}

func Discover(source string, cfg *config.Config) ([]Pair, []string, error) {
	if cfg == nil {
		return nil, nil, fmt.Errorf("配置为空，没法继续")
	}
	enRoot := filepath.Join(source, "EN")
	cnRoot := filepath.Join(source, "CN")

	re, err := regexp.Compile(cfg.Filename.Pattern)
	if err != nil {
		return nil, nil, fmt.Errorf("文件名规则有问题，请检查正则：%w", err)
	}

	enFiles, warnEN, err := scanSide(enRoot, re, cfg)
	if err != nil {
		return nil, nil, err
	}
	cnFiles, warnCN, err := scanSide(cnRoot, re, cfg)
	if err != nil {
		return nil, nil, err
	}
	warnings := append(warnEN, warnCN...)

	missing := make([]string, 0)
	for rel := range enFiles {
		if _, ok := cnFiles[rel]; !ok {
			missing = append(missing, fmt.Sprintf("中文目录缺少同名文件：%s", rel))
		}
	}
	for rel := range cnFiles {
		if _, ok := enFiles[rel]; !ok {
			missing = append(missing, fmt.Sprintf("英文目录缺少同名文件：%s", rel))
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return nil, warnings, fmt.Errorf("中英文文件没配齐：%s", strings.Join(missing, "；"))
	}

	sideRank := make(map[string]int, len(cfg.Filename.Order.Side))
	for i, side := range cfg.Filename.Order.Side {
		sideRank[side] = i
	}

	pairs := make([]Pair, 0, len(enFiles))
	for rel, enFile := range enFiles {
		cnFile := cnFiles[rel]
		pairs = append(pairs, Pair{
			RelPath: rel,
			ENPath:  enFile.absPath,
			CNPath:  cnFile.absPath,
			Domain:  enFile.domain,
			Card:    enFile.card,
			Side:    enFile.side,
		})
	}

	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].Card != pairs[j].Card {
			return pairs[i].Card < pairs[j].Card
		}
		ri, okI := sideRank[pairs[i].Side]
		rj, okJ := sideRank[pairs[j].Side]
		if okI && okJ && ri != rj {
			return ri < rj
		}
		if pairs[i].Domain != pairs[j].Domain {
			return pairs[i].Domain < pairs[j].Domain
		}
		return pairs[i].RelPath < pairs[j].RelPath
	})

	return pairs, warnings, nil
}

func scanSide(root string, re *regexp.Regexp, cfg *config.Config) (map[string]parsedFile, []string, error) {
	entries := make(map[string]parsedFile)
	warnings := make([]string, 0)

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if strings.ToLower(filepath.Ext(d.Name())) != ".md" {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		name := filepath.Base(path)
		m := re.FindStringSubmatch(name)
		if m == nil {
			if cfg.Filename.IgnoreUnmatched {
				warnings = append(warnings, fmt.Sprintf("这个文件名不符合规则，先跳过：%s", rel))
				return nil
			}
			return fmt.Errorf("文件名不符合规则：%s", rel)
		}

		domain, err := atoiGroup(m, cfg.Filename.Groups.Domain)
		if err != nil {
			return fmt.Errorf("解析域编号失败（%s）：%w", rel, err)
		}
		card, err := atoiGroup(m, cfg.Filename.Groups.Card)
		if err != nil {
			return fmt.Errorf("解析卡片编号失败（%s）：%w", rel, err)
		}
		side, err := strGroup(m, cfg.Filename.Groups.Side)
		if err != nil {
			return fmt.Errorf("解析 Front/Back 失败（%s）：%w", rel, err)
		}

		if _, exists := entries[rel]; exists {
			return fmt.Errorf("检测到重复文件：%s", rel)
		}
		entries[rel] = parsedFile{
			relPath: rel,
			absPath: path,
			domain:  domain,
			card:    card,
			side:    side,
		}
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("扫描目录失败（%s）：%w", root, err)
	}

	return entries, warnings, nil
}

func atoiGroup(m []string, idx int) (int, error) {
	s, err := strGroup(m, idx)
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func strGroup(m []string, idx int) (string, error) {
	if idx <= 0 || idx >= len(m) {
		return "", fmt.Errorf("分组索引越界：%d", idx)
	}
	return m[idx], nil
}

package discovery

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"syl-md2ppt/internal/config"
)

type Pair struct {
	RelPath  string
	ENPath   string
	CNPath   string
	Numbers  []int
	sideRank int
}

type parsedFile struct {
	relPath    string
	absPath    string
	dirRel     string
	numberKey  string
	numberList []int
	sideRank   int
	sortHint   string
}

func Discover(source string, cfg *config.Config) ([]Pair, []string, error) {
	if cfg == nil {
		return nil, nil, fmt.Errorf("配置为空，没法继续")
	}
	enRoot := filepath.Join(source, "EN")
	cnRoot := filepath.Join(source, "CN")

	enGroups, warnEN, err := scanSide(enRoot, cfg)
	if err != nil {
		return nil, nil, err
	}
	cnGroups, warnCN, err := scanSide(cnRoot, cfg)
	if err != nil {
		return nil, nil, err
	}
	warnings := append(warnEN, warnCN...)

	keys := make(map[string]struct{}, len(enGroups)+len(cnGroups))
	for k := range enGroups {
		keys[k] = struct{}{}
	}
	for k := range cnGroups {
		keys[k] = struct{}{}
	}
	keyList := make([]string, 0, len(keys))
	for k := range keys {
		keyList = append(keyList, k)
	}
	sort.Strings(keyList)

	pairs := make([]Pair, 0)
	missing := make([]string, 0)
	for _, key := range keyList {
		enGroup := enGroups[key]
		cnGroup := cnGroups[key]
		switch {
		case len(enGroup) == 0:
			missing = append(missing, fmt.Sprintf("英文目录缺少对应编号文件：%s", displayGroupKey(key)))
			continue
		case len(cnGroup) == 0:
			missing = append(missing, fmt.Sprintf("中文目录缺少对应编号文件：%s", displayGroupKey(key)))
			continue
		case len(enGroup) != len(cnGroup):
			missing = append(missing, fmt.Sprintf("中英文编号组文件数量不一致（%s）：EN=%d，CN=%d", displayGroupKey(key), len(enGroup), len(cnGroup)))
			continue
		}

		sortParsedFiles(enGroup)
		sortParsedFiles(cnGroup)
		if isConflictGroup(enGroup, cnGroup) {
			warnings = append(warnings, fmt.Sprintf(
				"冲突组：%s；同一数字键对应多个候选，请人工确认。EN=[%s]；CN=[%s]",
				displayGroupKey(key),
				joinRelPaths(enGroup),
				joinRelPaths(cnGroup),
			))
		}
		for i := range enGroup {
			pairs = append(pairs, Pair{
				RelPath:  enGroup[i].relPath,
				ENPath:   enGroup[i].absPath,
				CNPath:   cnGroup[i].absPath,
				Numbers:  enGroup[i].numberList,
				sideRank: enGroup[i].sideRank,
			})
		}
	}

	if len(missing) > 0 {
		sort.Strings(missing)
		return nil, warnings, fmt.Errorf("中英文文件没配齐：%s", strings.Join(missing, "；"))
	}

	sort.Slice(pairs, func(i, j int) bool {
		if c := compareIntSlice(pairs[i].Numbers, pairs[j].Numbers); c != 0 {
			return c < 0
		}
		if pairs[i].sideRank != pairs[j].sideRank {
			return pairs[i].sideRank < pairs[j].sideRank
		}
		return pairs[i].RelPath < pairs[j].RelPath
	})

	return pairs, warnings, nil
}

func scanSide(root string, cfg *config.Config) (map[string][]parsedFile, []string, error) {
	entries := make(map[string][]parsedFile)
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

		numberKey, numberList := numberFingerprint(filepath.Base(path))
		if numberKey == "" {
			if cfg.Filename.IgnoreUnmatched {
				warnings = append(warnings, fmt.Sprintf("这个文件名里没有可配对的唯一数字，先跳过：%s", rel))
				return nil
			}
			return fmt.Errorf("文件名里没有可配对的唯一数字：%s", rel)
		}

		dirRel := filepath.ToSlash(filepath.Dir(rel))
		if dirRel == "." {
			dirRel = ""
		}
		groupKey := makeGroupKey(dirRel, numberKey)
		entries[groupKey] = append(entries[groupKey], parsedFile{
			relPath:    rel,
			absPath:    path,
			dirRel:     dirRel,
			numberKey:  numberKey,
			numberList: numberList,
			sideRank:   detectSideRank(filepath.Base(path)),
			sortHint:   normalizeFilenameHint(filepath.Base(path)),
		})
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("扫描目录失败（%s）：%w", root, err)
	}

	return entries, warnings, nil
}

func makeGroupKey(dirRel, numKey string) string {
	return dirRel + "::" + numKey
}

func displayGroupKey(groupKey string) string {
	if i := strings.Index(groupKey, "::"); i >= 0 {
		dir := groupKey[:i]
		num := groupKey[i+2:]
		if dir == "" {
			return num
		}
		return dir + "/" + num
	}
	return groupKey
}

func numberFingerprint(name string) (string, []int) {
	base := strings.TrimSuffix(name, filepath.Ext(name))
	all := extractDigitRuns(base)
	if len(all) == 0 {
		return "", nil
	}

	count := make(map[string]int, len(all))
	for _, token := range all {
		count[token]++
	}

	unique := make([]string, 0, len(all))
	seen := make(map[string]struct{}, len(all))
	for _, token := range all {
		if count[token] != 1 {
			continue
		}
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		unique = append(unique, token)
	}
	if len(unique) == 0 {
		return "", nil
	}

	ints := make([]int, 0, len(unique))
	for _, token := range unique {
		n, err := strconv.Atoi(token)
		if err != nil {
			continue
		}
		ints = append(ints, n)
	}
	if len(ints) == 0 {
		return "", nil
	}

	return strings.Join(unique, "-"), ints
}

func extractDigitRuns(s string) []string {
	out := make([]string, 0)
	start := -1
	for i, ch := range s {
		isDigit := ch >= '0' && ch <= '9'
		if isDigit && start < 0 {
			start = i
			continue
		}
		if !isDigit && start >= 0 {
			out = append(out, s[start:i])
			start = -1
		}
	}
	if start >= 0 {
		out = append(out, s[start:])
	}
	return out
}

func normalizeFilenameHint(name string) string {
	base := strings.TrimSuffix(name, filepath.Ext(name))
	var b strings.Builder
	lastDash := false
	for _, ch := range strings.ToLower(base) {
		if ch >= '0' && ch <= '9' {
			lastDash = false
			continue
		}
		isAlpha := (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
		if isAlpha {
			b.WriteRune(ch)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func sortParsedFiles(in []parsedFile) {
	sort.Slice(in, func(i, j int) bool {
		if c := compareIntSlice(in[i].numberList, in[j].numberList); c != 0 {
			return c < 0
		}
		if in[i].sideRank != in[j].sideRank {
			return in[i].sideRank < in[j].sideRank
		}
		if in[i].sortHint != in[j].sortHint {
			return in[i].sortHint < in[j].sortHint
		}
		return in[i].relPath < in[j].relPath
	})
}

func detectSideRank(name string) int {
	base := strings.ToLower(strings.TrimSuffix(name, filepath.Ext(name)))
	tokens := splitAlphaNumTokens(base)
	hasFront := false
	hasBack := false
	for _, tok := range tokens {
		switch tok {
		case "front", "a":
			hasFront = true
		case "back", "b":
			hasBack = true
		}
	}
	if hasFront && !hasBack {
		return 0
	}
	if hasBack && !hasFront {
		return 1
	}
	return 2
}

func splitAlphaNumTokens(s string) []string {
	out := make([]string, 0)
	start := -1
	for i, ch := range s {
		isAlphaNum := (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9')
		if isAlphaNum && start < 0 {
			start = i
			continue
		}
		if !isAlphaNum && start >= 0 {
			out = append(out, s[start:i])
			start = -1
		}
	}
	if start >= 0 {
		out = append(out, s[start:])
	}
	return out
}

func compareIntSlice(a, b []int) int {
	min := len(a)
	if len(b) < min {
		min = len(b)
	}
	for i := 0; i < min; i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return 0
}

func isConflictGroup(enGroup, cnGroup []parsedFile) bool {
	if len(enGroup) <= 1 && len(cnGroup) <= 1 {
		return false
	}
	if len(enGroup) > 2 || len(cnGroup) > 2 {
		return true
	}
	ef, eb, eu := sideStats(enGroup)
	cf, cb, cu := sideStats(cnGroup)

	if ef > 1 || eb > 1 || cf > 1 || cb > 1 {
		return true
	}
	if (eu > 0 && len(enGroup) > 1) || (cu > 0 && len(cnGroup) > 1) {
		return true
	}
	if ef != cf || eb != cb {
		return true
	}
	return false
}

func sideStats(group []parsedFile) (front int, back int, unknown int) {
	for _, f := range group {
		switch f.sideRank {
		case 0:
			front++
		case 1:
			back++
		default:
			unknown++
		}
	}
	return front, back, unknown
}

func joinRelPaths(group []parsedFile) string {
	if len(group) == 0 {
		return ""
	}
	parts := make([]string, 0, len(group))
	for _, f := range group {
		parts = append(parts, f.relPath)
	}
	return strings.Join(parts, ", ")
}

package output

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
)

const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func ResolveOutputPath(outputArg, cwd string, now time.Time, rand io.Reader) (string, error) {
	if strings.TrimSpace(cwd) == "" {
		return "", fmt.Errorf("当前目录为空，没法确定输出位置")
	}

	if strings.TrimSpace(outputArg) == "" {
		name, err := defaultName(now, rand)
		if err != nil {
			return "", err
		}
		return filepath.Join(cwd, name), nil
	}

	if strings.HasSuffix(strings.ToLower(outputArg), ".pptx") {
		if filepath.IsAbs(outputArg) {
			return outputArg, nil
		}
		return filepath.Join(cwd, outputArg), nil
	}

	name, err := defaultName(now, rand)
	if err != nil {
		return "", err
	}

	if filepath.IsAbs(outputArg) {
		return filepath.Join(outputArg, name), nil
	}
	return filepath.Join(cwd, outputArg, name), nil
}

func defaultName(now time.Time, rand io.Reader) (string, error) {
	sfx, err := randomSuffix(rand, 6)
	if err != nil {
		return "", err
	}
	return now.Format("20060102_150405") + "_" + sfx + ".pptx", nil
}

func randomSuffix(rand io.Reader, n int) (string, error) {
	if rand == nil {
		return "", fmt.Errorf("随机数源不可用，暂时没法生成默认文件名")
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(rand, buf); err != nil {
		return "", fmt.Errorf("读取随机数失败：%w", err)
	}
	out := make([]byte, n)
	for i := range buf {
		c := buf[i]
		if c >= 'a' && c <= 'z' {
			c = c - 'a' + 'A'
		}
		if strings.IndexByte(alphabet, c) >= 0 {
			out[i] = c
			continue
		}
		out[i] = alphabet[int(c)%len(alphabet)]
	}
	return string(out), nil
}

package cmd

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var unknownCmdPattern = regexp.MustCompile(`unknown command "([^"]+)" for "([^"]+)"`)
var errAlreadyPrinted = errors.New("错误信息已输出")

func FriendlyError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, errAlreadyPrinted) {
		return ""
	}
	msg := strings.TrimSpace(err.Error())
	if msg == "" {
		return "出了点问题，但错误信息是空的"
	}

	if m := unknownCmdPattern.FindStringSubmatch(msg); len(m) == 3 {
		return fmt.Sprintf("这个命令我不认识：%s。可以先输入 `%s --help` 看看怎么用。", m[1], m[2])
	}
	if strings.HasPrefix(msg, "unknown flag: ") {
		flagName := strings.TrimPrefix(msg, "unknown flag: ")
		return fmt.Sprintf("这个参数我不认识：%s。可以先输入 `syl-md2ppt --help` 看看可用参数。", flagName)
	}
	if strings.Contains(msg, "required flag(s)") {
		return "有参数没填完整，先输入 `syl-md2ppt --help` 看下示例。"
	}
	if strings.Contains(msg, "accepts") && strings.Contains(msg, "arg(s)") {
		return "参数数量不对，输入 `syl-md2ppt --help` 看看正确写法。"
	}
	return msg
}

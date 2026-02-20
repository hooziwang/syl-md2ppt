package cmd

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"syl-md2ppt/internal/app"
)

type buildFlags struct {
	outputArg string
	configArg string
}

const dataSourceRequirementsHelp = `
数据源要求：
1. 数据源目录必须包含 EN/ 和 CN/ 两个子目录。
2. 程序会递归扫描 EN/、CN/ 下所有 .md 文件。
3. 配对时提取文件名中的数字，只使用“非重复数字”作为配对键。
4. 同一相对目录下，EN/CN 的数字键一致才会配成一对。
5. 正反面识别：Front/A 为正面，Back/B 为反面（正面在前）。

目录结构示意：
SPI/
├─ EN/
│  ├─ 00_Intro/0-001-Front.md
│  └─ 00_Intro/0-001-Back.md
└─ CN/
   ├─ 00_Intro/0-001-Front.md
   └─ 00_Intro/0-001-Back.md
`

func Execute() error {
	root := NewRootCmd(time.Now, rand.Reader, os.Stdout, os.Stderr)
	root.SetArgs(normalizeArgs(os.Args[1:]))
	return root.Execute()
}

func NewRootCmd(nowFn func() time.Time, randSrc io.Reader, stdout io.Writer, stderr io.Writer) *cobra.Command {
	flags := &buildFlags{}
	showVersion := false

	root := &cobra.Command{
		Use:           "syl-md2ppt [data_source_dir]",
		Short:         "把双语 Markdown 一键转成一个 PPTX",
		Long:          "把双语 Markdown 一键转成一个 PPTX。\n" + dataSourceRequirementsHelp,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          runBuild(nowFn, randSrc, stdout, stderr, flags, false, &showVersion),
	}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.CompletionOptions.HiddenDefaultCmd = true
	bindBuildFlags(root, flags)
	root.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "显示版本信息")

	buildCmd := &cobra.Command{
		Use:           "build <data_source_dir>",
		Short:         "从 EN/CN Markdown 生成 PPTX",
		Long:          "从 EN/CN Markdown 生成 PPTX。\n" + dataSourceRequirementsHelp,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          runBuild(nowFn, randSrc, stdout, stderr, flags, true, &showVersion),
	}
	root.AddCommand(buildCmd)

	checkCmd := &cobra.Command{
		Use:           "check <data_source_dir>",
		Short:         "检查 EN/CN 数据源是否可生成 PPTX",
		Long:          "检查 EN/CN 数据源是否可生成 PPTX。\n" + dataSourceRequirementsHelp,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          runCheck(stdout, stderr, flags, &showVersion),
	}
	root.AddCommand(checkCmd)

	versionCmd := &cobra.Command{
		Use:           "version",
		Short:         "显示版本信息",
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			printVersion(stdout)
		},
	}
	root.AddCommand(versionCmd)
	return root
}

func bindBuildFlags(cmd *cobra.Command, flags *buildFlags) {
	cmd.PersistentFlags().StringVar(&flags.outputArg, "output", "", "输出文件路径或输出目录")
	cmd.PersistentFlags().StringVar(&flags.configArg, "config", "", "YAML 配置文件路径")
}

func runBuild(nowFn func() time.Time, randSrc io.Reader, stdout io.Writer, stderr io.Writer, flags *buildFlags, subcommand bool, showVersion *bool) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if showVersion != nil && *showVersion {
			printVersion(stdout)
			return nil
		}
		if len(args) == 0 {
			if !subcommand {
				fmt.Fprintln(stderr, "还没给数据源目录。用法：syl-md2ppt <数据源目录>")
				fmt.Fprintln(stderr)
				prevOut := cmd.OutOrStdout()
				cmd.SetOut(stderr)
				_ = cmd.Help()
				cmd.SetOut(prevOut)
				return errAlreadyPrinted
			}
			_ = cmd.Help()
			return fmt.Errorf("还没给数据源目录。用法：syl-md2ppt <数据源目录>")
		}
		if len(args) > 1 {
			return fmt.Errorf("参数有点多了，只需要一个数据源目录")
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("读取当前目录失败：%w", err)
		}

		res, err := app.Run(app.Options{
			SourceDir:  args[0],
			OutputArg:  flags.outputArg,
			ConfigPath: flags.configArg,
			CWD:        cwd,
			Now:        nowFn(),
			Rand:       randSrc,
		})
		if err != nil {
			return err
		}

		for _, w := range res.Warnings {
			fmt.Fprintln(stderr, w)
		}
		fmt.Fprintf(stdout, "搞定啦，PPT 已生成：%s\n", res.OutputPath)
		_ = subcommand
		return nil
	}
}

func runCheck(stdout io.Writer, stderr io.Writer, flags *buildFlags, showVersion *bool) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if showVersion != nil && *showVersion {
			printVersion(stdout)
			return nil
		}
		if len(args) == 0 {
			_ = cmd.Help()
			return fmt.Errorf("还没给数据源目录。用法：syl-md2ppt check <数据源目录>")
		}
		if len(args) > 1 {
			return fmt.Errorf("参数有点多了，只需要一个数据源目录")
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("读取当前目录失败：%w", err)
		}
		res, err := app.Check(app.Options{
			SourceDir:  args[0],
			ConfigPath: flags.configArg,
			CWD:        cwd,
		})
		if err != nil {
			return err
		}
		for _, w := range res.Warnings {
			fmt.Fprintln(stderr, w)
		}
		for _, it := range res.Items {
			fmt.Fprintf(stdout, "[%03d] - %s\n", it.No, it.ENPath)
			fmt.Fprintf(stdout, "[%03d] - %s\n", it.No, it.CNPath)
		}
		fmt.Fprintf(stdout, "检查通过：共识别 %d 对双语文件，可生成 %d 页 PPT\n", res.PairCount, res.PairCount)
		return nil
	}
}

func normalizeArgs(args []string) []string {
	if len(args) == 0 {
		return args
	}
	first := args[0]
	switch first {
	case "build", "check", "help", "completion", "version":
		return args
	}
	if first == "-h" || first == "--help" || first == "-v" || first == "--version" {
		return args
	}
	if !containsPositionalSource(args) {
		return args
	}
	return append([]string{"build"}, args...)
}

func containsPositionalSource(args []string) bool {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			return i+1 < len(args)
		}
		if arg == "--output" || arg == "--config" {
			i++
			continue
		}
		if strings.HasPrefix(arg, "--output=") || strings.HasPrefix(arg, "--config=") {
			continue
		}
		if strings.HasPrefix(arg, "-") {
			continue
		}
		return true
	}
	return false
}

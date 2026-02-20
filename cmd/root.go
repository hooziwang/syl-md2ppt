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
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          runBuild(nowFn, randSrc, stdout, stderr, flags, false, &showVersion),
	}
	root.SetOut(stdout)
	root.SetErr(stderr)
	bindBuildFlags(root, flags)
	root.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "显示版本信息")

	buildCmd := &cobra.Command{
		Use:           "build <data_source_dir>",
		Short:         "从 EN/CN Markdown 生成 PPTX",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          runBuild(nowFn, randSrc, stdout, stderr, flags, true, &showVersion),
	}
	root.AddCommand(buildCmd)

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

func normalizeArgs(args []string) []string {
	if len(args) == 0 {
		return args
	}
	first := args[0]
	switch first {
	case "build", "help", "completion", "version":
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

# syl-md2ppt Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 构建一个 Go+Cobra 命令行工具，读取 `EN/CN` 双语 Markdown，按可配置模板生成单个 PPTX。

**Architecture:** 采用分层结构：`cmd` 负责 CLI，`internal/config` 管理 YAML，`internal/discovery` 负责扫描/解析/排序/配对，`internal/render` 负责 Markdown 样式解析与版式计算，`internal/pptx` 负责 OpenXML 写入。输出风格保持极简，错误即退出，告警可统计。

**Tech Stack:** Go 1.22+, Cobra, yaml.v3, stretchr/testify（测试）

---

### Task 1: 初始化项目骨架与 CLI 入口

**Files:**
- Create: `go.mod`
- Create: `main.go`
- Create: `cmd/root.go`
- Create: `cmd/build.go`

**Step 1: 写失败测试（CLI 参数缺失）**

```go
func TestBuildRequiresSourceArg(t *testing.T) {
    code, out := runCLI([]string{"build"})
    require.NotEqual(t, 0, code)
    require.Contains(t, out, "missing data_source_dir")
}
```

**Step 2: 运行测试确认失败**

Run: `go test ./...`
Expected: FAIL，提示命令或测试桩缺失。

**Step 3: 实现最小 CLI 结构**

```go
var rootCmd = &cobra.Command{Use: "syl-md2ppt"}
var buildCmd = &cobra.Command{Use: "build <data_source_dir>", Args: cobra.ExactArgs(1)}
```

**Step 4: 再跑测试确认通过/进入下一个失败点**

Run: `go test ./...`
Expected: 对应测试通过，其他未实现模块继续失败。

**Step 5: Commit**

```bash
git add go.mod main.go cmd/root.go cmd/build.go
git commit -m "feat(cli): 初始化命令入口与 build 子命令"
```

### Task 2: 输出路径判定与默认文件名

**Files:**
- Create: `internal/output/path.go`
- Create: `internal/output/path_test.go`
- Modify: `cmd/build.go`

**Step 1: 写失败测试（目录/文件/默认输出）**

```go
func TestResolveOutputPath_DefaultFileName(t *testing.T) { /* ... */ }
func TestResolveOutputPath_OutputAsDir(t *testing.T) { /* ... */ }
func TestResolveOutputPath_OutputAsFile(t *testing.T) { /* ... */ }
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/output -v`
Expected: FAIL，函数未实现。

**Step 3: 实现最小逻辑**

```go
func ResolveOutputPath(outputArg string, now time.Time, rand io.Reader) (string, error) { /* ... */ }
```

**Step 4: 运行测试确认通过**

Run: `go test ./internal/output -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/output/path.go internal/output/path_test.go cmd/build.go
git commit -m "feat(output): 实现输出路径与默认文件名解析"
```

### Task 3: YAML 配置加载与默认值

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/default.yaml`
- Create: `internal/config/load.go`
- Create: `internal/config/load_test.go`
- Modify: `cmd/build.go`

**Step 1: 写失败测试（优先级与默认回退）**

```go
func TestLoadConfig_Priority(t *testing.T) { /* --config > ./syl-md2ppt.yaml > embedded */ }
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/config -v`
Expected: FAIL

**Step 3: 实现配置读取**

```go
func Load(pathArg string, cwd string) (*Config, error) { /* ... */ }
```

**Step 4: 测试通过**

Run: `go test ./internal/config -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/config/* cmd/build.go
git commit -m "feat(config): 支持配置优先级与内置默认模板"
```

### Task 4: 文件名解析、排序、配对校验

**Files:**
- Create: `internal/discovery/model.go`
- Create: `internal/discovery/scan.go`
- Create: `internal/discovery/sort.go`
- Create: `internal/discovery/pair.go`
- Create: `internal/discovery/discovery_test.go`

**Step 1: 写失败测试（规则解析、Front/Back 顺序、配对失败）**

```go
func TestParseAndSort(t *testing.T) { /* 按 card_no + side 顺序 */ }
func TestPairingMustMatch(t *testing.T) { /* 缺 CN 时失败 */ }
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/discovery -v`
Expected: FAIL

**Step 3: 实现最小扫描与配对逻辑**

```go
func Discover(source string, cfg *config.Config) ([]Pair, []string, error) { /* ... */ }
```

**Step 4: 测试通过**

Run: `go test ./internal/discovery -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/discovery/*
git commit -m "feat(discovery): 实现文件发现、智能排序与配对校验"
```

### Task 5: Markdown 行内样式解析

**Files:**
- Create: `internal/render/token.go`
- Create: `internal/render/parse.go`
- Create: `internal/render/parse_test.go`

**Step 1: 写失败测试（粗体、斜体、$公式$、标识行）**

```go
func TestParseInlineStyles(t *testing.T) { /* ... */ }
func TestParseMarkerLine(t *testing.T) { /* ★ ● ▲ */ }
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/render -v`
Expected: FAIL

**Step 3: 实现最小解析器**

```go
func ParseMarkdown(raw string, cfg *config.Config) []Block { /* ... */ }
```

**Step 4: 测试通过**

Run: `go test ./internal/render -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/render/*
git commit -m "feat(render): 实现 markdown 样式与标识解析"
```

### Task 6: PPTX(OpenXML) 写入器

**Files:**
- Create: `internal/pptx/doc.go`
- Create: `internal/pptx/types.go`
- Create: `internal/pptx/write.go`
- Create: `internal/pptx/write_test.go`

**Step 1: 写失败测试（生成可打开的最小 pptx + 页数匹配）**

```go
func TestWritePPTX_MinimalPackage(t *testing.T) { /* zip entries exist */ }
func TestWritePPTX_SlideCount(t *testing.T) { /* slide count == pairs */ }
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/pptx -v`
Expected: FAIL

**Step 3: 实现最小 OpenXML 生成**

```go
func Write(out string, deck Deck) error { /* zip + xml parts */ }
```

**Step 4: 测试通过**

Run: `go test ./internal/pptx -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/pptx/*
git commit -m "feat(pptx): 实现 openxml 写入与幻灯片生成"
```

### Task 7: 排版引擎（双栏 + 缩放/截断）

**Files:**
- Create: `internal/render/layout.go`
- Create: `internal/render/fit.go`
- Create: `internal/render/layout_test.go`
- Modify: `internal/pptx/types.go`

**Step 1: 写失败测试（双栏布局、最小字号、截断告警）**

```go
func TestLayoutTwoColumns(t *testing.T) { /* EN left CN right */ }
func TestFitAndTruncate(t *testing.T) { /* shrink then truncate */ }
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/render -v`
Expected: FAIL

**Step 3: 实现布局与拟合**

```go
func BuildSlide(en, cn string, cfg *config.Config) (Slide, []Warning) { /* ... */ }
```

**Step 4: 测试通过**

Run: `go test ./internal/render -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/render/* internal/pptx/types.go
git commit -m "feat(layout): 实现双栏排版与溢出处理"
```

### Task 8: 端到端串联与简洁输出

**Files:**
- Create: `internal/app/run.go`
- Create: `internal/app/run_test.go`
- Modify: `cmd/build.go`

**Step 1: 写失败测试（小样本目录生成 ppt）**

```go
func TestRun_EndToEnd(t *testing.T) { /* 2 files -> 2 slides */ }
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/app -v`
Expected: FAIL

**Step 3: 实现流程编排**

```go
func Run(opts Options) error {
  // load config -> discover pairs -> render -> write pptx
}
```

**Step 4: 测试通过并全量回归**

Run: `go test ./...`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/app/* cmd/build.go
git commit -m "feat(app): 串联全流程并输出最终 pptx"
```

### Task 9: 文档与示例配置

**Files:**
- Create: `README.md`
- Create: `examples/syl-md2ppt.yaml`

**Step 1: 写文档测试清单（命令示例可执行）**

```bash
go run . ./testdata/SPI --output ./out
```

**Step 2: 补充 README（最小必要信息）**
- 安装/运行
- 参数说明
- 配置覆盖说明
- 错误码说明

**Step 3: 验证示例命令**

Run: `go test ./... && go run . ./testdata/SPI --output ./tmp`
Expected: PASS + 生成 pptx

**Step 4: Commit**

```bash
git add README.md examples/syl-md2ppt.yaml
git commit -m "docs: 补充使用说明与配置示例"
```


# syl-md2ppt

将双语 Markdown 数据源（`EN/` + `CN/`）转换为单个 `pptx`。

## 特性

- Go + Cobra 命令行工具
- 一个 `md` 文件对应一页 PPT
- 左栏英文，右栏中文
- 支持 `**粗体**`、`*斜体*`、`★/●/▲` 标识、`$...$` 公式亮色高亮
- 模板化 YAML 配置（布局、字体、颜色、文件名解析规则）

## 安装

### 方式 1：Homebrew（macOS，推荐）

1. 检查是否已安装 Homebrew：

```bash
brew --version
```

如果提示 `command not found`，先安装 Homebrew：

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

Apple Silicon（M1/M2/M3）如果装完还找不到 `brew`，执行：

```bash
echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zprofile && eval "$(/opt/homebrew/bin/brew shellenv)"
```

2. 安装 `syl-md2ppt`：

```bash
brew tap hooziwang/tap && brew install syl-md2ppt
```

3. 验证安装：

```bash
syl-md2ppt -v
```

4. 后续升级：

```bash
brew update && brew upgrade syl-md2ppt
```

5. 卸载：

```bash
brew uninstall syl-md2ppt
```

### 方式 2：Scoop（Windows，推荐）

1. 以普通用户打开 PowerShell，先检查是否已安装 Scoop：

```powershell
scoop --version
```

如果提示找不到命令，先安装 Scoop：

```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser -Force; irm get.scoop.sh | iex
```

2. 安装 `syl-md2ppt`：

```powershell
scoop bucket add hooziwang https://github.com/hooziwang/scoop-bucket; scoop install syl-md2ppt
```

3. 验证安装：

```powershell
syl-md2ppt -v
```

4. 后续升级：

```powershell
scoop update; scoop update syl-md2ppt
```

5. 卸载：

```powershell
scoop uninstall syl-md2ppt
```

## 运行

### 入口 1（直跑）

```bash
./syl-md2ppt <data_source_dir> [--output ...] [--config ...]
```

### 入口 2（子命令）

```bash
syl-md2ppt build <data_source_dir> [--output ...] [--config ...]
```

### 入口 3（仅检查，不生成文件）

```bash
syl-md2ppt check <data_source_dir> [--config ...]
```

## 数据源要求

推荐的数据源目录结构示意：

```text
SPI/
├─ EN/
│  ├─ 00_Intro/
│  │  ├─ 0-001-Front.md
│  │  └─ 0-001-Back.md
│  └─ 01_Domain1/
│     ├─ 1-002-Front.md
│     └─ 1-002-Back.md
└─ CN/
   ├─ 00_Intro/
   │  ├─ 0-001-Front.md
   │  └─ 0-001-Back.md
   └─ 01_Domain1/
      ├─ 1-002-Front.md
      └─ 1-002-Back.md
```

要求：

1. 数据源目录必须包含 `EN/` 和 `CN/` 两个子目录。
2. 程序会递归扫描 `EN/`、`CN/` 下所有 `.md` 文件。
3. 建议 `EN/` 与 `CN/` 使用相同的子目录层级，便于一一配对。
4. 配对时不依赖固定命名模板，只提取文件名中的数字，并仅使用“非重复数字”作为配对键。
5. 同一相对目录下，`EN` 与 `CN` 的非重复数字键一致，才会配成一对（`001` 与 `1` 视为同一个数字）。
6. 同一卡片内会识别正反面并排序：`Front`/`A` 视为正面，`Back`/`B` 视为反面（正面在前、反面在后）。
7. 如果某个编号组在 `EN`/`CN` 数量不一致，会报错退出；如果同一编号组存在多个候选，会提示“冲突组”并建议人工确认。
8. 文件名中没有可用非重复数字时，按 `filename.ignore_unmatched` 决定跳过或报错。

## 参数

- `<data_source_dir>`：必填位置参数。目录内必须包含 `EN/` 和 `CN/`。
- `--output`：可选。
  - 以 `.pptx` 结尾 -> 视为输出文件
  - 否则 -> 视为输出目录
  - 未提供 -> 当前目录自动生成：`yyyyMMdd_HHmmss_<6位随机码>.pptx`
- `--config`：可选。
  - 优先级：`--config` > 当前目录 `syl-md2ppt.yaml` > 内置默认模板

## 文件名智能配对规则

程序会从文件名里提取数字，并只使用“非重复数字”做配对键：

- 不依赖固定命名模板，不要求 `Front/Back` 这种固定格式。
- 同一文件名里重复出现的数字会被忽略，只保留非重复数字。
- EN 和 CN 在同一相对目录下，非重复数字键一致，就会被视为一对。
- 如果某个文件名里没有可用的非重复数字，会按配置决定跳过或报错。

## 退出行为

- 配对失败（EN/CN 缺文件）、参数错误、目录结构错误 -> 非 0 退出
- 非匹配文件、内容截断 -> `warn:` 单行告警，不中断生成

## 示例

```bash
syl-md2ppt ~/Downloads/SPI --output /tmp/syl.pptx
```

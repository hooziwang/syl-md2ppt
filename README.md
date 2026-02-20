# syl-md2ppt

将双语 Markdown 数据源（`EN/` + `CN/`）转换为单个 `pptx`。

## 特性

- Go + Cobra 命令行工具
- 一个 `md` 文件对应一页 PPT
- 左栏英文，右栏中文
- 支持 `**粗体**`、`*斜体*`、`★/●/▲` 标识、`$...$` 公式亮色高亮
- 模板化 YAML 配置（布局、字体、颜色、文件名解析规则）

## 安装与运行

```bash
go build -o syl-md2ppt .
```

### 入口 1（直跑）

```bash
./syl-md2ppt <data_source_dir> [--output ...] [--config ...]
```

### 入口 2（子命令）

```bash
./syl-md2ppt build <data_source_dir> [--output ...] [--config ...]
```

## 参数

- `<data_source_dir>`：必填位置参数。目录内必须包含 `EN/` 和 `CN/`。
- `--output`：可选。
  - 以 `.pptx` 结尾 -> 视为输出文件
  - 否则 -> 视为输出目录
  - 未提供 -> 当前目录自动生成：`yyyyMMdd_HHmmss_<6位随机码>.pptx`
- `--config`：可选。
  - 优先级：`--config` > 当前目录 `syl-md2ppt.yaml` > 内置默认模板

## 文件名规则（可配置）

默认正则：

```regex
^(\d+)-(\d{3})-(Front|Back)\.md$
```

默认排序：
1. `card_no` 升序
2. 同号内 `Front` 在前，`Back` 在后

## 退出行为

- 配对失败（EN/CN 缺文件）、参数错误、目录结构错误 -> 非 0 退出
- 非匹配文件、内容截断 -> `warn:` 单行告警，不中断生成

## 示例

```bash
./syl-md2ppt /Users/wxy/Downloads/SPI --output /tmp/syl.pptx
```

## 说明

`/Users/wxy/Downloads/spi 2` 的混合内容拆分与中文补齐是前置数据整理，不属于本程序功能。

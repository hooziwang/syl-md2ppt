# syl-md2ppt 设计文档

**日期**: 2026-02-20  
**状态**: 已确认

## 1. 目标与边界

`syl-md2ppt` 是一个 Go 命令行程序，将指定数据源目录中的双语 Markdown 文件转换为单个 PPTX。

边界约束：
- 程序只处理已经整理好的目录结构：`<data_source_dir>/EN` 与 `<data_source_dir>/CN`。
- 混合目录拆分和中文补齐属于编码前一次性前置工作，不是程序功能。
- 一个 Markdown 文件对应一页 PPT。
- 左栏英文、右栏中文。

## 2. CLI 设计

支持两种等价入口：
1. `syl-md2ppt <data_source_dir> [--output ...] [--config ...]`
2. `syl-md2ppt build <data_source_dir> [--output ...] [--config ...]`

参数规则：
- `data_source_dir`：第一个位置参数，必填。
- `--output`：可选。
  - 以 `.pptx` 结尾：视为输出文件路径。
  - 否则：视为输出目录，文件名自动生成。
- 未传 `--output`：输出到当前目录，文件名 `yyyyMMdd_HHmmss_<6位随机码>.pptx`。
- `--config`：可选。
  - 优先级：`--config` > 当前目录 `syl-md2ppt.yaml` > 内置默认模板。

## 3. 数据发现与排序

### 3.1 文件发现
- 递归扫描 `EN/` 和 `CN/` 下所有 `*.md`。

### 3.2 文件名解析规则
- 从 YAML 读取正则表达式，默认：`^(\d+)-(\d{3})-(Front|Back)\.md$`。
- 解析得到：域编号、卡片编号、正反面标识。

### 3.3 智能排序
- 排序键：
  1. `card_no` 升序。
  2. 同一 `card_no` 内按 `Front` 再 `Back`（顺序可配）。
- 不匹配规则文件默认忽略并输出简短告警。

### 3.4 配对校验
- `EN` 与 `CN` 必须同名一一对应。
- 发现不配对时直接失败退出（非 0）。

## 4. 页面排版与样式

### 4.1 布局
- 一文件一页。
- 左右双栏：左 EN、右 CN。
- 不单独抽标题，按 Markdown 原文内容排版。

### 4.2 标识样式映射
- `**...**`：加粗。
- `*...*`：斜体。
- `★` 行：强调样式。
- `●` 行：普通要点样式。
- `▲` 行：警示高亮样式。
- `$...$`：亮色高亮显示。

### 4.3 超长处理
- 先逐步缩小字体，直到最小阈值。
- 仍溢出则截断，并记录告警。

## 5. 模板化配置（YAML）

采用 `YAML + 内置默认版式`。

示意结构：

```yaml
filename:
  pattern: '^(\d+)-(\d{3})-(Front|Back)\\.md$'
  order:
    side: ["Front", "Back"]
  ignore_unmatched: true

layout:
  slide:
    width: 13.333
    height: 7.5
    unit: in
  columns:
    left_ratio: 0.5
    gap: 0.2
    padding: 0.3
  typography:
    font_family: "Calibri"
    base_size: 20
    min_size: 12
    line_spacing: 1.2

styles:
  markers:
    star: { prefix: "★", accent_bar: true, color: "#8A6D1D" }
    dot:  { prefix: "●", color: "#1F2937" }
    warn: { prefix: "▲", highlight: "#FFE8B3", color: "#9A3412" }
  inline_formula:
    delimiter: "$"
    highlight: "#FFF176"
    color: "#111827"

output:
  default_name:
    timestamp_format: "20060102_150405"
    random_suffix_len: 6
```

## 6. 错误模型与输出风格

- 参数错误、输入目录非法、缺 `EN/CN`、配对失败：非 0 退出。
- 非匹配文件、截断：告警并统计。
- 输出尽量简洁，单行、可 grep。

## 7. 技术路线

采用方案 A：纯 Go 生成 PPTX（OpenXML），使用开源依赖，避免商业组件。

## 8. 前置数据处理约定

编码前执行一次性整理：
- 来源目录：`/Users/wxy/Downloads/spi 2`（不可更改）。
- 目标目录：`~/Downloads/SPI/EN` 与 `~/Downloads/SPI/CN`。
- 存在同名文件时覆盖。
- 若中文缺失，由本次执行阶段直接补齐。


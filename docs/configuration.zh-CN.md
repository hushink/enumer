# 配置指南

[English](configuration.md) | **简体中文**

本指南介绍 `enumer` 的常见配置方式。完整字段参考见
[清单 schema](manifest-schema.zh-CN.md)。

## 最小清单

```yaml
version: 1
package: example

enums:
  - name: Color
    type: string
    values:
      - red
      - green
      - blue
```

`enumer generate` 会在清单旁生成 `color_enum.go`。生成文件包含 `Color` 类型、
`ColorRed`/`ColorGreen`/`ColorBlue`，以及 String、Parse、IsValid、文本序列化、
names 与 values。

如果目录内已有 Go 文件，`package` 可以省略，enumer 会从 package 子句推断；
不确定时建议显式设置。

## 清单放置与发现

- 任意目录都可以放置 `enumer.yaml`。
- `enumer generate ./...` 递归发现 `enumer.yaml` 和 `enumer.yml`，跳过
  `vendor/`、`testdata/`、隐藏目录和 `_` 前缀目录。
- 显式传入文件路径时，文件名不要求是 `enumer.yaml`。

## 枚举与取值

```yaml
enums:
  - name: Status
    type: int32
    start: 100
    values:
      - pending
      - name: running
        value: 110
        comment: Running is the active state.
      - name: failed
        aliases: [error]
        deprecated: use failed_permanently
```

- `type` 可以是 `string`、有符号整数或无符号整数。
- 字符串枚举的 `value` 是规范文本，默认等于取值名。
- 数字枚举的 `value` 是数字。未提供时从 `start`（默认 `0`）自动递增；
  显式值会重置后续序列。
- `aliases` 是额外可解析文本，不生成常量。
- `deprecated` 添加 Go `Deprecated:` 注释。
- `comment` 覆盖生成注释，优先于 `doc`。

取值名会转换为 Go 标识符：

```yaml
values:
  - http_status   # ColorHTTPStatus
  - blue-green    # ColorBlueGreen
  - red_orange    # ColorRedOrange
```

## 能力

默认能力集：

```yaml
defaults:
  features: [string, parse, valid, text, names, values]
```

按枚举增量添加：

```yaml
defaults:
  features: [string, parse, valid, text, names, values]

enums:
  - name: Color
    features: [json, sql]        # 在默认集合上追加
    disableFeatures: [names]     # 从默认集合移除
```

可用能力：`string`、`parse`、`valid`、`flags`、`mustparse`、`text`、`json`、
`json:number`、`yaml`、`sql`、`sql:int`、`sql:null`、`flag`、`names`、`values`、
`ptr`。依赖会自动补齐。完整表格见
[清单 schema](manifest-schema.zh-CN.md#能力)。

显式空 `features: []` 会关闭全部能力：

```yaml
defaults:
  features: []
```

## 输出布局

每枚举输出默认是清单目录下的 `<snake(name)>_enum.go`：

```yaml
defaults:
  output:
    suffix: _enum.go
    dir: generated

enums:
  - name: Color
    output: custom/color.go  # 精确路径，相对清单目录
```

聚合输出把清单内所有枚举合并到一个文件：

```yaml
aggregate: true
output: enums_enum.go
```

聚合清单要求所有枚举使用同一个包，且禁止枚举级 `output`。

## 移除枚举后的清理

从清单移除枚举后，下一次 `enumer generate` 会删除对应的旧 `.go` 文件。只有带有
当前清单精确生成头的文件才会被清理；手写文件、其它清单生成的文件和自定义目标
输出不会被触碰。`enumer check` 会把孤儿生成文件报告为过期；`--no-prune` 可关闭
清理。

## 注释

```yaml
defaults:
  comments: auto              # auto | none

enums:
  - name: Color
    comment: Color is the public color enumeration.
    comments: none            # 枚举级覆盖
    values:
      - name: red
        comment: Red is the default color.
```

- `auto` 在没有用户注释时生成注释。
- `none` 关闭自动注释；用户注释和弃用说明仍保留。
- `doc` 与 `comment` 都是注释覆盖，`comment` 优先。

## 类型声明

`typeMode` 控制生成文件是否声明类型：

| 模式 | 行为 |
| --- | --- |
| `auto`（默认） | 类型不存在时生成；已存在时使用现有声明；别名、泛型或底层类型不一致会报错。 |
| `emit` | 总是生成；其它文件已有声明时报错。 |
| `existing` | 从不生成；声明必须存在且匹配。 |

```yaml
# types.go
package example

// Color is a hand-written type.
type Color string
```

默认 `auto` 下，下面的清单只生成常量与方法，因为 `Color` 已存在：

```yaml
enums:
  - name: Color
    type: string
    values: [red, green, blue]
```

检测只使用 `go/parser` 对输出包内非测试 `.go` 文件做语法扫描，不加载或类型检查
整个包。

## 通过 extends 共享默认配置

```yaml
# base.yaml
version: 1
defaults:
  features: [string, parse, valid, text, names, values]
  naming:
    prefix: type

# color/enumer.yaml
version: 1
package: color
extends: ../base.yaml

defaults:
  naming:
    identifier: camel
enums:
  - name: Color
    values: [red, green, blue]
```

只继承 `defaults`。嵌套的 `naming`、`parse`、`output` 逐字段合并；列表整体替换。
循环继承会报错。基础文件只有在命名为 `enumer.yaml` 时才会被 `./...` 发现。

## 自定义目标

```yaml
targets:
  - template: ./templates/enums.md.tmpl
    output: ENUMS.md
  - template: ./templates/schema.go.tmpl
    output: schema_generated.go
```

模板数据：

- `.Manifest`
- `.Enums`
- `.Enum`（`perEnum: true` 时）

模板函数：`camel`、`snake`、`kebab`、`lower`、`upper`、`title`、`quote`、
`join`、`json`、`jsonPretty`、`default`。

输出路径以 `.go` 结尾的目标会使用 `go/format.Source` 格式化，因此必须渲染完整
Go 文件。

## 位标志

```yaml
enums:
  - name: Permission
    type: uint32
    features: [flags]
    values:
      - name: none
        value: 0
      - name: read
        value: 1
      - name: write
        value: 2
```

`flags` 要求无符号整数类型。取值必须是 0 或 2 的幂、不能重复，最多一个零值。
`String`、`Parse` 与 `IsValid` 均支持组合。

## 诊断

```sh
enumer generate ./...             # 文本诊断，出错时退出码 1
enumer --json generate ./...      # JSON 诊断
enumer --strict generate ./...    # 警告升级为错误
enumer check --diff ./...         # 生成物过期时失败
```

退出码：`0` 成功，`1` 校验/生成/过期，`2` 用法错误。

## 格式化与幂等

- 内置 Go 输出使用 `go/format.Source`。
- `.go` 自定义目标同样格式化。
- 写入是原子的；内容完全一致的文件不会重写。
- `enumer check` 比较内存渲染结果与磁盘内容。

## 合规

本项目使用 MIT 许可证。CI 会检查 SPDX 头、依赖许可证和已知漏洞。具体命令见
仓库 README。

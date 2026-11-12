# 清单 schema

[English](manifest-schema.md) | **简体中文**

一个 `enumer.yaml` 文件声明一个或多个枚举，并可选声明附加的模板目标。清单是
唯一真相源；`enumer` 不解析业务 Go 源码。

清单采用严格解码。未知字段、不支持的版本和非法取值都会带文件位置与稳定错误码
报告。

## 顶层字段

| 字段 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `version` | integer | `1` | 清单 schema 版本。目前只支持 `1`。 |
| `extends` | string | `""` | 基础清单路径，继承其 `defaults`。只继承 `defaults`。 |
| `package` | string | 推断 | Go 包名。省略时从目录内 Go 文件推断；目录内没有 Go 文件时必须显式设置。 |
| `aggregate` | bool | `false` | 把本清单所有枚举输出到一个文件。 |
| `output` | string | 派生 | 聚合输出路径，仅 `aggregate: true` 时有效。相对清单目录。 |
| `defaults` | mapping | — | 本清单所有枚举共享的默认配置。 |
| `enums` | list | — | 枚举声明。至少需要一个。 |
| `targets` | list | — | 自定义模板输出。 |

## `defaults`

| 字段 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `features` | list of strings | `[string, parse, valid, text, names, values]` | 基础能力集。省略时使用内置默认集合；显式空列表表示关闭全部能力。 |
| `disableFeatures` | list of strings | `[]` | 从基础集合移除的能力。禁用 `sql` 会同时移除 `sql:int`/`sql:null`；禁用 `json` 会同时移除 `json:number`。 |
| `naming.prefix` | string | `type` | `type` 用枚举名作常量前缀；`none` 不加前缀；`literal:<text>` 使用固定前缀。 |
| `naming.identifier` | string | `camel` | `camel` 把取值名转为 CamelCase；`raw` 原样使用，要求是合法 Go 标识符。 |
| `naming.extraInitialisms` | list of strings | `[]` | 额外保持全大写的缩写，例如 `GRPC`。 |
| `naming.disableInitialisms` | bool | `false` | 关闭内置缩写表（`HTTP`、`URL`、`ID` 等）。 |
| `parse.caseSensitive` | bool | `true` | 为 false 时解析和别名匹配忽略大小写；仅大小写不同的取值文本会报错。 |
| `output.suffix` | string | `_enum.go` | 每个枚举输出文件的后缀。缺少 `.go` 时自动补上。 |
| `output.dir` | string | `""` | 每枚举输出目录，相对清单目录。 |
| `comments` | string | `auto` | `auto` 在缺少用户注释时生成注释；`none` 关闭自动注释，但保留用户注释、注释覆盖和弃用说明。 |
| `allowDuplicateValues` | bool | `false` | 允许多个名称映射到同一数字值。 |
| `typeMode` | string | `auto` | `auto` 检测已有声明，仅在缺失时生成类型；`emit` 总是生成；`existing` 从不生成，要求已有匹配声明。 |

## `enums`

| 字段 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `name` | string | 必填 | Go 类型名。 |
| `type` | string | `string` | 底层类型：`string`、任意有符号整数或无符号整数。 |
| `doc` | string | `""` | 类型文档注释。 |
| `comment` | string | `""` | 注释覆盖，优先于 `doc`。 |
| `comments` | string | 继承 | 枚举级 `auto` 或 `none` 注释模式。 |
| `package` | string | 清单包名 | 枚举级包名覆盖。非聚合模式下需要显式 `output`。 |
| `output` | string | 派生 | 枚举级输出路径，相对清单目录，必须以 `.go` 结尾。 |
| `features` | list of strings | `[]` | 在默认集合基础上追加的能力。 |
| `disableFeatures` | list of strings | `[]` | 从默认集合移除的能力。 |
| `typeMode` | string | 继承 | 枚举级 `auto`/`emit`/`existing` 覆盖。 |
| `start` | integer | `0` | 数字枚举自动递增起始值。字符串枚举不可用。 |
| `allowDuplicateValues` | bool | 继承 | 枚举级覆盖。 |
| `values` | list | 必填 | 至少一个取值。 |

## `values`

取值可以是标量或映射：

```yaml
values:
  - red
  - name: green
    value: GREEN
    doc: A custom green.
    comment: 注释覆盖。
    aliases: [g]
    deprecated: use teal
```

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `name` | string | 取值名称。映射形式必填；标量形式只设置名称。 |
| `value` | string 或 integer | 字符串枚举表示规范文本（默认等于名称）；数字枚举表示整数（默认自动递增）。 |
| `number` | integer | 字符串枚举启用 `sql:int` 时使用的整数表示。启用 `sql:int` 时每个取值都必须提供。 |
| `doc` | string | 生成常量的文档注释。 |
| `comment` | string | 注释覆盖，优先于 `doc`。 |
| `aliases` | list of strings | 额外可解析文本。别名不会生成常量。 |
| `deprecated` | bool 或 string | 添加 Go `Deprecated:` 注释。`true` 使用 “Do not use.”，字符串作为原因。 |

取值名会按 `naming` 转换为标识符。`http_status` 在默认缩写表下变成
`HTTPStatus`；`blue-green` 变成 `BlueGreen`。

数字枚举的规范文本就是取值名原文。

## `targets`

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `template` | string | `text/template` 模板路径，相对清单目录。 |
| `output` | string | 输出路径模板，相对清单目录，可引用模板数据。 |
| `perEnum` | bool | 按枚举执行一次，而不是按清单执行一次。 |

模板上下文：

```go
type TargetData struct {
	Manifest *model.Manifest
	Enums    []*model.EnumSpec
	Enum     *model.EnumSpec // 仅 perEnum: true 时非 nil
}
```

模板函数：`camel`、`snake`、`kebab`、`lower`、`upper`、`title`、`quote`、
`join`、`json`、`jsonPretty`、`default`。

## 通过 extends 共享默认配置

`extends` 指向另一个 YAML 清单并继承其 `defaults`：

```yaml
# base.yaml
version: 1
defaults:
  features: [string, parse, valid, text, names, values]
  naming:
    prefix: type
    identifier: camel
```

```yaml
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

合并规则：

- 只继承 `defaults`。基础文件中的 `package`、`aggregate`、`output`、`enums`
  和 `targets` 会被忽略。
- 嵌套的 `naming`、`parse`、`output` 逐字段合并；子清单未设置的字段继承基础值。
- `features` 等列表整体替换基础列表。`features: []` 表示清空；省略表示继承。
- `extends` 链递归解析，循环会报错。
- 基础文件只有在命名为 `enumer.yaml` 时才会被 `./...` 发现。

## 类型声明

`typeMode` 控制生成文件是否声明枚举类型：

- `auto`（默认）：扫描目标包。没有声明时生成类型；已有声明时引用；别名、泛型
  和底层类型不一致会报错。
- `emit`：总是生成类型。其它文件中已有声明时报错。
- `existing`：从不生成；声明必须存在且匹配。

```yaml
enums:
  - name: Color
    type: string
    values: [red, green, blue]
```

```go
// types.go
package example

type Color string
```

默认 `auto` 会检测该声明，只生成常量与方法；文件不存在时按清单生成类型。检测
只使用 `go/parser` 做语法扫描，不加载或类型检查包。

## 注释

- `doc` 与 `comment` 都可以覆盖生成注释，`comment` 优先。
- `comments: auto|none` 可在 defaults 和枚举级控制自动注释。
- 即使 `comments: none`，用户注释仍会输出。

## 格式化

生成的 Go 文件与输出路径以 `.go` 结尾的自定义目标都会使用标准库
`go/format.Source` 格式化。`.go` 目标必须渲染完整 Go 文件。

## 能力

| 能力 | 生成 API | 依赖 |
| --- | --- | --- |
| `string` | `String() string` | — |
| `parse` | `ErrInvalid<Name>`、`Parse<Name>` 与取值映射 | — |
| `valid` | `IsValid() bool` | `parse` |
| `flags` | 位标志辅助方法与组合感知 `String`/`Parse`/`IsValid` | `string`、`parse`、`valid` |
| `mustparse` | `MustParse<Name>` | `parse` |
| `text` | `MarshalText` / `UnmarshalText` | `parse`、`string` |
| `json` | 基于名称的 `MarshalJSON` / `UnmarshalJSON` | `parse`、`string` |
| `json:number` | 数字 `MarshalJSON` / `UnmarshalJSON` | `parse` |
| `yaml` | `MarshalYAML` / `UnmarshalYAML` | `parse`、`string` |
| `sql` | `Scan` / `Value` | `parse`、`string` |
| `sql:int` | `sql` 语义，使用整数存储 | `parse`、`string` |
| `sql:null` | `Null<Name>` 可空包装 | `sql` 与 `sql:int` 中恰好一个 |
| `flag` | `Set` / `Type` | `parse`、`string` |
| `names` | `<Name>Names() []string` | — |
| `values` | `<Name>Values() []<Name>` | — |
| `ptr` | `Ptr() *<Name>` | — |

`json` 与 `json:number` 互斥；`sql` 与 `sql:int` 互斥。`json:number` 只适用于
数字枚举；`sql:int` 只适用于字符串枚举。

## 位标志

`flags` 只适用于无符号整数枚举。每个取值必须是 0 或 2 的幂；最多一个零值；
取值不能重复。

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

生成的方法：

- `Has(flag) bool`
- `With(flag)`、`Clear(flag)`、`Toggle(flag)`
- `Flags() []<Name>`
- `String()` 用 `|` 连接已设置的标志名；未知位显示为 `Name(n)`。
- `Parse` 接受单个名称或 `|` 分隔的组合。
- `IsValid` 检查是否只包含已声明位。

`flags` 会自动添加 `string`、`parse`、`valid`，除非显式禁用。

## 输出路径解析

- 每枚举输出默认 `<dir>/<snake_case(enum_name)><suffix>`。
- 显式 `output` 相对清单目录解析，必须以 `.go` 结尾。
- 聚合输出默认 `<dir>/<清单文件名主干><suffix>`。
- 输出路径按大小写不敏感比较，避免跨平台冲突。
- 按包检查生成符号：类型名、常量名、`Parse…`、`ErrInvalid…`、`Null…`、
  `…Names`、`…Values` 以及内部生成标识符不能冲突。
- 带有当前清单精确生成头、但已不再产出的 `.go` 文件会在 `generate` 时清理。
  手写文件、其它清单生成的文件和自定义目标输出不会被清理。`--no-prune` 可关闭
  清理；`check` 会把孤儿文件报告为过期。

## 校验规则

- 枚举名必须是合法 Go 标识符。
- 命名转换后的常量名必须是合法 Go 标识符。
- 重复常量名、重复规范文本、重复别名、别名与文本冲突都会报错。
- 重复数字值默认报错；`allowDuplicateValues` 可开启。开启后 `String()` 返回
  首个声明的名称，`Parse` 接受全部名称。
- 数字值必须落在底层类型范围内；自动递增溢出会报错。
- `values` 为空或清单没有枚举会报错。
- `int` 和 `uint` 按 64 位校验；用于 32 位目标时需保证取值可表示。
- 字符串枚举不能使用 `start`。
- 数字枚举不能使用 `number`；请使用 `value`。
- `sql:int` 要求每个取值都有 `number`。
- `sql` 与 `sql:int` 的 `Scan` 不校验取值是否已声明；`IsValid` 是显式校验 API。
  `sql:int` 必须把整数映射到已声明值，因此会拒绝未知数字。

## 诊断

文本输出：

```
path/to/enumer.yaml:12:5: enumer: duplicate text "red" in enum Color (first declared at path/to/enumer.yaml:8:3)
```

`--json` 输出：

```json
{
  "version": 1,
  "diagnostics": [
    {
      "severity": "error",
      "code": "duplicate-text",
      "message": "duplicate text \"red\" in enum Color (first declared at ...)",
      "file": "path/to/enumer.yaml",
      "line": 12,
      "column": 5,
      "enum": "Color",
      "value": "red"
    }
  ]
}
```

退出码：`0` 成功，`1` 校验/生成/过期检查失败，`2` 命令行用法错误。

## 示例：聚合输出

```yaml
version: 1
package: http

aggregate: true
output: http_enum.go

enums:
  - name: Method
    type: string
    values: [GET, POST, PUT, DELETE]
  - name: Status
    type: int
    start: 100
    values: [continue, ok, created, accepted, no_content]
```

`enumer generate` 会生成单个 `http_enum.go`，包含两个枚举。

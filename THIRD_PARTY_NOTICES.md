# Third-party notices

`enumer` is distributed under the MIT License. It depends on the following
modules. All are permissively licensed and compatible with MIT.

| Module | License | Use |
| --- | --- | --- |
| github.com/spf13/cobra | Apache-2.0 | CLI |
| github.com/spf13/pflag | BSD-3-Clause | CLI flags (indirect) |
| github.com/inconshreveable/mousetrap | Apache-2.0 | Windows command handling (indirect) |
| gopkg.in/yaml.v3 | MIT and Apache-2.0 | Manifest parsing |
| github.com/pmezard/go-difflib | BSD-3-Clause | `enumer check --diff` |
| github.com/stretchr/testify | MIT | Tests |
| go.yaml.in/yaml/v3 | MIT and Apache-2.0 | Tests through testify (indirect) |

The MIT and Apache-2.0 licenses of `gopkg.in/yaml.v3` and `go.yaml.in/yaml/v3`
are compatible with this project's MIT license.

This file is informational. The automated dependency license check is
`license-eye dependency check` (see `.licenserc.yaml` and CI).

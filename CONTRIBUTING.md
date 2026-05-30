# Contributing to ComputeCoin

感谢你的兴趣。本项目目前由创始人(水镜先生)主导开发,采用 MIT 许可。

## 如何贡献

1. Fork 本仓库
2. 创建你的特性分支 (`git checkout -b feature/my-feature`)
3. 阅读 `CLAUDE.md` 了解编码约定和不变量
4. 确保 `make lint && make test` 全绿
5. 提交 Pull Request

## 贡献准则

- **不违反 CLAUDE.md 的 8 条不变量**(尤其是账本守恒、供给上限、来源合法声明)
- **不引入合规红线行为**(见 `docs/compliance.md`)
- 所有金额/数量必须用 `Decimal`,禁止 `float`
- 跨模块只通过 service 层互调
- 一个 PR 做一件事

## 行为准则

本项目遵循 [Contributor Covenant 2.1](https://www.contributor-covenant.org/version/2/1/code_of_conduct/)。

## 问题与讨论

- 技术问题: 开 GitHub Issue
- 商务合作 / 投资: zpmtyz@gmail.com (水镜先生)

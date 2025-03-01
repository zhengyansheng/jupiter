# JUPITER: Governance-oriented Microservice Framework

![logo](doc/logo.png)

[![GoTest](https://github.com/zhengyansheng/jupiter/workflows/unit-test/badge.svg)](https://github.com/zhengyansheng/jupiter/actions)
[![codecov](https://codecov.io/gh/zhengyansheng/jupiter/branch/master/graph/badge.svg)](https://codecov.io/gh/zhengyansheng/jupiter)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/zhengyansheng/jupiter?tab=doc)
[![Go Report Card](https://goreportcard.com/badge/github.com/zhengyansheng/jupiter)](https://goreportcard.com/report/github.com/douyu/jupiter)
![license](https://img.shields.io/badge/license-Apache--2.0-green.svg)

## Introduction

JUPITER is a governance-oriented microservice framework, which is being used for years at [Douyu](https://www.zhengyansheng.com).

## Online Demo

[Jupiter Console (Juno)](https://jupiterconsole.zhengyansheng.com)

```
Username: admin
Password: admin
```

## Documentation

See the [中文文档](http://jupiter.zhengyansheng.com/) for the Chinese documentation.

## Requirements

- Go version >= 1.19
- Docker

## Quick Start

1. Install [jupiter](https://github.com/zhengyansheng/jupiter/tree/master/cmd/jupiter) toolkit
1. Create example project from [jupiter-layout](https://github.com/zhengyansheng/jupiter-layout)
1. Download go mod dependencies
1. Run the example project with [jupiter](https://github.com/douyu/jupiter/tree/master/cmd/jupiter) toolkit
1. Just code yourself :-)

```bash
go install github.com/zhengyansheng/jupiter/cmd/jupiter@latest
jupiter new example-go
cd example-go
go mod tidy
docker compose -f test/docker-compose.yml up -d
jupiter run -c cmd/exampleserver/.jupiter.toml
```

## Learn More:
- [Juno](https://github.com/douyu/juno): **Microservice Governance System** for jupiter
- [Jupiter Layout](https://github.com/zhengyansheng/jupiter-layout): **Project Template** for jupiter
- [Examples](https://github.com/zhengyansheng/jupiter-examples)

## Bugs and Feedback

For bug report, questions and discussions please submit an issue.

## Contributing

Contributions are always welcomed! Please see [CONTRIBUTING](CONTRIBUTING.md) for detailed guidelines.

You can start with the issues labeled with good first issue.

<a href="https://github.com/zhengyansheng/jupiter/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=zhengyansheng/jupiter" />
</a>

## Contact

- [Wechat Group](https://jupiter.zhengyansheng.com/join/#%E5%BE%AE%E4%BF%A1)

```　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　
　◆◆◆◆　　　　　　◆◆◆◆◆　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　
　◆◆◆◆◆　　　　　◆◆◆◆◆　　　　　　　　　　　　　◆◆◆　　　　　　　　　　　　　　　　　　　　　
　◆◆◆◆◆　　　　　◆◆◆◆◆　　　　　　　　　　　　　◆◆◆　　　　　　　　　　　　　　　　　　　　　
　◆◆◆◆◆　　　　◆◆◆◆◆◆　　　　　　　　　　　　　◆◆◆　　　　　　　　　　　　　　　　　　　　　
　◆◆◆◆◆◆　　　◆◆◆◆◆◆　　　　◆◆◆◆◆◆　　◆◆◆◆◆◆　　　◆◆◆◆◆◆　　　◆◆◆◆◆◆　
　◆◆◆◆◆◆　　　◆◆◆◆◆◆　　◆◆◆◆◆◆◆◆◆　◆◆◆◆◆◆　◆◆◆◆◆◆◆◆◆　　◆◆◆◆◆◆　
　◆◆◆◆◆◆　　◆◆◆◆◆◆◆　　◆◆◆◆　　◆◆◆◆　◆◆◆　　　◆◆◆◆　　◆◆◆◆　◆◆◆◆　　　
　◆◆◆◆◆◆◆　◆◆◆◆◆◆◆　◆◆◆◆　　　◆◆◆◆　◆◆◆　　◆◆◆◆　　　◆◆◆◆　◆◆◆　　　　
　◆◆◆　◆◆◆　◆◆◆◆◆◆◆　◆◆◆　　　　　◆◆◆　◆◆◆　　◆◆◆　　　　　◆◆◆　◆◆◆　　　　
　◆◆◆　◆◆◆　◆◆◆◆◆◆◆　◆◆◆　　　　　◆◆◆　◆◆◆　　◆◆◆　　　　　◆◆◆　◆◆◆　　　　
　◆◆◆　◆◆◆◆◆◆　◆◆◆◆　◆◆◆　　　　　◆◆◆　◆◆◆　　◆◆◆　　　　　◆◆◆　◆◆◆　　　　
　◆◆◆　　◆◆◆◆◆　◆◆◆◆　◆◆◆◆　　　◆◆◆◆　◆◆◆　　◆◆◆◆　　　◆◆◆◆　◆◆◆　　　　
　◆◆◆　　◆◆◆◆◆　◆◆◆◆　　◆◆◆◆　　◆◆◆◆　◆◆◆　　　◆◆◆◆　　◆◆◆◆　◆◆◆　　　　
　◆◆◆　　◆◆◆◆　　◆◆◆◆　　◆◆◆◆◆◆◆◆◆　　◆◆◆◆◆　◆◆◆◆◆◆◆◆◆　　◆◆◆　　　　
　◆◆◆　　　◆◆◆　　◆◆◆◆　　　　◆◆◆◆◆◆　　　　◆◆◆◆　　　◆◆◆◆◆◆　　　◆◆◆　　　　
　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　　
```

基于[Gin](https://github.com/gin-gonic/gin) 框架的微服务脚手架，包含了一些常用的功能，如：命字服务、熔断、配置热加载、链路追踪、metrics、mysql、redis等。

## 主要内容
- http(s)服务
  - 平滑重启
  - pprof服务
  - gin框架无缝升级
- 中间件
  - Jwt中间件
  - accesslog中间件
  - 限流中间件
  - prometheus中间件
  - 链路追踪中间件
- 微服务组件
  - 配置管理，支持配置热加载，参考了[kratos](https://github.com/go-kratos/kratos)
  - Jwt认证
  - metrics，支持qps、请求耗时、错误请求数统计
  - 基于etcd的服务注册与发现
  - 服务熔断，基于[sentinel](https://github.com/alibaba/sentinel-golang)
  - 分布式链路追踪
- 存储
  - Mysql
  - Redis
## Features
- Http(s)服务： 支持gin框架无缝升级，封装了accesslog、jwt、ratelimit、trace、prometheus等常用中间件。
- Mysql&redis: 支持从名字服务和文件两种方式配置加载，支持配置平滑切换，并接入了trace。
- Trace: 基于opentracing和jaeger，实现分布式链路追踪
- Config: 支持配置热加载，参考了[kratos](https://github.com/go-kratos/kratos)
- Naming: 基于etcd的名字服务，实现了服务注册与服务发现

## Quick start

### Requirements
- Go version >= 1.13
- Go environment configure

```
export GOPROXY="https://goproxy.cn,direct"
```	

### 框架测试
- <font color=red>注意修改test/configs目录下的服务配置</font>

```shell
go get -u github.com/kaimixu/motor

go test -v ./...

```

### 使用Demo
参考[demo](https://github.com/kaimixu/motor_demo)

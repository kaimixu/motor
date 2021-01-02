# Motor

基于[Gin](https://github.com/gin-gonic/gin) 框架的微服务脚手架，封装了一些常用的功能，如：命字服务、熔断、配置热加载、链路追踪、metrics、mysql、redis等。

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
  - 配置管理，支持配置热加载，参考[kratos](https://github.com/go-kratos/kratos)
  - Jwt认证
  - metrics，支持qps、请求耗时、错误请求数统计
  - 基于etcd的服务注册与发现
  - 服务熔断，参考[sentinel](https://github.com/alibaba/sentinel-golang)
  - 分布式链路追踪
- 存储
  - Mysql
  - Redis
## 目录结构
- **conf:**&nbsp;配置管理
- **jwt:**&nbsp;jwt认证
- **log:**&nbsp;日志管理，基于[zap](https://pkg.go.dev/go.uber.org/zap)库
- **mysql:**&nbsp;支持名字服务和文件两个配置加载方式
- **naming:**&nbsp;基于etcd的名字服务
- **redis:**&nbsp;支持名字服务和文件两个配置加载方式
- **tolerant:**&nbsp;服务熔断
- **util:**&nbsp;项目中使用到的基础工具
- **test**&nbsp;项目go test使用的配置
## 框架测试
#### Requirements
- Go version >= 1.13
- Go environment configure

```
export GOPROXY="https://goproxy.cn,direct"
```	

- 修改test/configs目录下的配置

```
go test -v 
```

## 使用Demo
参考[demo](https://github.com/kaimixu/motor_demo)

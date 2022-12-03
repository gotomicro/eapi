<p align="center">
  <img width="144px" src="docs/logo.png" />
</p>

<h1 align="center">eAPI</h1>
<p align="center">一个通过分析代码生成 OpenAPI 文档的工具</p>
<p align="center"><b>还不用写注解</b></p>

## 介绍

eAPI 通过分析 AST 生成 **接口文档** 及 **前端代码**。与 swaggo/swag 等工具不同之处在于，eAPI 无需编写注解即可使用。另外，eAPI 还支持生成 Typescript 类型代码 和 前端接口请求代码。

eAPI 首先解析出代码中的路由（方法/路径）声明，得到接口的 Path、Method 及对应的 Handler 函数。然后再对 Handler 函数进行解析，得到 请求参数（Query/FormData/JSON-Payload等）、响应数据等信息。最终生成一份符合 OpenAPI 3 标准的 JSON 文档。

eAPI 目前支持了 gin 框架的文档生成，echo 等其他主流框架在计划中。如果你需要将 eAPI 应用在其他未被支持的框架，可以通过编写自定义插件的方式进行实现，或者给我们提交 PR。

## 安装

```shell
go install github.com/gotomicro/eapi/cmd/eapi@latest
```

## 如何使用

### 通过配置文件

**eapi.yaml**:
```yaml
output: docs
plugin: gin
dir: . # 需要解析的代码根目录
```

[完整的配置说明](#配置)

在代码根目录下执行:
```shell
eapi -c eapi.yaml
```

### 通过命令行参数

在代码根目录下执行:
```shell
eapi --output docs --plugin gin --dir .
```

执行以上命令后，会在 `/docs` 目录下生成 `swagger.json` 文件。

## 配置

如下是完整的配置文件示例:

```yaml
output: docs # 输出文档的目录
plugin: gin # 暂时只支持 gin
dir: '.' # 需要解析的代码目录

# 可选. 请求/响应数据中依赖的类型对应的包
depends:
 - github.com/gotomicro/gotoant
 - gorm.io/datatypes

# 可选. 插件配置. 用于自定义请求响应的函数调用
properties:
  # 自定义请求参数绑定
  request:
    - type: '*server/pkg/handler.CustomContext'
      method: 'Bind'
      return:
        data:
          type: 'args[0]' # 指定第一个函数参数为请求参数
  # 自定义响应函数
  response:
    - type: '*server/pkg/handler.CustomContext'
      method: 'JSONOK'
      return:
        contentType: 'application/json'  # 指定响应的 content-type
        data: # 这是一个嵌套的数据格式示例 '{"code":0,"msg":"hello",data:{...}}'
          type: 'object'
          properties:
            code:
              type: 'number'
            msg:
              type: 'string'
            data:
              optional: true # 是否可选. 默认 false
              type: 'args[0]' # 指定为第一个函数参数
        status: 200 # 指定为 200 状态码

# 可选. 配置代码生成器
generators:
  - name: ts # 生成器名称. 暂时只支持 "ts" (用于生成 typescript 类型)
    output: ./src/types # 输出文件的目录. 执行完成之后会在该目录下生成TS类型文件
```

### 代码生成器配置

如果需要使用代码生成功能，需要在配置文件内添加如下配置:
```yaml
# 可选
generators:
  - name: ts # 生成器名称. 暂时支持 "ts" | "umi" 
    output: ./src/types # 输出文件的目录. 执行完成之后会在该目录下生成TS类型文件
```

#### umi-request 请求代码生成
   
   umi 代码生成器用于生成适用于使用 `umi.js` 框架的前端接口请求代码及 TypeScript 类型。
   示例配置：
   ```yaml
   generators:
     - name: umi
       output: ./src/requests # 输出文件的目录
   ```
  
#### Typescript 类型生成

   ts 代码生成器用于生成 TypeScript 类型代码。
   示例配置：
   ```yaml
   generators:
     - name: ts
       output: ./src/types # 输出文件的目录
   ```

# TODO
- [ ] support for "echo" framework

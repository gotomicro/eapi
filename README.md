# eapi

EAPI 是一个通过分析 AST 生成 **接口文档** 及 **前端接口请求代码** 的命令行工具。作为 swaggo 等工具的替代品，EAPI 无需编写复杂的注解即可使用。

eapi 通过分析 AST 解析出代码中的路由（方法/路径）和 Handler 函数（请求响应参数/响应状态码等）获取接口的信息，进而生成 OpenAPI(Swagger) 文档。eapi 目前支持了 gin 框架的文档生成，echo 等其他主流框架在计划中。如果你需要将 eapi 应用在其他未被支持的框架，可以通过编写自定义插件的方式进行实现，或者给我们提交 PR。

由于本工具只分析 AST 中的关键部分（类似于静态分析）而不执行代码，因此如果你要在项目中使用本工具，必须要按照一定的约定编写代码，否则将不能生成完整的文档。虽然这给你的代码编写带来了一定的约束，但从另一个角度看，这也使得代码风格更加统一。

## 安装 Install

> goproxy 可能会存在缓存，可以优先尝试指定版本号进行安装。
> 
```shell
go install github.com/gotomicro/eapi/cmd/eapi@latest
```

## 如何使用
### 通过配置文件
**eapi.yaml**:
```yaml
output: docs
plugin: gin
dir: 'api' # 需要解析的代码根目录
```
> [完整的配置文件](#配置文件)

在代码根目录下执行:
```shell
eapi --config eapi.yaml
```

### 通过命令行参数

在代码根目录下执行:
```shell
eapi --output docs --plugin gin --dir api
```

执行以上命令后，会在 `/docs` 目录下生成 `swagger.json` 文件。


## 配置文件
如下是完整的配置文件示例:

> 不同版本的配置格式可能会发生一些调整，请确保安装的是最新的版本。另 goproxy 可能会存在缓存，可以优先尝试指定版本号进行安装。

```yaml
output: docs # 输出文档的目录
plugin: gin # 暂时只支持 gin
dir: 'api' # 需要解析的代码目录

# 可选. 请求/响应数据中依赖的类型对应的包
depends:
 - github.com/gotomicro/gotoant
 - gorm.io/datatypes

# 可选. 插件属性. 用于自定义请求响应的函数调用
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
              type: 'args[0]' # 指定为第一个函数参数
        status: 200 # 指定为 200 状态码

# 可选. 配置代码生成器
generators:
  - name: ts # 生成器名称. 暂时只支持 "ts" (用于生成 typescript 类型)
    output: ./src/types # 输出文件的目录. 执行完成之后会在该目录下生成TS类型文件
```

## 代码生成
如果需要使用代码生成功能，需要在配置文件内添加如下配置:
```yaml
# 可选
generators:
  - name: ts # 生成器名称. 暂时支持 "ts" | "umi" 
    output: ./src/types # 输出文件的目录. 执行完成之后会在该目录下生成TS类型文件
```

### 代码生成器
1. umi 
   
   umi 代码生成器用于生成适用于使用 `umi.js` 框架的前端接口请求代码及 TypeScript 类型。
   示例配置：
   ```yaml
   generators:
     - name: umi
       output: ./src/requests # 输出文件的目录
   ```
2. ts

   ts 代码生成器用于生成 TypeScript 类型代码。
   示例配置：
   ```yaml
   generators:
     - name: ts
       output: ./src/types # 输出文件的目录
   ```

# TODO
- [ ] support for `echo` framework

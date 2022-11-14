# ego-gen [WIP]

`ego-gen-api` 通过分析 AST 解析出代码中的路由（方法/路径）和 Handler 函数（请求响应参数/响应状态码等）获取接口的信息，进而生成 Swagger 文档。

由于本工具只分析 AST 中的关键部分（类似于静态分析）而不执行代码，因此并不能覆盖到任一种使用场景。如果你要在项目中使用本工具，必须要按照一定的约定编写代码，否则将不能生成完整的文档。虽然这给你的代码编写带来了一定的约束，但从另一个角度看，这也使得代码风格更加统一。

本工具目前支持了 `gin` 框架的文档生成。如果你需要将本项目进行扩展或应用于未支持的框架，可以通过编写自定义插件的方式进行实现。

## 安装 Install
```shell
go install github.com/gotomicro/ego-gen-api/cmd/egogen@latest
```

## 如何使用 How to Use
### 通过配置文件
**egen.yaml**:
```yaml
output: docs
plugin: gin
dir: 'api' # 需要解析的代码根目录
```
> [完整的配置文件](#配置文件)

在代码根目录下执行:
```shell
egen --config egn.yaml
```

### 通过命令行参数

在代码根目录下执行:
```shell
egen --output docs --plugin gin --dir api
```

执行以上命令后，会在 `/docs` 目录下生成 `swagger.json` 文件。


## 配置文件
如下是完整的配置文件示例:
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
    - type: '*github.com/clickvisual/clickvisual/api/pkg/component/core.Context'
      method: 'Bind'
      return:
        data: 'args[0]' # 指定第一个函数参数为请求参数
  # 自定义响应函数
  response:
    - type: '*github.com/clickvisual/clickvisual/api/pkg/component/core.Context'
      method: 'JSONOK'
      return:
        contentType: 'application/json' # 指定响应的 content-type
        data: 'args[0]' # 指定为第一个参数为接口响应参数
        status: 200 # 指定为 200 状态码
```

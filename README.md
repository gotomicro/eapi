# ego-gen-api
每次Go项目都要写接口文档，或者和前端对接API
哪怕用swagger，也要写好久的API注释

为什么不能直接使用`ast`，生成接口文档和对应的ts、flutter代码，提交到对应npm中，减少对接成本。

## 思路
* 如何确定一个ego的gin router
* 在返回值中，只要有一个参数是*egin.Component，就是我们的router函数
* 我们在从return中找到对应顺序的变量名
* 通过解析这个函数里数据，如果是这个变量名，并且是函数调用方法，使用的Group、GET、POST等名称，那么就是我们的方法
* 通过import找到对应的文件路径，解析他的dto
* 根据模板生成接口文档和前端代码，提交到npm中

https://yuroyoro.github.io/goast-viewer/index.html
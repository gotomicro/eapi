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
* 响应数据，通过解析到时候配置到函数名，找到最后一个，解析他，拿到响应schema
* 根据模板生成接口文档和前端代码，提交到npm中

## 指令
```bash
make runtmpls
```

## AST调试网站
https://yuroyoro.github.io/goast-viewer/index.html

## 开发思路
解析里面最麻烦的是响应数据。目前是通过命令行中获取响应的函数。
然后通过这个函数名，找到他return的数据。这个数据可能是以下几种类型
* 1 直接返回结构体、指针数据
```
c.JSONOK(Struct{})
c.JSONOK(&Struct{})
c.JSONOK(dto.Struct{})
c.JSONOK(&dto.Struct{})
```
* 2 定义的结构体、指针数据
```
var res Struct{}
var res &Struct{}
var res dto.Struct{}
var res &dto.Struct{}
c.JSONOK(res)
```
* 3 返回一个赋值的变量
```
res := Struct{}
res := &Struct{}
res := &dto.Struct{}
res := &dto.Struct{}
c.JSONOK(req)
```
* 4 返回一个赋值的变量的成员变量
```
req := Struct{
    Data: Struct{}
}
req := &Struct{
    Data: Struct{}

req := dto.Struct{
    Data: Struct{}
}
req := &dto.Struct{
    Data: Struct{}
}
c.JSONOK(req.Data)
```
* 5 函数返回一个变量
```
req := Func() // 同包下的函数
req := pkgName.Func() // 另一个包下的函数
req := pkgName.Params.Func() // 另一个包下的变量下的函数
c.JSONOK(req)
```
* 6 函数返回一个变量的成员变量
```
req := Func() // 同包下的函数
req := pkgName.Func() // 另一个包下的函数
req := pkgName.Params.Func() // 另一个包下的变量下的函数
c.JSONOK(req.Data)
```



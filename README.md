# godoc

godoc是一个基于注释翻译的多语言文档翻译项目，基于[ash](https://golang.org/pkg/go/ast/)和[godoc](https://pkg.go.dev/golang.org/x/tools/godoc?tab=doc)开发。

[online](http://doc.eudore.cn)，[cn](http://doc.eudore.cn/lang/cn/pkg/)。每分钟会从[wiki](https://github.com/eudore/godoc/wiki)同步一次数据。

**欢迎各位向wiki提交注释翻译**

# install

```bash
git clone https://github.com/eudore/godoc.git
cd godoc
docker build -t godoc .
docker run -d -p 8080:8080 godoc
```
# data

数据文件使名称用pkg.name.lang.txt的格式。

pkg是包名称'/'转换成'-',例如net/http包为net-http；name是这个文件的名称，可以忽略；lang是语言类型，需要是倒数第二段；txt是文件后缀也可以是md等其他后缀。

例如：net-http.cn.txt、net-http.cn.md、net-http.eudore.cn.txt

数据内容每两行是一条翻译数据，多数据间使用换行分割，具体查看data中的默认数据格式。

## 单行注释

```txt
// NewCond returns a new Cond with Locker l.

// NewCond函数返回带有Locker的新Cond。
```

## 多行注释

```txt
// Package sync provides basic synchronization primitives such as mutual
// exclusion locks. Other than the Once and WaitGroup types, most are intended
// for use by low-level library routines. Higher-level synchronization is
// better done via channels and communication.
//
// Values containing the types defined in this package should not be copied.

// sync包提供基本的同步操作，例如互斥锁。
// 除Once和WaitGroup类型外，大多数都供低级库例程使用。
// 更高级别的同步最好通过Channels和通信来完成。
//
//
// 在此包中定义的类型的值不应复制。
```
## 段注释

```txt
/*
// Package reflect implements run-time reflection, allowing a program to
// manipulate objects with arbitrary types. The typical use is to take a value
// with static type interface{} and extract its dynamic type information by
// calling TypeOf, which returns a Type.
//
// A call to ValueOf returns a Value representing the run-time data.
// Zero takes a Type and returns a Value representing a zero value
// for that type.
//
// See "The Laws of Reflection" for an introduction to reflection in Go:
*/

/*
// reflect包实现了运行时反射，从而允许程序处理任意类型的对象。 
// 典型的用法是使用静态类型interface{}来获取值，
// 并通过调用TypeOf来提取其动态类型信息，
// 该类型将返回Type。
//
// 调用ValueOf返回一个代表runtime数据的Value。 
// 零采用一个类型，
// 并返回一个表示该类型的零值的值。
//
// 有关Go语言中反射的介绍，请参见“反射法则”：
*/
```
# 目录结构
- api
- auth 鉴权相关
- database 数据库相关
- deploy ts部署相关
- instance ts实例相关
- monitor 系统、ts监控相关
- router 路由管理
- config 配置相关
- server 服务相关
- users 用户相关
- utils 工具模块

# 开发流程
先编写技术方案，然后编写具体结构体，再编写业务逻辑

# 主要功能
一键部署teamspeak服务，并自动生成ts3配置文件，并自动生成ts3管理员账号，并且监控teamspeak服务的稳定性和系统的稳定性，并加上多用户管理

# 技术栈
go 1.24.3 go-gin gorm sqlite yaml go-ts3 jwt
这是一个基于Go语言开发的Teamspeak服务一键部署工具，该工具使用Gin框架作为HTTP服务器，使用GORM作为数据库ORM，使用YAML作为配置文件，使用JWT作为用户认证。

# 服务器配置
cpu 2核 内存 2g 硬盘 50g

# 程序性能要求
由于存在监控模块，因此需要保证服务能够运行在2核2G内存50G硬盘的服务器上。可以适当增加缓存和延时

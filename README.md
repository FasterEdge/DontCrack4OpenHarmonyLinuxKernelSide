<div align="center">
<img src="https://s2.loli.net/2025/10/30/whQl7sJryj1GbHU.png" style="width:100px;" width="100"/>
<h2>DontCrack4OpenHarmonyLinuxKernelSide</h2>
<h3>开源鸿蒙Linux Kernel侧专用进程管理器</h3>
</div>


### 一、功能简介
- 用于提高开源鸿蒙Linux Kernel的`init.cfg`中用户自定义进程健壮性、可用性、时序稳定性
- 有效避免因`init.cfg`因细微格式错误导致设备无法开机、系统进程无法顺利启动、卡开机Logo
- 通过高层面隔离实现了与所管理进程的低耦合，不挑管理对象：`二进制可执行程序`、`sh脚本`、`Python脚本`、`Node脚本`、`Perl脚本`、`Ruby脚本`
- 实现将进程对应到端口号，可通Restful API实现获取日志、开关进程等操作
- 启动的进程可以配置独立的：程序路径、环境变量、启动参数、预处理脚本、是否自动重启、崩溃自动重启次数、是否立即启动、端口号、日志最大缓存行数、单行日志最大字节数
- 支持跨架构，免CGO，支持任何可被GO编译器编译程序的架构使用

### 二、基础用法
```
./DontCrack -path /home/test_program.sh -args "-key=test123 -shell=/bin/bash" -start-now
```
| 配置项                | 类型     | 默认值     | 说明                                                            | 
| ------------------ | ------ | ------- | ------------------------------------------------------------- |
| path               | string | ""      | 要管理的程序路径（支持可执行文件、shell脚本等）                                    | 
| args               | string | ""      | 传递给程序的参数（可选）                                                  | 
| pre                | string | ""      | 启动前要执行的命令（在sh中执行，可用&&连接多条命令） |
| auto-restart       | bool   | false   | 是否自动重启                                                        |  数                                                        |  
| start-now          | bool   | false   | 是否立即启动                                                        |  
| port               | string | "11883" | HTTP服务端口                                                      | 
| env                | string | ""      | 为子进程追加环境变量，如: "PATH=/usr/local/bin:/usr/bin FOO=bar"；用空格或分号分隔 |  
| log-capacity       | int    | 200     | 日志缓存的最大行数（默认200）                                              |  
| log-max-line-bytes | int    | 1048576 | 单行日志的最大字节数（用于bufio.Scanner，默认1MiB）                            | 
- 通过指定的IP:端口/startup 启动进程
- 通过指定的IP:端口/heartbeat 查看缓存的日志
- 通过指定的IP:端口/shutdown 终止进程
### 三、TodoList
- [ ] 重构代码结构，使其易于迭代
- [ ] 可选的日志本地完整存储、定时清除
- [ ] 自动重启次数二次修改、复位
- [ ] 无限次重启开关
- [ ] 远程访问和操作进程和查看日志时支持加密

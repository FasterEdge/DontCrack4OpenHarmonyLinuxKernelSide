# 示例子进程

这是一个用来全面测试进程管理器行为的示例子进程。

## 功能
- 打印启动参数与关键环境变量，便于确认传参是否生效
- 周期性输出 stdout/stderr，帮助验证日志抓取
- 支持模拟崩溃、优雅退出、挂起不退出等场景，便于测试重启和超时
- 响应 SIGINT/SIGTERM，执行清理后退出
- 可选通过文件记录重启计数，方便观察自动重启

## 构建
```bash
cd /childproc
GO111MODULE=on go build -o childproc ./childproc
```

## 运行示例
```bash
./childproc -mode normal -interval 1s -lifetime 10s
./childproc -mode crash -interval 500ms            # 模拟崩溃退出(退出码=42)
./childproc -mode hang                             # 挂起，不退出，用于测试管理器超时/强杀
./childproc -mode graceful -lifetime 5s            # 依赖信号优雅退出
./childproc -mode normal -state-file /tmp/counter  # 重启计数写入文件
```

## 重要参数
- `-mode`：`normal`(默认) / `crash` / `hang` / `graceful`
- `-interval`：输出间隔，默认 `1s`
- `-lifetime`：在 normal/crash 模式下运行多久后退出；空值表示一直运行。
- `-state-file`：若设置，则会读写该文件的计数用于观测重启次数。
- `-message`：自定义输出前缀，便于区分不同实例。

## 环境变量
- `EXTRA_INFO`：会被打印出来，方便验证环境传递。
- `RESTART_ENV_COUNT`：若存在，会尝试解析为整数，自增后打印回显，用于观察管理器是否覆盖/继承环境。

## 信号
- 支持 SIGINT/SIGTERM 触发优雅退出。
- 对于 hang 模式，信号会被记录但不会退出，便于测试管理器的强制 kill。

## 与管理器联动
- 将管理器的 `-path` 指向构建好的 `example/childproc` 可执行文件。
- 使用管理器的 `-args` 传递上述参数，例如：
  ```bash
  -args "-mode crash -interval 200ms -message manager"
  ```
- 配合 `-env "EXTRA_INFO=from_manager RESTART_ENV_COUNT=0"` 观察环境是否生效。


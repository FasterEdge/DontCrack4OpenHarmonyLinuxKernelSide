<div align="center">  
<img src="./Logo.png" alt="logo" width="100"/>  
<h2>DontCrack4OpenHarmonyLinuxKernelSide</h2>  
<h3>Dedicated Process Manager for the OpenHarmony Linux Kernel</h3>  
</div>  


### 1. Features

- Improves the robustness, availability and timing stability of user-defined processes in the OpenHarmony Linux Kernel's `init.cfg`
- Effectively avoids boot failures, system processes failing to start, or hanging on the boot logo caused by subtle formatting errors in `init.cfg`
- Supports managing these process types: `binary executable`, `sh script`
- Maps processes to port numbers; RESTful API for getting logs, starting/stopping processes, etc.
- Each managed process can be configured independently with: program path, environment variables, startup arguments, pre-processing script, auto-restart toggle, max crash-retry count, start-now flag, port number, log cache line limit, per-line byte limit, local log storage path, log retention period, etc.
- Cross-architecture, no CGO required; supports any architecture that the Go compiler can target

### 2. Basic Usage

```  
./DontCrack \
  -path "/Volumes/老固态/Projects/FasterEdge/DontCrack4OpenHarmonyLinuxKernelSide/example/childproc/childproc" \
  -args "-mode normal -interval 500ms -lifetime 5s" \
  -env "EXTRA_INFO=from_manager RESTART_ENV_COUNT=0" \
  -file-log -log-path ./example/logs/ -log-life-day 7 \
  -auto-restart -max-retries 2 \
  -start-now \
  -password 123456
```  

| Config | Type | Default | Description |
|--------|------|---------|-------------|
| path | string | "" | Path of the program to manage (supports executables, shell scripts etc.) |
| args | string | "" | Arguments passed to the program (optional) |
| pre | string | "" | Command to run before startup (executed in sh; multiple commands can be chained with && / ; / \|\|, default empty) |
| env | string | "" | Environment variables to append for the child process, e.g. "PATH=/usr/local/bin:/usr/bin FOO=bar"; separated by spaces or semicolons |
| auto-restart | bool | false | Whether to auto-restart on crash |
| max-retries | int | 3 | Max retry count (-1 means unlimited, default 3) |
| start-now | bool | false | Whether to start immediately |
| port | int | 11883 | HTTP service port |
| password | string | "" | Password for managing the process (optional; no password protection if empty) |
| log-capacity | int | 200 | Max lines of cached logs (default 200) |
| log-max-line-bytes | int | 1048576 | Max bytes per log line (for bufio.Scanner, default 1 MiB) |
| file-log | bool | false | Whether to enable file logging (default false) |
| log-path | string | /data/logs/proc_manager/ | Local log file directory (default /data/logs/proc_manager/, subdirectories created per process name) |
| log-life-day | int | 7 | Log retention in days (default 7; expired logs cleaned when new logs are written) |

### 3. API Documentation

> /startup

- Description: Starts the process and resets the retry count
- Method: GET, POST
- Request parameters:
  ```
  password: secret key (optional params parameter)
  ```
- Response type: text
- Example response:
    ```
    ok
    ```

> /heartbeat

- Description: Returns heartbeat information, including startup status and cached logs (logs are cleared after reading)
- Method: GET, POST
- Request parameters:
  ```
  password: secret key (optional params parameter)
  ```
- Response type: JSON
- Example response:
     ```
	{
	"version": "1.0.20260901",
	"state": "stopped",
	"info": "Process manager running normally",
	"timestamp": "2026-02-24 15:28:04",
	"logs": [
	"[STDERR] 2026/02/24 15:27:55.647714 env restart count -> 1",
	"[STDERR] 2026/02/24 15:27:55.648316 childproc start | pid=32054 | mode=normal | interval=1s | lifetime=5s | msg=",
	"[STDERR] 2026/02/24 15:27:55.648345 args: /Volumes/老固态/Projects/FasterEdge/DontCrack4OpenHarmonyLinuxKernelSide/example/childproc/childproc -mode normal -interval 1s -lifetime 5s",
	"[STDERR] 2026/02/24 15:27:55.648352 env EXTRA_INFO=from_manager",
	"[STDERR] 2026/02/24 15:27:55.648353 env RESTART_ENV_COUNT=0",
	"[STDERR] 2026/02/24 15:27:56.668982 tick at 2026-02-24T15:27:56.64879725+08:00",
	"[STDERR] stderr burst at 2026-02-24T15:27:56.64879725+08:00",
	"[STDERR] 2026/02/24 15:27:57.686757 tick at 2026-02-24T15:27:57.648801166+08:00",
	"[STDERR] stderr burst at 2026-02-24T15:27:57.648801166+08:00",
	"[STDERR] 2026/02/24 15:27:58.652718 tick at 2026-02-24T15:27:58.648811791+08:00",
	"[STDERR] stderr burst at 2026-02-24T15:27:58.648811791+08:00",
	"[STDERR] 2026/02/24 15:27:59.668692 tick at 2026-02-24T15:27:59.648816958+08:00",
	"[STDERR] stderr burst at 2026-02-24T15:27:59.648816958+08:00",
	"[STDERR] 2026/02/24 15:28:00.686286 tick at 2026-02-24T15:28:00.648823625+08:00",
	"[STDERR] stderr burst at 2026-02-24T15:28:00.648823625+08:00",
	"[STDERR] 2026/02/24 15:28:00.686300 lifetime reached, exiting normally"
	],
	"process_pid": 0,
	"process_path": "/Volumes/老固态/Projects/FasterEdge/DontCrack4OpenHarmonyLinuxKernelSide/example/childproc/childproc",
	"restart_count": 0,
	"file_type": "binary_executable",
	"last_exit_time": "2026-02-24 15:28:00",
	"program_args": "-mode normal -interval 1s -lifetime 5s",
	"extra_env_raw": "PATH=/opt/tools/bin EXTRA_INFO=from_manager RESTART_ENV_COUNT=0"
	}
	```

> /shutdown

- Description: Terminates the process
- Method: GET, POST
- Request parameters:
  ```
  password: secret key (optional params parameter)
  ```
- Response type: text
- Example response:
  ```
  ok
  ```

### 4. Details

- Use full paths for the managed process's Path whenever possible
- OpenHarmony generally has only one sh command tool at `/bin/sh`, no bash, but it can still execute scripts
- Files ending with `.sh` or containing `#!` on the first line are recognized as scripts and executed by sh
- When password protection is enabled, API requests must include the `password` parameter in the URL, e.g. `xxx/startup?password=123456`

### 5. Tips

- When used standalone, if `&` is used at the end of a command in a session, the operation may be killed when the session ends
- Go's `log.Printf` writes to `os.Stderr` by default, so log entries from child processes will show as `[STDERR]`; switch to `fmt.Println` to get non-`[STDERR]` messages
- Since processes can be controlled via encrypted HTTP, you can combine this with the OpenHarmony device's role, AI + MCP (or Skills) to implement various automation operations
- Besides using direct functionality for process robustness, the repeated restart mechanism can also be used for process polling, and the pre-script can implement delayed startup, waiting for dependent processes, etc.
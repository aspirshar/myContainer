# MyContainer

MyContainer 是一个用 Go 语言实现的简单容器运行时，提供了基本的容器功能，包括容器生命周期管理、资源限制、网络管理等特性。

## 主要特性

### 1. 容器管理
- ✅ 创建和运行容器（支持交互式和后台运行）
- ✅ 停止和删除容器
- ✅ 查看容器列表
- ✅ 查看容器日志
- ✅ 在运行中的容器中执行命令（exec）
- ✅ 容器提交（保存容器状态为镜像）
- ✅ 容器命名

### 2. 资源隔离与限制
- ✅ 使用 Linux Namespace 实现资源隔离
  - PID Namespace：进程隔离
  - Mount Namespace：挂载点隔离
  - Network Namespace：网络隔离
  - UTS Namespace：主机名隔离
  - IPC Namespace：进程间通信隔离
- ✅ 使用 Cgroups 实现资源限制（同时支持 cgroup v1 和 v2）
  - CPU 限制（通过 -cpu 参数设置，如 0.5 表示 50%）
  - 内存限制（通过 -mem 参数设置，如 100m）
  - CPU Set 限制（通过 -cpuset 参数设置，如 0,1）

### 3. 网络管理
- ✅ 创建和管理网络
- ✅ Bridge 网络驱动
- ✅ 容器网络连接
- ✅ 端口映射（-p 参数）
- ✅ IP 地址管理（IPAM）

### 4. 存储管理
- ✅ 容器根文件系统（基于 overlay2）
- ✅ Volume 数据卷挂载（-v 参数）
- ✅ 文件系统隔离（pivot_root）
- ✅ 镜像管理（支持 tar 格式）

### 5. 环境变量管理
- ✅ 设置容器环境变量（-e 参数）
- ✅ 环境变量传递和隔离

## 项目结构

```
myContainer/
├── cgroups/           # Cgroups 资源限制实现
│   ├── cgroup_manager_v1.go
│   ├── cgroup_manager_v2.go
│   ├── fs/            # cgroup v1 文件系统操作
│   ├── fs2/           # cgroup v2 文件系统操作
│   └── resource/      # 资源配置
├── container/         # 容器核心功能实现
│   ├── container_info.go
│   ├── container_process.go
│   ├── init.go
│   ├── rootfs.go
│   └── volume.go
├── network/           # 容器网络实现
│   ├── bridge_driver.go
│   ├── ipam.go
│   ├── model.go
│   └── network.go
├── nsenter/           # Namespace 操作（C代码）
│   └── nsenter_linux.go
├── utils/             # 工具函数
├── config/            # 配置管理
├── constant/          # 常量定义
├── images/            # 镜像文件
│   ├── busybox.tar
│   └── echo.tar
├── overlay2/          # overlay2 文件系统
├── main.go            # 主程序入口
├── main_command.go    # 命令行定义
├── run.go             # 运行容器
├── exec.go            # 执行命令
├── stop.go            # 停止容器
├── list.go            # 列出容器
├── logs.go            # 查看日志
└── commit.go          # 提交容器
```

## 主要命令

### 容器生命周期管理

```bash
# 运行容器（交互式）
./myContainer run -it busybox /bin/sh

# 运行容器（后台运行）
./myContainer run -d -name my-container busybox /bin/sh -c "while true; do sleep 1; done"

# 运行容器（带资源限制）
./myContainer run -mem 100m -cpu 0.5 -cpuset 0,1 busybox /bin/sh

# 运行容器（带数据卷）
./myContainer run -v /host/path:/container/path busybox /bin/sh

# 运行容器（带环境变量）
./myContainer run -e MY_VAR=value busybox /bin/sh

# 运行容器（带网络）
./myContainer run -net mynet -p 8080:80 busybox /bin/sh

# 查看容器列表
./myContainer ps

# 查看容器日志
./myContainer logs [container_id]

# 进入容器执行命令
./myContainer exec [container_id] /bin/sh

# 停止容器
./myContainer stop [container_id]

# 删除容器
./myContainer rm [container_id]

# 强制删除运行中的容器
./myContainer rm -f [container_id]

# 提交容器为镜像
./myContainer commit [container_id] [image_name]
```

### 网络管理

```bash
# 创建网络
./myContainer network create --driver bridge --subnet 192.168.0.0/24 [network_name]

# 列出网络
./myContainer network list

# 删除网络
./myContainer network remove [network_name]
```

## 支持的镜像

目前支持以下镜像格式：
- `busybox.tar` - BusyBox 轻量级 Linux 发行版
- `echo.tar` - 简单的 echo 服务镜像

## 技术栈

- **Go 1.23.0** - 主要开发语言
- **Linux Namespace** - 资源隔离
- **Cgroups** - 资源限制（v1/v2）
- **Overlay2** - 文件系统
- **Bridge 网络** - 容器网络
- **iptables** - 网络规则
- **netlink** - 网络配置

## 依赖库

- `github.com/pkg/errors` - 错误处理
- `github.com/sirupsen/logrus` - 结构化日志
- `github.com/urfave/cli` - 命令行工具
- `github.com/vishvananda/netlink` - 网络配置
- `github.com/vishvananda/netns` - 网络命名空间操作
- `golang.org/x/sys` - 系统调用

## 使用示例

### 1. 基本容器操作

```bash
# 编译项目
go build -o myContainer .

# 设置环境变量
export MYCONTAINER_ROOT=/root/go/myContainer

# 运行交互式容器
./myContainer run -it busybox /bin/sh

# 在另一个终端查看容器
./myContainer ps

# 进入运行中的容器
./myContainer exec [container_id] /bin/sh -c "echo Hello World"
```

### 2. 资源限制测试

```bash
# 限制内存为 100MB
./myContainer run -mem 100m -name mem-test busybox /bin/sh

# 限制 CPU 为 50%
./myContainer run -cpu 0.5 -name cpu-test busybox /bin/sh

# 限制 CPU 核心为 0,1
./myContainer run -cpuset 0,1 -name cpuset-test busybox /bin/sh
```

### 3. 数据卷挂载

```bash
# 创建测试目录
mkdir -p /tmp/test-volume
echo "Hello from host" > /tmp/test-volume/test.txt

# 挂载到容器
./myContainer run -v /tmp/test-volume:/data busybox /bin/sh -c "cat /data/test.txt"
```

### 4. 网络功能

```bash
# 创建网络
./myContainer network create --driver bridge --subnet 192.168.0.0/24 mynet

# 运行带网络的容器
./myContainer run -net mynet -p 8080:80 busybox /bin/sh
```

## 系统要求

- **操作系统**: Linux（内核版本 >= 3.10）
- **权限**: 需要 root 权限运行
- **内核特性**: 需要支持 Namespace 和 Cgroups
- **工具**: 需要安装 iptables、bridge-utils

## 注意事项

1. **权限要求**: 必须使用 root 权限运行
2. **系统兼容性**: 仅支持 Linux 系统
3. **内核要求**: 需要内核支持 Namespace 和 Cgroups 特性
4. **网络依赖**: 需要 iptables 和 bridge-utils 工具
5. **镜像格式**: 目前支持 tar 格式的镜像文件
6. **开发目的**: 这是一个学习和教育目的的容器实现

## 故障排除

### 常见问题

1. **权限错误**: 确保使用 root 权限运行
2. **网络创建失败**: 检查 iptables 是否安装
3. **资源限制不生效**: 检查 cgroup 控制器是否启用
4. **镜像加载失败**: 确保镜像文件存在且格式正确

### 调试方法

```bash
# 查看详细日志
./myContainer run -it busybox /bin/sh 2>&1 | tee container.log

# 检查 cgroup 配置
cat /sys/fs/cgroup/memory/memory.max
cat /sys/fs/cgroup/cpu/cpu.max

# 检查网络配置
ip link show
iptables -L
```

## 开发说明

这是一个学习和教育目的的容器实现，展示了容器技术的核心概念和基本实现：

- **Namespace 隔离**: 演示了进程、网络、文件系统等资源的隔离
- **Cgroups 限制**: 展示了 CPU、内存等资源的限制机制
- **网络管理**: 实现了基本的容器网络功能
- **文件系统**: 使用 overlay2 实现分层文件系统

虽然不适合在生产环境中使用，但可以帮助理解容器技术的工作原理。

## 许可证

MIT License
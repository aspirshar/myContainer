# MyContainer

MyContainer 是一个用 Go 语言实现的简单容器运行时，提供了基本的容器功能，包括容器生命周期管理、资源限制、网络管理等特性。

## 主要特性

### 1. 容器管理
- 创建和运行容器
- 停止和删除容器
- 查看容器列表
- 查看容器日志
- 在运行中的容器中执行命令
- 容器提交（保存容器状态）

### 2. 资源隔离与限制
- 使用 Linux Namespace 实现资源隔离
  - PID Namespace：进程隔离
  - Mount Namespace：挂载点隔离
  - Network Namespace：网络隔离
- 使用 Cgroups 实现资源限制（同时支持 cgroup v1 和 v2）
  - CPU 限制（通过 -cpu 参数设置）
  - 内存限制（通过 -mem 参数设置）
  - CPU Set 限制（通过 -cpuset 参数设置）

### 3. 网络管理
- 创建和管理网络
- Bridge 网络驱动
- 容器网络连接和断开
- 端口映射
- IP 地址管理（IPAM）

### 4. 存储管理
- 容器根文件系统
- Volume 数据卷
- 文件系统隔离（pivot_root）

## 项目结构

```
myContainer/
├── cgroups/        # Cgroups 资源限制实现
├── container/      # 容器核心功能实现
├── network/        # 容器网络实现
├── nsenter/        # Namespace 操作
└── utils/          # 工具函数
```

## 主要命令

```bash
# 运行容器
./myContainer run [command]

# 查看容器列表
./myContainer ps

# 查看容器日志
./myContainer logs [container_id]

# 进入容器执行命令
./myContainer exec [container_id] [command]

# 停止容器
./myContainer stop [container_id]

# 删除容器
./myContainer rm [container_id]

# 提交容器
./myContainer commit [container_id] [image_name]

# 网络管理
./myContainer network create --driver bridge --subnet 192.168.0.0/24 [network_name]
./myContainer network list
./myContainer network rm [network_name]
```

## 技术栈

- Go 1.23.0
- Linux 系统调用
- Namespace
- Cgroups
- Bridge 网络
- iptables

## 依赖

- github.com/pkg/errors - 错误处理
- github.com/sirupsen/logrus - 日志库
- github.com/urfave/cli - 命令行工具库
- github.com/vishvananda/netlink - 网络配置
- github.com/vishvananda/netns - 网络命名空间操作
- golang.org/x/sys - 系统调用

## 使用示例

1. 运行一个简单的容器：
```bash
./myContainer run -it ubuntu /bin/bash
```

2. 创建一个网络并运行带网络的容器：
```bash
./myContainer network create --driver bridge --subnet 192.168.0.0/24 mynet
./myContainer run -net mynet -p 8080:80 nginx
```

## 注意事项

1. 需要 root 权限运行
2. 仅支持 Linux 系统
3. 需要内核支持 Namespace 和 Cgroups 特性

## 开发说明

这是一个学习和教育目的的容器实现，展示了容器技术的核心概念和基本实现。它不适合在生产环境中使用，但可以帮助理解容器技术的工作原理。

## 许可证

MIT License
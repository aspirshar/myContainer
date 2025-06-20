package main

import (
	"myContainer/cgroups"
	"myContainer/cgroups/resource"
	"myContainer/container"
	"myContainer/network"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

/*
Run函数主要做了以下几件事：
1.生成容器ID
2.调用container.NewParentProcess创建父进程
3.调用parent.Start启动父进程
4.创建cgroup manager, 并通过调用set和apply设置资源限制并使限制在容器上生效
5.如果指定了网络信息则进行配置
6.记录容器信息
7.在子进程创建后才能通过pipe来发送参数
8.如果是前台运行，则等待容器进程结束
9.如果是后台运行，则等待子进程退出，并清理工作
*/

func Run(tty bool, comArray, envSlice []string, res *resource.ResourceConfig, volume, containerName, imageName string,
	net string, portMapping []string) {
	containerId := container.GenerateContainerID() // 生成 10 位容器 id

	// 创建父进程
	parent, writePipe := container.NewParentProcess(tty, volume, containerId, imageName, envSlice)
	if parent == nil {
		log.Errorf("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Errorf("Run parent.Start err:%v", err)
		return
	}

	// 创建cgroup manager, 并通过调用set和apply设置资源限制并使限制在容器上生效
	cgroupManager := cgroups.NewCgroupManager("mycontainer-cgroup")
	//defer cgroupManager.Destroy() // 由单独的 goroutine 来处理
	_ = cgroupManager.Set(res)
	_ = cgroupManager.Apply(parent.Process.Pid)

	var containerIP string
	// 如果指定了网络信息则进行配置
	if net != "" {
		containerInfo := &container.Info{
			Id:          containerId,
			Pid:         strconv.Itoa(parent.Process.Pid),
			Name:        containerName,
			PortMapping: portMapping,
		}
		ip, err := network.Connect(net, containerInfo)
		if err != nil {
			log.Errorf("Error Connect Network %v", err)
			return
		}
		containerIP = ip.String()
	}

	// 记录容器信息
	containerInfo, err := container.RecordContainerInfo(parent.Process.Pid, comArray, containerName, containerId,
		volume, net, containerIP, portMapping)
	if err != nil {
		log.Errorf("Record container info error %v", err)
		return
	}

	// 在子进程创建后才能通过pipe来发送参数
	sendInitCommand(comArray, writePipe)

	if tty {
		_ = parent.Wait() // 前台运行，等待容器进程结束
	}
	// 然后创建一个 goroutine 来处理后台运行的清理工作
	go func() {
		if !tty {
			// 等待子进程退出
			_, _ = parent.Process.Wait()
		}

		// 清理工作
		container.DeleteWorkSpace(containerId, volume)
		container.DeleteContainerInfo(containerId)
		if net != "" {
			network.Disconnect(net, containerInfo)
		}

		// 销毁 cgroup
		cgroupManager.Destroy()
	}()
}

// sendInitCommand 通过writePipe将指令发送给子进程
func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Infof("command all is %s", command)
	_, _ = writePipe.WriteString(command)
	_ = writePipe.Close()
}

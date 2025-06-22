package main

import (
	"encoding/json"
	"fmt"
	"github.com/aspirshar/myContainer/network"
	"github.com/aspirshar/myContainer/utils"
	"os"
	"path"
	"strconv"
	"syscall"

	"github.com/aspirshar/myContainer/constant"
	"github.com/aspirshar/myContainer/container"

	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"
)

func stopContainer(containerId string) {
	// 1. 根据容器Id查询容器信息
	containerInfo, err := getInfoByContainerId(containerId)
	if err != nil {
		log.Errorf("Get container %s info error %v", containerId, err)
		return
	}
	pidInt, err := strconv.Atoi(containerInfo.Pid)
	if err != nil {
		log.Errorf("Conver pid from string to int error %v", err)
		return
	}
	// 2.发送SIGTERM信号
	if err = syscall.Kill(pidInt, syscall.SIGTERM); err != nil {
		log.Errorf("Stop container %s error %v", containerId, err)
		// 如果进程不存在，直接更新状态为STOP
		if err == syscall.ESRCH {
			log.Infof("Process %d does not exist, updating container status to STOP", pidInt)
		} else {
			return
		}
	}
	// 3.修改容器信息，将容器置为STOP状态，并清空PID
	containerInfo.Status = container.STOP
	containerInfo.Pid = " "
	newContentBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Errorf("Json marshal %s error %v", containerId, err)
		return
	}
	// 4.重新写回存储容器信息的文件
	dirPath := fmt.Sprintf(container.InfoLocFormat, containerId)
	configFilePath := path.Join(dirPath, container.ConfigName)
	if err = os.WriteFile(configFilePath, newContentBytes, constant.Perm0622); err != nil {
		log.Errorf("Write file %s error:%v", configFilePath, err)
	}
}

func getInfoByContainerId(containerId string) (*container.Info, error) {
	dirPath := fmt.Sprintf(container.InfoLocFormat, containerId)
	configFilePath := path.Join(dirPath, container.ConfigName)
	contentBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "read file %s", configFilePath)
	}
	var containerInfo container.Info
	if err = json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return nil, err
	}
	return &containerInfo, nil
}

func removeContainer(containerId string, force bool) {
	containerInfo, err := getInfoByContainerId(containerId)
	if err != nil {
		log.Errorf("Get container %s info error %v", containerId, err)
		fmt.Printf("Error: Container '%s' does not exist\n", containerId)
		return
	}

	switch containerInfo.Status {
	case container.STOP: // STOP 状态容器直接删除即可
		// 先删除配置目录，再删除rootfs 目录
		if err = container.DeleteContainerInfo(containerId); err != nil {
			log.Errorf("Remove container [%s]'s config failed, detail: %v", containerId, err)
			return
		}
		utils.DeleteWorkSpace(utils.GetRoot(containerId), containerInfo.Volume)
		if containerInfo.NetworkName != "" { // 清理网络资源
			if err = network.Disconnect(containerInfo.NetworkName, containerInfo); err != nil {
				log.Errorf("Remove container [%s]'s config failed, detail: %v", containerId, err)
				return
			}
		}
		fmt.Printf("Container '%s' has been removed\n", containerId)
	case container.RUNNING: // RUNNING 状态容器如果指定了 force 则先 stop 然后再删除
		if !force {
			log.Errorf("Couldn't remove running container [%s], Stop the container before attempting removal or"+
				" force remove", containerId)
			fmt.Printf("Error: Couldn't remove running container '%s'. Stop the container before attempting removal or force remove with -f\n", containerId)
			return
		}
		log.Infof("force delete running container [%s]", containerId)
		fmt.Printf("Force removing running container '%s'...\n", containerId)
		stopContainer(containerId)
		// 重新加载容器信息
		containerInfo, err = getInfoByContainerId(containerId)
		if err != nil {
			log.Errorf("Get container %s info error %v", containerId, err)
			return
		}
		if containerInfo.Status == container.STOP {
			removeContainer(containerId, force)
		} else {
			log.Errorf("Couldn't remove container, stop failed")
			fmt.Printf("Error: Failed to stop container '%s'\n", containerId)
		}
	default:
		log.Errorf("Couldn't remove container,invalid status %s", containerInfo.Status)
		fmt.Printf("Error: Couldn't remove container '%s', invalid status: %s\n", containerId, containerInfo.Status)
		return
	}
}
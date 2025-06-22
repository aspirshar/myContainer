package main

import (
	"fmt"
	"io"
	"os"

	"github.com/aspirshar/myContainer/container"

	log "github.com/sirupsen/logrus"
)

func logContainer(containerIdOrName string) {
	// 首先尝试通过容器名称获取容器ID
	containerId := containerIdOrName
	containerInfo, err := container.GetContainerInfoByName(containerIdOrName)
	if err != nil {
		// 如果通过名称查找失败，假设输入的是容器ID，直接使用
		log.Infof("Container name '%s' not found, treating as container ID", containerIdOrName)
	} else {
		// 如果通过名称找到了容器，使用其ID
		containerId = containerInfo.Id
		log.Infof("Found container '%s' with ID: %s", containerIdOrName, containerId)
	}

	logFileLocation := fmt.Sprintf(container.InfoLocFormat, containerId) + container.GetLogfile(containerId)
	file, err := os.Open(logFileLocation)
	defer file.Close()
	if err != nil {
		log.Errorf("Log container open file %s error %v", logFileLocation, err)
		return
	}
	content, err := io.ReadAll(file)
	if err != nil {
		log.Errorf("Log container read file %s error %v", logFileLocation, err)
		return
	}
	_, err = fmt.Fprint(os.Stdout, string(content))
	if err != nil {
		log.Errorf("Log container Fprint  error %v", err)
		return
	}
}
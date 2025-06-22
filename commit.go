package main

import (
	"github.com/aspirshar/myContainer/container"
	"github.com/aspirshar/myContainer/utils"
	"os/exec"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var ErrImageAlreadyExists = errors.New("Image Already Exists")

func commitContainer(containerIDOrName, imageName string) error {
	// 首先尝试通过容器名称获取容器ID
	containerID := containerIDOrName
	containerInfo, err := container.GetContainerInfoByName(containerIDOrName)
	if err != nil {
		// 如果通过名称查找失败，假设输入的是容器ID，直接使用
		log.Infof("Container name '%s' not found, treating as container ID", containerIDOrName)
	} else {
		// 如果通过名称找到了容器，使用其ID
		containerID = containerInfo.Id
		log.Infof("Found container '%s' with ID: %s", containerIDOrName, containerID)
	}

	mntPath := utils.GetMerged(containerID)
	imageTar := utils.GetImage(imageName)
	exists, err := utils.PathExists(imageTar)
	if err != nil {
		return errors.WithMessagef(err, "check is image [%s/%s] exist failed", imageName, imageTar)
	}
	if exists {
		return ErrImageAlreadyExists
	}
	log.Infof("commitContainer imageTar:%s", imageTar)
	if _, err = exec.Command("tar", "-czf", imageTar, "-C", mntPath, ".").CombinedOutput(); err != nil {
		return errors.WithMessagef(err, "tar folder %s failed", mntPath)
	}
	return nil
}
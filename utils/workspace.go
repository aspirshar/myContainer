package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
)

// DeleteWorkSpace 删除容器文件系统
func DeleteWorkSpace(rootPath, volume string) {
	log.Infof("开始删除容器文件系统: %s", rootPath)
	
	if volume != "" {
		// 卸载数据卷
		mntPath := GetMerged(path.Base(rootPath))
		_, containerPath, err := volumeExtract(volume)
		if err != nil {
			log.Errorf("extract volume failed, maybe volume contains separator, err: %v", err)
			return
		}
		umountVolume(mntPath, containerPath)
	}
	
	// 卸载并删除整个 overlay2 目录
	mntPath := GetMerged(path.Base(rootPath))
	log.Infof("卸载overlay文件系统: %s", mntPath)
	umountOverlay(mntPath)
	
	// 检查目录是否存在
	if exists, _ := PathExists(rootPath); exists {
		log.Infof("删除目录: %s", rootPath)
		if err := os.RemoveAll(rootPath); err != nil {
			log.Errorf("remove workspace dir %s failed, err: %v", rootPath, err)
		} else {
			log.Infof("成功删除目录: %s", rootPath)
		}
	} else {
		log.Infof("目录不存在，无需删除: %s", rootPath)
	}
}

// umountOverlay 卸载 overlayfs
func umountOverlay(mntPath string) {
	// 检查挂载点是否存在
	if exists, _ := PathExists(mntPath); !exists {
		log.Infof("挂载点不存在，无需卸载: %s", mntPath)
		return
	}
	
	cmd := exec.Command("umount", mntPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("umount overlayfs %s failed, err: %v", mntPath, err)
		// 尝试强制卸载
		log.Infof("尝试强制卸载: %s", mntPath)
		forceCmd := exec.Command("umount", "-f", mntPath)
		forceCmd.Stdout = os.Stdout
		forceCmd.Stderr = os.Stderr
		if forceErr := forceCmd.Run(); forceErr != nil {
			log.Errorf("强制卸载也失败: %v", forceErr)
		} else {
			log.Infof("强制卸载成功: %s", mntPath)
		}
	} else {
		log.Infof("成功卸载overlay文件系统: %s", mntPath)
	}
}

// volumeExtract 通过冒号分割解析volume目录
func volumeExtract(volume string) (sourcePath, destinationPath string, err error) {
	parts := strings.Split(volume, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid volume [%s], must split by `:`", volume)
	}
	sourcePath, destinationPath = parts[0], parts[1]
	if sourcePath == "" || destinationPath == "" {
		return "", "", fmt.Errorf("invalid volume [%s], path can't be empty", volume)
	}
	return sourcePath, destinationPath, nil
}

// umountVolume 卸载数据卷
func umountVolume(mntPath, containerPath string) {
	containerPathInHost := path.Join(mntPath, containerPath)
	cmd := exec.Command("umount", containerPathInHost)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("umount volume %s failed, err: %v", containerPathInHost, err)
	}
} 
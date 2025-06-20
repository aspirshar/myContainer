package utils

import (
	"fmt"
	"github.com/aspirshar/myContainer/config"
)

// 容器相关目录格式
var (
	lowerDirFormat  = "%s%s/lower"
	upperDirFormat  = "%s%s/upper"
	workDirFormat   = "%s%s/work"
	mergedDirFormat = "%s%s/merged"
	overlayFSFormat = "lowerdir=%s,upperdir=%s,workdir=%s"
)

func GetRoot(containerID string) string { 
	return config.RootPath + containerID 
}

func GetImage(imageName string) string { 
	return fmt.Sprintf("%s%s.tar", config.ImagesPath, imageName) 
}

func GetLower(containerID string) string {
	return fmt.Sprintf(lowerDirFormat, config.RootPath, containerID)
}

func GetUpper(containerID string) string {
	return fmt.Sprintf(upperDirFormat, config.RootPath, containerID)
}

func GetWorker(containerID string) string {
	return fmt.Sprintf(workDirFormat, config.RootPath, containerID)
}

func GetMerged(containerID string) string { 
	return fmt.Sprintf(mergedDirFormat, config.RootPath, containerID) 
}

func GetOverlayFSDirs(lower, upper, worker string) string {
	return fmt.Sprintf(overlayFSFormat, lower, upper, worker)
}
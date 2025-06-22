package container

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"github.com/aspirshar/myContainer/constant"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func RecordContainerInfo(containerPID int, commandArray []string, containerName, containerId, volume, networkName, ip string, portMapping []string) (*Info, error) {
	// 如果未指定容器名，则使用随机生成的containerID
	if containerName == "" {
		containerName = containerId
	}
	command := strings.Join(commandArray, "")
	containerInfo := &Info{
		Pid:         strconv.Itoa(containerPID),
		Id:          containerId,
		Name:        containerName,
		Command:     command,
		CreatedTime: time.Now().Format("2006-01-02 15:04:05"),
		Status:      RUNNING,
		Volume:      volume,
		NetworkName: networkName,
		PortMapping: portMapping,
		IP:          ip,
	}

	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		return containerInfo, errors.WithMessage(err, "container info marshal failed")
	}
	jsonStr := string(jsonBytes)
	// 拼接出存储容器信息文件的路径，如果目录不存在则级联创建
	dirPath := fmt.Sprintf(InfoLocFormat, containerId)
	if err = os.MkdirAll(dirPath, constant.Perm0622); err != nil {
		return containerInfo, errors.WithMessagef(err, "mkdir %s failed", dirPath)
	}
	// 将容器信息写入文件
	fileName := path.Join(dirPath, ConfigName)
	file, err := os.Create(fileName)
	if err != nil {
		return containerInfo, errors.WithMessagef(err, "create file %s failed", fileName)
	}
	defer file.Close()
	if _, err = file.WriteString(jsonStr); err != nil {
		return containerInfo, errors.WithMessagef(err, "write container info to  file %s failed", fileName)
	}
	return containerInfo, nil
}

// GetContainerInfoByName 通过容器名称获取容器信息
func GetContainerInfoByName(containerName string) (*Info, error) {
	// 读取存放容器信息目录下的所有文件
	files, err := os.ReadDir(InfoLoc)
	if err != nil {
		return nil, errors.WithMessagef(err, "read dir %s failed", InfoLoc)
	}
	
	// 遍历所有容器目录，查找匹配的容器名称
	for _, file := range files {
		configFileDir := fmt.Sprintf(InfoLocFormat, file.Name())
		configFilePath := path.Join(configFileDir, ConfigName)
		
		// 读取容器配置文件
		content, err := os.ReadFile(configFilePath)
		if err != nil {
			continue // 跳过无法读取的配置文件
		}
		
		info := new(Info)
		if err = json.Unmarshal(content, info); err != nil {
			continue // 跳过无法解析的配置文件
		}
		
		// 如果找到匹配的容器名称，返回容器信息
		if info.Name == containerName {
			return info, nil
		}
	}
	
	return nil, errors.Errorf("container with name '%s' not found", containerName)
}

// GetContainerIDByName 通过容器名称获取容器ID
func GetContainerIDByName(containerName string) (string, error) {
	info, err := GetContainerInfoByName(containerName)
	if err != nil {
		return "", err
	}
	return info.Id, nil
}

func DeleteContainerInfo(containerID string) error {
	dirPath := fmt.Sprintf(InfoLocFormat, containerID)
	if err := os.RemoveAll(dirPath); err != nil {
		return errors.WithMessagef(err, "remove dir %s failed", dirPath)
	}
	return nil
}

func GenerateContainerID() string {
	return randStringBytes(IDLength)
}

func randStringBytes(n int) string {
	letterBytes := "1234567890"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func GetLogfile(containerId string) string {
	return fmt.Sprintf(LogFile, containerId)
}
package main

import (
	"encoding/json"
	"fmt"
	"github.com/aspirshar/myContainer/constant"
	"os"
	"path"
	"strconv"
	"syscall"
	"text/tabwriter"

	"github.com/aspirshar/myContainer/container"

	log "github.com/sirupsen/logrus"
)

func ListContainers() {
	// 读取存放容器信息目录下的所有文件
	files, err := os.ReadDir(container.InfoLoc)
	if err != nil {
		log.Errorf("read dir %s error %v", container.InfoLoc, err)
		return
	}
	containers := make([]*container.Info, 0, len(files))
	for _, file := range files {
		tmpContainer, err := getContainerInfo(file)
		if err != nil {
			log.Errorf("get container info error %v", err)
			continue
		}
		containers = append(containers, tmpContainer)
	}
	// 使用tabwriter.NewWriter在控制台打印出容器信息
	// tabwriter 是引用的text/tabwriter类库，用于在控制台打印对齐的表格
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	_, err = fmt.Fprint(w, "ID\tNAME\tPID\tIP\tSTATUS\tCOMMAND\tCREATED\n")
	if err != nil {
		log.Errorf("Fprint error %v", err)
	}
	for _, item := range containers {
		_, err = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id,
			item.Name,
			item.Pid,
			item.IP,
			item.Status,
			item.Command,
			item.CreatedTime)
		if err != nil {
			log.Errorf("Fprint error %v", err)
		}
	}
	if err = w.Flush(); err != nil {
		log.Errorf("Flush error %v", err)
	}
}

func getContainerInfo(file os.DirEntry) (*container.Info, error) {
	configFileDir := fmt.Sprintf(container.InfoLocFormat, file.Name())
	configFilePath := path.Join(configFileDir, container.ConfigName)
	// 读取容器配置文件
	content, err := os.ReadFile(configFilePath)
	if err != nil {
		log.Errorf("read file %s error %v", configFilePath, err)
		return nil, err
	}
	info := new(container.Info)
	if err = json.Unmarshal(content, info); err != nil {
		log.Errorf("json unmarshal error %v", err)
		return nil, err
	}

	// 如果容器状态为 RUNNING，则检查进程是否存在
	if info.Status == container.RUNNING {
		pid, err := strconv.Atoi(info.Pid)
		if err != nil {
			log.Errorf("convert pid from string to int error %v", err)
			return nil, err
		}
		// 向进程发送 signal 0，如果返回 ESRCH 错误，则说明进程不存在
		if err := syscall.Kill(pid, 0); err == syscall.ESRCH {
			log.Infof("container %s process %d not exist, update status to stopped", info.Id, pid)
			info.Status = container.STOP
			info.Pid = " "
			// 将更新后的信息写回 config.json
			newContent, err := json.Marshal(info)
			if err != nil {
				log.Errorf("json marshal error %v", err)
				return nil, err
			}
			if err := os.WriteFile(configFilePath, newContent, constant.Perm0622); err != nil {
				log.Errorf("write file %s error %v", configFilePath, err)
				return nil, err
			}
		}
	}

	return info, nil
}
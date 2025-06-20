package config

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// 容器相关目录
var (
	// 项目根目录
	RootDir string

	// 镜像和容器层目录
	ImagePath  string
	RootPath   string
	ImagesPath string // 用于存储容器镜像的路径
)

// Init 初始化配置
func Init() error {
	var err error

	// 获取项目根目录
	RootDir, err = filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return err
	}

	// 设置镜像和容器层目录
	ImagePath = filepath.Join(RootDir, "images") + "/"
	RootPath = filepath.Join(RootDir, "overlay2") + "/"
	ImagesPath = filepath.Join(RootDir, "images") + "/"  // 使用与ImagePath相同的路径

	// 创建必要的目录
	if err := ensureDirectories(); err != nil {
		return err
	}

	return nil
}

// ensureDirectories 检查并创建必要的目录结构
func ensureDirectories() error {
	// 创建镜像目录
	if err := os.MkdirAll(ImagePath, 0755); err != nil {
		log.Errorf("Failed to create image directory: %v", err)
		return err
	}

	// 创建容器层目录
	if err := os.MkdirAll(RootPath, 0755); err != nil {
		log.Errorf("Failed to create overlay2 directory: %v", err)
		return err
	}

	// 创建ImagesPath目录
	if err := os.MkdirAll(ImagesPath, 0755); err != nil {
		log.Errorf("Failed to create images path directory: %v", err)
		return err
	}

	log.Infof("Container directories created successfully: %s, %s, %s", ImagePath, RootPath, ImagesPath)
	return nil
}
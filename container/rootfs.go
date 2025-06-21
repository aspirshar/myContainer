package container

import (
	"encoding/json"
	"github.com/aspirshar/myContainer/utils"
	"os"
	"os/exec"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// OCI 镜像相关结构
type OCIManifest struct {
	Config  string   `json:"Config"`
	RepoTags []string `json:"RepoTags"`
	Layers  []string `json:"Layers"`
}

// NewWorkSpace Create an Overlay2 filesystem as container root workspace
/*
1）创建lower层
2）创建upper、worker层
3）创建merged目录并挂载overlayFS
4）如果有指定volume则挂载volume
*/
func NewWorkSpace(containerID, imageName, volume string) {
	createLower(containerID, imageName)
	createDirs(containerID)
	mountOverlayFS(containerID)

	// 如果指定了volume则还需要mount volume
	if volume != "" {
		mntPath := utils.GetMerged(containerID)
		hostPath, containerPath, err := volumeExtract(volume)
		if err != nil {
			log.Errorf("extract volume failed, maybe volume parameter input is not correct, detail:%v", err)
			return
		}
		mountVolume(mntPath, hostPath, containerPath)
	}
}

// DeleteWorkSpace Delete the UFS filesystem while container exit
/*
和创建相反
1）有volume则卸载volume
2）卸载并移除merged目录
3）卸载并移除upper、worker层
*/
func DeleteWorkSpace(containerID, volume string) {
	// 如果指定了volume则需要umount volume
	// NOTE: 一定要要先 umount volume ，然后再删除目录，否则由于 bind mount 存在，删除临时目录会导致 volume 目录中的数据丢失。
	if volume != "" {
		_, containerPath, err := volumeExtract(volume)
		if err != nil {
			log.Errorf("extract volume failed，maybe volume parameter input is not correct，detail:%v", err)
			return
		}
		mntPath := utils.GetMerged(containerID)
		umountVolume(mntPath, containerPath)
	}

	umountOverlayFS(containerID)
	deleteDirs(containerID)
}

// createLower 根据 containerID, imageName 准备 lower 层目录
func createLower(containerID, imageName string) {
	// 根据 containerID 拼接出 lower 目录
	// 根据 imageName 找到镜像 tar，并解压到 lower 目录中
	lowerPath := utils.GetLower(containerID)
	imagePath := utils.GetImage(imageName)
	log.Infof("lower:%s image.tar:%s", lowerPath, imagePath)
	// 检查目录是否已经存在
	exist, err := utils.PathExists(lowerPath)
	if err != nil {
		log.Infof("Fail to judge whether dir %s exists. %v", lowerPath, err)
	}
	// 不存在则创建目录并将image.tar解压到lower文件夹中
	if !exist {
		log.Infof("Creating lower directory at %s", lowerPath)
		if err = os.MkdirAll(lowerPath, 0777); err != nil {
			log.Errorf("Failed to create lower directory %s: %v", lowerPath, err)
			return
		}

		// 检查镜像文件是否存在
		if _, err := os.Stat(imagePath); err != nil {
			log.Errorf("Image file %s not found: %v", imagePath, err)
			return
		}

		// 创建临时目录来提取 OCI 镜像
		tempDir, err := os.MkdirTemp("", "oci-extract-*")
		if err != nil {
			log.Errorf("Failed to create temp directory: %v", err)
			return
		}
		defer os.RemoveAll(tempDir)

		// 提取整个镜像到临时目录
		log.Infof("Extracting OCI image %s to temp directory %s", imagePath, tempDir)
		cmd := exec.Command("tar", "-xf", imagePath, "-C", tempDir)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Errorf("Failed to extract OCI image %s: %v\nCommand output: %s", 
				imagePath, err, string(output))
			return
		}

		// 读取并解析 manifest.json
		manifestPath := filepath.Join(tempDir, "manifest.json")
		manifestData, err := os.ReadFile(manifestPath)
		if err != nil {
			log.Errorf("Failed to read manifest.json: %v", err)
			return
		}

		var manifests []OCIManifest
		if err := json.Unmarshal(manifestData, &manifests); err != nil {
			log.Errorf("Failed to parse manifest.json: %v", err)
			return
		}

		if len(manifests) == 0 {
			log.Errorf("No manifests found in image")
			return
		}

		manifest := manifests[0]
		log.Infof("Found %d layers in image", len(manifest.Layers))

		// 提取并解压所有层
		for i, layer := range manifest.Layers {
			layerPath := filepath.Join(tempDir, layer)
			log.Infof("Extracting layer %d: %s", i+1, layer)
			
			// 解压层到 lower 目录
			cmd := exec.Command("tar", "-xf", layerPath, "-C", lowerPath)
			output, err := cmd.CombinedOutput()
			if err != nil {
				log.Errorf("Failed to extract layer %s: %v\nCommand output: %s", 
					layer, err, string(output))
				return
			}
		}

		log.Infof("Successfully extracted image to %s", lowerPath)
	}
}

// createDirs 创建overlayfs需要的的merged、upper、worker目录
func createDirs(containerID string) {
	dirs := []string{
		utils.GetMerged(containerID),
		utils.GetUpper(containerID),
		utils.GetWorker(containerID),
	}

	for _, dir := range dirs {
		if err := os.Mkdir(dir, 0777); err != nil {
			log.Errorf("mkdir dir %s error. %v", dir, err)
		}
	}
}

// mountOverlayFS 挂载overlayfs
func mountOverlayFS(containerID string) {
	// 拼接参数
	// e.g. lowerdir=/root/busybox,upperdir=/root/upper,workdir=/root/work
	dirs := utils.GetOverlayFSDirs(utils.GetLower(containerID), utils.GetUpper(containerID), utils.GetWorker(containerID))
	mergedPath := utils.GetMerged(containerID)
	//完整命令：mount -t overlay overlay -o lowerdir=/root/{containerID}/lower,upperdir=/root/{containerID}/upper,workdir=/root/{containerID}/work /root/{containerID}/merged
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, mergedPath)
	log.Infof("mount overlayfs: [%s]", cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}
}

func umountOverlayFS(containerID string) {
	mntPath := utils.GetMerged(containerID)
	cmd := exec.Command("umount", mntPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Infof("umountOverlayFS,cmd:%v", cmd.String())
	if err := cmd.Run(); err != nil {
		log.Errorf("%v", err)
	}
}

func deleteDirs(containerID string) {
	dirs := []string{
		utils.GetMerged(containerID),
		utils.GetUpper(containerID),
		utils.GetWorker(containerID),
		utils.GetLower(containerID),
		utils.GetRoot(containerID), // root 目录也要删除
	}

	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil {
			log.Errorf("Remove dir %s error %v", dir, err)
		}
	}
}
package fs2

import (
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/aspirshar/myContainer/constant"
	"github.com/pkg/errors"
)

// getCgroupPath 找到cgroup在文件系统中的绝对路径
/*
将根目录和cgroup名称拼接成一个路径。
如果指定了自动创建，就先检测一下是否存在，如果对应的目录不存在，则说明cgroup不存在，这里就给创建一个
*/
func getCgroupPath(cgroupPath string, autoCreate bool) (string, error) {
	// 不需要自动创建就直接返回
	cgroupRoot := UnifiedMountpoint
	absPath := path.Join(cgroupRoot, cgroupPath)
	if !autoCreate {
		return absPath, nil
	}
	// 指定自动创建时才判断是否存在
	_, err := os.Stat(absPath)
	if err != nil && os.IsNotExist(err) {
		err = os.Mkdir(absPath, constant.Perm0755)
		if err != nil {
			return absPath, err
		}
		// 在 cgroup v2 中，需要启用 CPU 控制器
		if err = enableCPUController(absPath); err != nil {
			return absPath, errors.Wrap(err, "enable cpu controller")
		}
		return absPath, nil
	}
	return absPath, errors.Wrap(err, "create cgroup")
}

// enableCPUController 启用 CPU 控制器
func enableCPUController(cgroupPath string) error {
	// 在 cgroup v2 中，通过写入 cgroup.subtree_control 来启用控制器
	// 格式为 "+cpu" 表示启用 CPU 控制器
	subtreeControlPath := path.Join(cgroupPath, "cgroup.subtree_control")
	
	// 首先检查父级 cgroup 是否已启用 CPU 和 cpuset 控制器
	parentSubtreeControlPath := path.Join(UnifiedMountpoint, "cgroup.subtree_control")
	parentContent, err := os.ReadFile(parentSubtreeControlPath)
	if err != nil {
		return errors.Wrap(err, "read parent subtree_control")
	}
	
	// 如果父级没有启用 CPU 控制器，先启用它
	if !contains(string(parentContent), "cpu") {
		if err = os.WriteFile(parentSubtreeControlPath, []byte("+cpu"), constant.Perm0644); err != nil {
			return errors.Wrap(err, "enable cpu controller in parent")
		}
	}
	
	// 如果父级没有启用 cpuset 控制器，先启用它
	if !contains(string(parentContent), "cpuset") {
		if err = os.WriteFile(parentSubtreeControlPath, []byte("+cpuset"), constant.Perm0644); err != nil {
			return errors.Wrap(err, "enable cpuset controller in parent")
		}
	}
	
	// 启用当前 cgroup 的 CPU 和 cpuset 控制器
	return os.WriteFile(subtreeControlPath, []byte("+cpu +cpuset"), constant.Perm0644)
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		containsSubstring(s, substr))))
}

// containsSubstring 检查字符串中间是否包含子字符串
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func applyCgroup(pid int, cgroupPath string) error {
	subCgroupPath, err := getCgroupPath(cgroupPath, true)
	if err != nil {
		return errors.Wrapf(err, "get cgroup %s", cgroupPath)
	}
	if err = os.WriteFile(path.Join(subCgroupPath, "cgroup.procs"), []byte(strconv.Itoa(pid)),
		constant.Perm0644); err != nil {
		return fmt.Errorf("set cgroup proc fail %v", err)
	}
	return nil
}
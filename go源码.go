package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

// 配置文件结构
type Config struct {
	Websites []string `json:"websites"`
}

// 主函数
func main() {
	// 读取配置文件
	config, err := loadConfig("config.json")
	if err != nil {
		fmt.Printf("读取配置文件失败: %v\n", err)
		fmt.Println("请确保当前目录下存在config.json文件")
		fmt.Println("按任意键退出...")
		fmt.Scanln()
		return
	}

	// 查找Chrome浏览器快捷方式
	chromeShortcuts, err := findChromeShortcuts()
	if err != nil {
		fmt.Printf("查找Chrome浏览器快捷方式失败: %v\n", err)
		fmt.Println("请确保当前文件夹中有Chrome浏览器的快捷方式")
		fmt.Println("按任意键退出...")
		fmt.Scanln()
		return
	}

	if len(chromeShortcuts) == 0 {
		fmt.Println("未找到Chrome浏览器快捷方式")
		fmt.Println("按任意键退出...")
		fmt.Scanln()
		return
	}

	// 打开第一个Chrome快捷方式
	fmt.Printf("找到Chrome浏览器快捷方式: %s\n", filepath.Base(chromeShortcuts[0]))
	fmt.Println("正在打开第一个Chrome浏览器...")

	openChrome(chromeShortcuts[0])

	// 等待浏览器启动
	time.Sleep(1 * time.Second)

	// 打开配置的网站
	if len(config.Websites) > 0 {
		fmt.Printf("正在打开 %d 个配置的网站...\n", len(config.Websites))
		openWebsites(config.Websites)
	} else {
		fmt.Println("没有配置要打开的网站")
	}

	// Windows下检测浏览器进程
	if runtime.GOOS == "windows" {
		fmt.Println("正在监控Chrome浏览器进程...")
		waitForChromeToEnd()
		fmt.Println("第一个Chrome浏览器已关闭")

		// 检查是否存在第二个Chrome快捷方式
		if len(chromeShortcuts) > 1 {
			fmt.Printf("找到第二个Chrome浏览器快捷方式: %s\n", filepath.Base(chromeShortcuts[1]))
			fmt.Println("正在打开第二个Chrome浏览器...")

			openChrome(chromeShortcuts[1])

			// 等待第二个浏览器启动
			time.Sleep(1 * time.Second)

			// 为第二个浏览器打开相同的网站
			if len(config.Websites) > 0 {
				fmt.Printf("在第二个浏览器中打开 %d 个配置的网站...\n", len(config.Websites))
				openWebsites(config.Websites)
			}

			// 等待第二个浏览器关闭
			fmt.Println("正在监控第二个Chrome浏览器进程...")
			waitForChromeToEnd()
			fmt.Println("第二个Chrome浏览器已关闭")
		} else {
			fmt.Println("未找到第二个Chrome浏览器快捷方式")
		}
	} else {
		// 非Windows系统，等待用户退出
		fmt.Println("程序已完成，按Enter键退出...")
		fmt.Scanln()
	}

	fmt.Println("程序执行完毕，按Enter键退出...")
	fmt.Scanln()
}

// 打开Chrome浏览器
func openChrome(chromePath string) error {
	// 根据操作系统选择适当的打开方式
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", chromePath)
	case "darwin":
		cmd = exec.Command("open", chromePath)
	case "linux":
		cmd = exec.Command("xdg-open", chromePath)
	}

	// 设置进程属性，以便获取进程ID
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
		}
	}

	return cmd.Start()
}

// 查找所有Chrome浏览器快捷方式
func findChromeShortcuts() ([]string, error) {
	// 获取当前目录
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	var chromeShortcuts []string

	// 查找所有.lnk和.url文件
	err = filepath.Walk(currentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 只处理当前目录下的文件（不包括子目录）
		if filepath.Dir(path) != currentDir {
			return nil
		}

		// 检查文件扩展名和文件名
		ext := strings.ToLower(filepath.Ext(path))
		fileName := strings.ToLower(filepath.Base(path))

		if (ext == ".lnk" || ext == ".url") && strings.Contains(fileName, "chrome") {
			chromeShortcuts = append(chromeShortcuts, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return chromeShortcuts, nil
}

// 打开配置的网站
func openWebsites(websites []string) {
	for i, website := range websites {
		fmt.Printf("打开网站 %d/%d: %s\n", i+1, len(websites), website)

		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			cmd = exec.Command("cmd", "/c", "start", "chrome", website)
		case "darwin":
			cmd = exec.Command("open", "-a", "Google Chrome", website)
		case "linux":
			cmd = exec.Command("google-chrome", website)
		}

		err := cmd.Start()
		if err != nil {
			fmt.Printf("打开网站失败: %v\n", err)
		}

		// 给浏览器一些时间来打开网站
		time.Sleep(1 * time.Second)
	}
}

// 加载配置文件
func loadConfig(configPath string) (*Config, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		// 如果配置文件不存在，创建一个示例配置
		if os.IsNotExist(err) {
			exampleConfig := &Config{
				Websites: []string{
					"https://www.google.com",
				},
			}

			exampleData, _ := json.MarshalIndent(exampleConfig, "", "  ")
			ioutil.WriteFile(configPath, exampleData, 0644)

			return nil, fmt.Errorf("配置文件不存在，已创建示例配置文件 %s，请修改后重新运行", configPath)
		}
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// 等待Chrome浏览器进程结束（Windows特定）
func waitForChromeToEnd() {
	if runtime.GOOS != "windows" {
		return
	}

	// 检测Chrome进程是否存在
	checkChromeExists := func() bool {
		cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq chrome.exe", "/NH")
		output, err := cmd.Output()
		if err != nil {
			return false
		}

		return strings.Contains(string(output), "chrome.exe")
	}

	// 等待Chrome进程出现
	initialWait := 0
	for !checkChromeExists() {
		initialWait++
		if initialWait > 10 { // 最多等待10次（10秒）
			fmt.Println("未检测到Chrome浏览器进程启动")
			return
		}
		time.Sleep(1 * time.Second)
	}

	fmt.Println("检测到Chrome浏览器进程已启动，等待其关闭...")

	// 循环检查进程是否仍然存在
	for checkChromeExists() {
		time.Sleep(2 * time.Second) // 每2秒检查一次
	}
}

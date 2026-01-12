package core

// FasterEdge - 对称、可靠、安全的多场景边缘计算框架
// https://github.com/FasterEdge
// DontCrack开源鸿蒙(OpenHarmony)Linux内核专用版
// https://github.com/FasterEdge/DontCrack4OpenHarmonyLinuxKernelSide
// tyza66
// https://github.com/tyza66
import (
	"DontCrash/config"
	"log"
)

const logo = `
 ________  ________           _______          ________  ___  ___     
|\   ___ \|\   ____\         /  ___  \        |\   __  \|\  \|\  \    
\ \  \_|\ \ \  \___|        /__/|_/  /|       \ \  \|\  \ \  \\\  \   
 \ \  \ \\ \ \  \           |__|//  / /        \ \  \\\  \ \   __  \  
  \ \  \_\\ \ \  \____          /  /_/__        \ \  \\\  \ \  \ \  \ 
   \ \_______\ \_______\       |\________\       \ \_______\ \__\ \__\
    \|_______|\|_______|        \|_______|        \|_______|\|__|\|__|
`

func Start(cfg config.Config) {
	// 检查配置信息是否合法
	err := config.CheckConfig(cfg)
	if err != nil {
		log.Fatalf("配置检查失败: %v\n", err)
		return
	}
	// 输出启动Logo和版本等欢迎信息
	log.Printf(logo)
	log.Printf("DontCrack for OpenHarmony v%s 启动中...\n", cfg.Version)
	log.Printf("管理进程正在管理的程序: %s\n", cfg.Path)

	// 启动管理器核心逻辑

}

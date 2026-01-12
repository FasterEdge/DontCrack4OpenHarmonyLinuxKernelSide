package core

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
	err := config.CheckConfig(cfg)
	if err != nil {
		log.Fatalf("配置检查失败: %v\n", err)
		return
	}
	log.Printf(logo)

	log.Printf("DontCrack for OpenHarmony v%s 启动中...\n", cfg.Version)

}

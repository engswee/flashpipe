package logger

import (
	"fmt"
	"os"
)

func Error(a ...any) {
	fmt.Print("[\x1b[31mERROR\x1b[m] 🛑 ")
	fmt.Println(a...)
	os.Exit(1)
}

func Info(a ...any) {
	fmt.Print("[\x1b[32mINFO\x1b[m] ")
	fmt.Println(a...)
}

func Warn(a ...any) {
	fmt.Print("[\x1b[33mWARN\x1b[m] ⚠️ ")
	fmt.Println(a...)
}

func Debug(a ...any) {
	fmt.Print("[\x1b[34mDEBUG\x1b[m] ")
	fmt.Println(a...)
}

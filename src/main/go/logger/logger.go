package logger

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
)

func Error(a ...any) {
	fmt.Print("[\x1b[31mERROR\x1b[m] 🛑 ")
	fmt.Println(a...)
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
	if viper.GetBool("debug") {
		fmt.Print("[\x1b[34mDEBUG\x1b[m] ")
		fmt.Println(a...)
	}
}

func ExitIfError(err error) {
	if err != nil {
		Error(err)
		os.Exit(1)
	}
}

func ExitIfErrorWithMsg(err error, msg string) {
	if err != nil {
		Error(msg)
		os.Exit(1)
	}
}

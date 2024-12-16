package t_error

import (
	"fmt"
	"log"
	"runtime/debug"
)

func LogWarn(err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
}
func LogErr(err error){
	if err != nil {
		debug.PrintStack()
		log.Fatal(err.Error())
	}
}


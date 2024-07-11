package t_error

import (
	"fmt"
	"log"
)

func LogWarn(err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
}
func LogErr(err error){
	if err != nil {
		log.Panic(err.Error())
	}
}


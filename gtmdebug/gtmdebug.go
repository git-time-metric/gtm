
package gtmdebug

import (
	"log"
	"fmt"
)

func Debugf(module string, format string, v ...interface{}) {
	if false {
		log.Printf(fmt.Sprintf("[%s] ", module)+format, v...)
	}
}

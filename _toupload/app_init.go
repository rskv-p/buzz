// file: buzz/pkg/x_init/app_init.go

package x_init

import (
	"github.com/rskv-p/buzz/pkg/x_cst"
	"github.com/rskv-p/buzz/pkg/x_log"
)

func Init() {
	x_cst.Init()
	x_log.Init()
}

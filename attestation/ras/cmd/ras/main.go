/*
For test purpose, do the following steps:
1. open two terminal, one to run ras and another to run rac.
2. in terminal A, run command: go run ras/cmd/main.go
3. in terminal B, run command: go run rac/cmd/main.go
*/
package main

import (
	"gitee.com/openeuler/kunpengsecl/attestation/ras/clientapi"
	"gitee.com/openeuler/kunpengsecl/attestation/ras/config"
	"gitee.com/openeuler/kunpengsecl/attestation/ras/entity"
	"gitee.com/openeuler/kunpengsecl/attestation/ras/restapi"
	"gitee.com/openeuler/kunpengsecl/attestation/ras/trustmgr"
	"github.com/spf13/pflag"
)

type testValidator struct {
}

func (tv *testValidator) Validate(report *entity.Report) error {
	return nil
}

func init() {
	config.InitRasFlags()
}

func main() {
	pflag.Parse()
	cfg := config.GetDefault()

	// TODO: Wait for completing of validator
	trustmgr.SetValidator(&testValidator{})
	go clientapi.StartServer(cfg.GetPort())
	restapi.StartServer(cfg.GetRestPort())
}

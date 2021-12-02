package dao

import (
	"gitee.com/openeuler/kunpengsecl/attestation/ras/entity"
)

/*
	DAO is an interface for processing data in database
*/
type DAO interface {
	SaveReport(report *entity.Report) error
	RegisterClient(clientInfo *entity.ClientInfo, ic []byte) (int64, error)
	UnRegisterClient(clientID int64) error
	SaveBaseValue(clientID int64, meaInfo *entity.MeasurementInfo) error
	SelectAllRegisteredClientIds() ([]int64, error)
	SelectAllClientIds() ([]int64, error)
	SelectReportsById(clientId int64) ([]*entity.Report, error)
	SelectLatestReportById(clientId int64) (*entity.Report, error)
	SelectBaseValueById(clientId int64) (*entity.MeasurementInfo, error)
	SelectClientById(clientId int64) (*entity.RegisterClient, error)
	Destroy()
	SelectAllClientInfobyId(clientId int64) (map[string]string, error)
	SelectClientInfobyId(clientId int64, infoNames []string) (map[string]string, error)
	UpdateRegisterStatusById(clientId int64, isDeleted bool) error
}

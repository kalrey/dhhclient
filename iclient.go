package dhhclient



type IDHHClient interface {
	InitClient()*DHHError
	GenTaskDataFile(taskId string, startDate string, endDate string) *DHHError
	GenTaskDataExcel(taskData *DHHTaskDataResponseJson)
//	WriteExcel(sheetName string, items []*DHHTaskDataItem) (io.Reader, error)
	QueryTaskDataBatch(taskId string, startDate string, endDate string) (*DHHTaskDataResponseJson, *DHHError)
	QueryTaskData(taskId string, startDate string, endDate string, pageNum int, pageSize int) (*DHHTaskDataResponseJson, *DHHError)
	QueryTaskMaterial(taskId string, pageNum int, pageSize int) (int, *DHHError)
	QueryMaterialInfo(taskId string, materialId string, pageNum int, pageSize int, prefixFilter string) (string, int, int, *DHHError)
	ListAdSpaces(status int, pageNum int, pageSize int, prefix string) (string, int, int, *DHHError)
	AddAdspace(deliveryApp string, deliveryAddress string, deliveryAppFileName string, adName string) *DHHError
	BatchAdSpace(deliveryApp string, deliveryAddress string, deliveryAppFileName string, adNameFile string) *DHHError
}
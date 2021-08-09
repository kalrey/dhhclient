package dhhclient

import (
	"encoding/json"
)

const (
	DHH_ERROR_CODE_UNKNOWN          = 10000
	DHH_ERROR_CODE_HTTPERROR        = 10001
	DHH_ERROR_CODE_HTTPNOTOK        = 10002
	DHH_ERROR_CODE_HTTPRESP_READERR = 10003
	DHH_ERROR_CODE_NOTOKEN          = 10004
	DHH_ERROR_CODE_JSONUNMARSHAL    = 10005
	DHH_ERROR_CODE_UNLOGIN          = 10006
	DHH_ERROR_CODE_NOTSUCCESS       = 10007
	DHH_ERROR_CODE_OPENFILEFAILED   = 10008
	DHH_ERROR_CODE_IOERROR          = 10009

	DHH_API_ENDPOINT_LOGINROOT = "https://login.taobao.com"
	DHH_API_ENDPOINT_INITVIEWDATA = "https://login.taobao.com/member/login.jhtml"

	DHH_LOGIN_FORM_KEY_LOGINID          = "loginId"
	DHH_LOGIN_FORM_KEY_PASSWORD2        = "password2"
	DHH_LOGIN_FORM_KEY_KEEPLOGIN        = "keepLogin"
	DHH_LOGIN_FORM_KEY_UMIDGETSTATUSVAL = "umidGetStatusVal"

	DHH_LOGIN_QS_KEY_APPNAME  = "appName"
	DHH_LOGIN_QS_KEY_FROMSITE = "fromSite"

	DHH_API_ENDPOINT_ACCOUNT_CHECK = "https://login.taobao.com/newlogin/account/check.do"

	DHH_API_ENDPOINT_LOGIN = "https://login.taobao.com/newlogin/login.do"

	DHH_LOGIN_REDIRECTURL_SUCCESS_PREFIX = "https://i.taobao.com/my_taobao.htm"

	DHH_QUERY_TASKDATA_QS_KEY_TASKID   = "taskID"
	DHH_QUERY_TASKDATA_QS_KEY_APPID    = "appID"
	DHH_QUERY_TASKDATA_QS_KEY_CSRF     = "_csrf"
	DHH_QUERY_TASKDATA_QS_KEY_START    = "start"
	DHH_QUERY_TASKDATA_QS_KEY_END      = "end"
	DHH_QUERY_TASKDATA_QS_KEY_PAGENUM  = "pageNum"
	DHH_QUERY_TASKDATA_QS_KEY_PAGESIZE = "pageSize"

	DHH_API_ENDPOINT_QUERYTASKDATA = "https://dahanghai.taobao.com/polystar/api/v3/settlement_data/p02/page_show"

	DHH_API_ENDPOINT_GENTASKDATAFILE = "https://dahanghai.taobao.com/polystar/api/v3/settlement_data/p02/file_show/commands/generate"
	DHH_GEN_TASKDATA_QS_KEY_TASKID   = "taskID"
	DHH_GEN_TASKDATA_QS_KEY_APPID    = "appID"
	DHH_GEN_TASKDATA_QS_KEY_START    = "start"
	DHH_GEN_TASKDATA_QS_KEY_END      = "end"
	DHH_GEN_TASKDATA_FORM_KEY_CSRF   = "_csrf"

	DHH_API_ENDPOINT_QUERYMATERIAL     = "https://dahanghai.taobao.com/polystar/api/materials/list"
	DHH_QUERY_MATERIAL_FORM_APPID      = "appId"
	DHH_QUERY_MATERIAL_FORM_SOURCE     = "source"
	DHH_QUERY_MATERIAL_FORM_TASKTYPEID = "taskTypeId"
	DHH_QUERY_MATERIAL_FORM_STATUS     = "status"
	DHH_QUERY_MATERIAL_FORM_CSRF       = "_csrf"
	DHH_QUERY_MATERIAL_FORM_PAGE       = "page"
	DHH_QUERY_MATERIAL_FORM_PAGESIZE   = "pageSize"

	DHH_API_ENDPOINT_QUERYMATERIALINFO   = "https://dahanghai.taobao.com/polystar/api/materials/urls"
	DHH_QUERY_MATERIALINFO_QS_APPID      = "appId"
	DHH_QUERY_MATERIALINFO_QS_TASKTYPEID = "taskTypeId"
	DHH_QUERY_MATERIALINFO_QS_MATERIALID = "materialId"
	DHH_QUERY_MATERIALINFO_QS_PAGE       = "page"
	DHH_QUERY_MATERIALINFO_QS_PAGESIZE   = "pageSize"
	DHH_QUERY_MATERIALINFO_QS_CSRF       = "_csrf"

	DHH_API_ENDPOINT_APPLYADSPACE                   = "https://dahanghai.taobao.com/polystar/api/channel/mgmt/applyAdSpace"
	DHH_QUERY_APPLYADSPACE_FORM_CHANNELID           = "channelId"
	DHH_QUERY_APPLYADSPACE_FORM_ADNAME              = "name"
	DHH_QUERY_APPLYADSPACE_FORM_TYPE                = "type"
	DHH_QUERY_APPLYADSPACE_FORM_DELIVERYAPP         = "deliveryApp"
	DHH_QUERY_APPLYADSPACE_FORM_DELIVERYADDRESS     = "deliveryAddress"
	DHH_QUERY_APPLYADSPACE_FORM_EFFECTEXAMPLEFILEID = "effectExampleFileId"
	DHH_QUERY_APPLYADSPACE_FORM_EFFECTEXAMPLEURL    = "effectExampleUrl"
	DHH_QUERY_APPLYADSPACE_FORM_DELIVERYTYPE        = "deliveryType"
	DHH_QUERY_APPLYADSPACE_FORM_CSRF                = "_csrf"

	DHH_API_ENDPOINT_ADSPACES        = "https://dahanghai.taobao.com/polystar/api/channel/mgmt/adSpaces"
	DHH_QUERY_ADSPACES_FORM_STATUS   = "status"
	DHH_QUERY_ADSPACES_FORM_PAGE     = "page"
	DHH_QUERY_ADSPACES_FORM_PAGESIZE = "pageSize"
	DHH_QUERY_ADSPACES_FORM_CSRF     = "_csrf"

	DHH_API_ENDPOINT_BIZFILES = "https://dahanghai.taobao.com/polystar/api/biz_files"
)

type DHHLoginContentData struct {
	Redirect    bool   `json:"redirect"`
	RedirectUrl string `json:"redirectUrl"`
	TitleMsg    string `json:"titleMsg"`
	ResultCode  int    `json:"resultCode"`
}

type DHHLoginContent struct {
	Data    *DHHLoginContentData `json:"data"`
	Status  int                  `json:"status"`
	Success bool                 `json:"success"`
}

type DHHLoginResponseJson struct {
	Content		*DHHLoginContent `json:"content"`
	HasError	bool             `json:"hasError"`
	RgvFlag587	string 		  	`json:"rgv587_flag"`
	AuthUrl		string			`json:"url"`			
}

type DHHTaskDataItem struct {
	Date             int    `json:"date" Field:"日期"`
	AdSpaceName      string `json:"adSpaceName" Field:"广告位名称"`
	AdSpaceID        int    `json:"adSpaceID" Field:"广告位ID"`
	CustomOSName     string `json:"customOS_Name" Field:"操作系统"`
	ShowQualityScore string `json:"showQualityScore" Field:"质量分"`
	TotalAmount      int    `json:"totalAmount" Field:"总计促活"`
	TargetAmount     int    `json:"targetAmount" Field:"总计目标完成"`
	ShowAccountFee   string `json:"showAccountFee" Field:"预估佣金"`
}

type DHHTaskDataResponseJson struct {
	TotalAmount         int                `json:"totalAmount"`
	TargetAmount        int                `json:"targetAmount"`
	ShowTotalAccountFee float64            `json:"showTotalAccountFee"`
	Successful          bool               `json:"successful"`
	Message             string             `json:"message"`
	Total               int                `json:"total"`
	Page                int                `json:"page"`
	PageSize            int                `json:"pageSize"`
	List                []*DHHTaskDataItem `json:"list"`
	Code                int                `json:"code"`
}

type DHHTaskDataGenResponseJson struct {
	Successful bool   `json:"successful"`
	Code       int    `json:"code"`
	Message    string `json:"message"`
}

type TaskMaterialItemJson struct {
	SourceFileId int    `json:"sourceFileId"`
	TaskTypeId   int    `json:"taskTypeId"`
	MaterialId   int    `json:"materialId"`
	Start        string `json:"start"`  //"2020-08-26"
	End          string `json:"end"`    //"2020-09-30"
	Status       int    `json:"status"` //1为正常
}

type MaterialUrlItemJson struct {
	AdId        int    `json:"adId"`
	Adname      string `json:"adName"`
	ImgUrl      string `json:"imgUrl"`
	H5Url       string `json:"h5Url"`
	DeepLinkUrl string `json:"deepLinkUrl"`
}

type MaterialUrlResponseJson struct {
	Successful bool                   `json:"successful"`
	Code       int                    `json:"code"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"pageSize"`
	Total      int                    `json:"total"`
	List       []*MaterialUrlItemJson `json:"list"`
}

type TaskMaterialResponseJson struct {
	Successful bool                    `json:"successful"`
	Code       int                     `json:"code"`
	Page       int                     `json:"page"`
	PageSize   int                     `json:"pageSize"`
	Total      int                     `json:"total"`
	List       []*TaskMaterialItemJson `json:"list"`
}

type BizFileData struct {
	FileID             int    `json:"fileID"`
	Version            int    `json:"version"`
	FileProcessStateV1 string `json:"fileProcessStateV1"`
	Url                string `json:"url"`
}

type CommonItemResponseJson struct {
	Successful bool         `json:"successful"`
	Code       int          `json:"code"`
	Message    string       `json:"message"`
	Data       *BizFileData `json:"data"`
}

type CommonResponseJson struct {
	Successful bool             `json:"successful"`
	Code       int              `json:"code"`
	Message    string           `json:"message"`
	Data       *json.RawMessage `json:"data"`
}

type CommonDataJson struct {
	Successful bool               `json:"successful"`
	Code       int                `json:"code"`
	Page       int                `json:"page"`
	PageSize   int                `json:"pageSize"`
	Total      int                `json:"total"`
	List       []*json.RawMessage `json:"list"`
}

type AdSpaceJson struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Status int    `json:"status"`
}
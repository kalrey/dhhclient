package dhhclient

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kalrey/zlog"

	"github.com/tealeg/xlsx"
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
	Content  *DHHLoginContent `json:"content"`
	HasError bool             `json:"hasError"`
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

type DHHError struct {
	errCode int
	verbose string
}

func (this *DHHError) Error() string {
	return "Errcode(" + strconv.Itoa(this.errCode) + "), verbose: " + this.verbose
}

func (this *DHHError) Code() int {
	return this.errCode
}

func NewError(errCode int, verbose string) *DHHError {
	return &DHHError{errCode, verbose}
}

type DHHClient struct {
	csrf         string
	cookies      []*http.Cookie
	channelId    string
	ua           string
	initViewData map[string]interface{}
}

func NewClient(channelId string) *DHHClient {
	return &DHHClient{
		"",
		nil,
		channelId,
		"",
		nil}
}

func (this *DHHClient) getCsrfToken() (string, *DHHError) {
	resp, err := http.Get(DHH_CSRF_URL)

	if err != nil {
		return "", NewError(DHH_ERROR_CODE_HTTPERROR, "getCsrfToken failed, http get failed.")
	}

	if resp.StatusCode != http.StatusOK {
		return "", NewError(DHH_ERROR_CODE_HTTPNOTOK, "getCsrfToken failed,status code: "+resp.Status)
	}

	for _, v := range resp.Cookies() {
		if v.Name == "XSRF-TOKEN" {
			return v.Value, nil
		}
	}
	return "", NewError(DHH_ERROR_CODE_NOTOKEN, "getCsrfToken failed, cannot find XSRF-TOKEN")
}

func (this *DHHClient) setCookies(r *http.Request) {
	for _, v := range this.cookies {
		r.AddCookie(v)
	}
}

func (this *DHHClient) GenTaskDataFile(taskId string, startDate string, endDate string) *DHHError {
	qs := url.Values{
		DHH_GEN_TASKDATA_QS_KEY_TASKID: {taskId},
		DHH_GEN_TASKDATA_QS_KEY_APPID:  {"1"},
		DHH_GEN_TASKDATA_QS_KEY_START:  {startDate},
		DHH_GEN_TASKDATA_QS_KEY_END:    {endDate}}

	form := url.Values{
		DHH_GEN_TASKDATA_FORM_KEY_CSRF: {this.csrf}}

	url := DHH_API_ENDPOINT_GENTASKDATAFILE + "?" + qs.Encode()

	bodyReader := strings.NewReader(form.Encode())
	req, _ := http.NewRequest("POST", url, bodyReader)
	req.Header.Set("Connection", "keep-alive")
	this.setCookies(req)

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return NewError(DHH_ERROR_CODE_HTTPERROR, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return NewError(DHH_ERROR_CODE_HTTPNOTOK, "genTaskDataFile response status("+resp.Status+"), url: "+url)
	}
	data, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		return NewError(DHH_ERROR_CODE_HTTPRESP_READERR, err2.Error())
	}

	var respJson DHHTaskDataGenResponseJson

	err = json.Unmarshal(data, &respJson)

	if err != nil {
		return NewError(DHH_ERROR_CODE_JSONUNMARSHAL, err.Error())
	}

	if respJson.Successful || respJson.Code == 0 {
		return nil
	}
	return NewError(DHH_ERROR_CODE_UNKNOWN, "genTaskDataFile failed, verbose: "+respJson.Message)
}

func (this *DHHClient) GenTaskDataExcel(taskData *DHHTaskDataResponseJson) {

	reader, err := this.WriteExcel("手淘常规促活", taskData.List)

	buf, _ := ioutil.ReadAll(reader)

	println(buf, err)
}

func (this *DHHClient) WriteExcel(sheetName string, items []*DHHTaskDataItem) (io.Reader, error) {

	file := xlsx.NewFile()
	sheet, err := file.AddSheet(sheetName)

	if err != nil {
		return nil, err
	}

	first := items[0]

	t := reflect.TypeOf(first)

	cols := t.NumField()

	header := sheet.AddRow()
	for i := 0; i < cols; i++ {
		fieldType := t.Field(i)
		fieldTag := fieldType.Tag.Get("Field")
		header.AddCell().Value = fieldTag
	}

	for _, item := range items {
		v := reflect.ValueOf(item)
		for col := 0; col < cols; col++ {
			row := sheet.AddRow()
			row.AddCell().Value = v.Field(col).String()
		}
	}
	b := new(bytes.Buffer)

	err2 := file.Write(b)

	if err2 != nil {
		return nil, err2
	}
	return b, nil
}

func (this *DHHClient) QueryTaskDataBatch(taskId string, startDate string, endDate string) (*DHHTaskDataResponseJson, *DHHError) {
	pageNum := 1
	pageSize := 50
	result, err := this.QueryTaskData(taskId, startDate, endDate, pageNum, pageSize)

	if err != nil {
		return nil, err
	}

	//数据未产出
	if err == nil && result == nil {
		return nil, nil
	}

	totalPage := float32(result.Total) / float32(pageSize)

	for float32(pageNum) < totalPage {
		resp, err2 := this.QueryTaskData(taskId, startDate, endDate, pageNum+1, pageSize)

		if err2 != nil {
			break
		}

		result.List = append(result.List, resp.List...)

		pageNum = pageNum + 1
	}

	return result, nil
}

func (this *DHHClient) QueryTaskData(taskId string, startDate string, endDate string, pageNum int, pageSize int) (*DHHTaskDataResponseJson, *DHHError) {
	qs := url.Values{
		DHH_QUERY_TASKDATA_QS_KEY_TASKID:   {taskId},
		DHH_QUERY_TASKDATA_QS_KEY_APPID:    {"1"},
		DHH_QUERY_TASKDATA_QS_KEY_CSRF:     {this.csrf},
		DHH_QUERY_TASKDATA_QS_KEY_START:    {startDate},
		DHH_QUERY_TASKDATA_QS_KEY_END:      {endDate},
		DHH_QUERY_TASKDATA_QS_KEY_PAGENUM:  {strconv.Itoa(pageNum)},
		DHH_QUERY_TASKDATA_QS_KEY_PAGESIZE: {strconv.Itoa(pageSize)}}

	url := DHH_API_ENDPOINT_QUERYTASKDATA + "?" + qs.Encode()

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Set("Connection", "keep-alive")

	this.setCookies(req)

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, NewError(DHH_ERROR_CODE_HTTPERROR, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return nil, NewError(DHH_ERROR_CODE_HTTPNOTOK, "queryTaskData response status("+resp.Status+"), url: "+url)
	}
	data, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		return nil, NewError(DHH_ERROR_CODE_HTTPRESP_READERR, err2.Error())
	}

	var respJson DHHTaskDataResponseJson

	err = json.Unmarshal(data, &respJson)

	if err != nil {
		zlog.Logger.Error("QueryTaskData Response Unmarshal error, taskid:" + taskId + ", resp data: " + string(data))
		return nil, NewError(DHH_ERROR_CODE_JSONUNMARSHAL, err.Error())
	}

	if respJson.Successful {
		if respJson.Code == 0 && respJson.List != nil {
			return &respJson, nil
		} else if respJson.Code == 0 && respJson.List == nil {
			zlog.Logger.Info("QueryTaskData Query No data, taskid:" + taskId)
			return nil, nil
		}
	}

	if respJson.Code == 4003 {
		return nil, NewError(DHH_ERROR_CODE_UNLOGIN, "QueryTaskData Response unlogin, taskId: "+taskId+", code: "+strconv.Itoa(respJson.Code))
	}

	zlog.Logger.Error("queryTaskData Response error, taskid: " + taskId + ", resp data: " + string(data))
	return nil, NewError(DHH_ERROR_CODE_UNKNOWN, "queryTaskData Response error, taskid: "+taskId+", resp data: "+string(data))
}

func (this *DHHClient) getCommonLoginInfo() url.Values {
	commonLoginInfo := url.Values{
		"screenPixel":      {"1440*900"},
		"umidGetStatusVal": {"255"},
		"ua":               {"140#RosDzikfzzPEDzo225PuApSob1119NSvDpym194Jyz1m0Cfa4zW7H751DF7jRRZubAlVvodQaHjVnRThOf8+LFpblp1zz/UaXsynZbzx+2Koa6Jjzzrb22U3lp1xzWGUV2VjUQDa2Dc3V3pWdQ+V2bTgl3TDqx3ZNjLLMZ2aMXKkLIxrHf9MS+eBwbozrDyUC1DSerYucTMICwVZKB12LqSDBEUxrHnR7i3UQRhVQTiRTuvs+qhlZ+CQ3QBkhUDTkP9T7AuNWcjXeeuGsHKu34giH8TucSVjKvfM/7eFHJHCN/1rMbqr3rzawy5gqTHs06LWNMAniHyPGKq+szfa98sGVvyuLT7kw2ckOQF1buku/NUdDMT9adEe7+n6DTAGiWz1FtayBHqTGKQ4+KirrDnOW+Wt8K5WXlnGVdg9IxaFwQE706EO9I1LjwhaY312l23iRc0Gpyq5PD1N7t9UyIVZ+BqVdGSP/bz9n0mzWqc3RslDPC7aq+zzLiCZusD17+udKsnvRciRqIvqSWfNot4ukgeeBO0JASmdO51+Fh6S3Bq6QbAIPak2efKV9jmYK4R94Rn+ANbT4IMdQtIruyH47LWwg2I3aadwz2eJ9C/G6IEWY1skeZSTzhvWXueZuH/gUuACOxC0uDEsqTxJtiokK0QIlpb8ID8dmEtACALdccL3Q2XSojwaNPgZGDM2d6FjEDUoo3d9i3/xC/8P4FCEy4ExWeg2ZclTGTb+jy3zvUQJoek6Ym0FJuq5PAqnA3SB3ND48JD6qldBIT5eJEfZ4Dpn86In5emsCv8mCJCvOK7FbZR1PAPpdM8et1DaVbQkeJAXq8OWLSfnhPt9zzugqlkU5Y4kIejNWPLx+V8aL/9naKHWReX1Cdbr72A71vfBP9Exi+L6zuaD6QZYxnfoQpF8cby1WBnZhqDmHHDD/u8evS3L4UIBSYsWDo/vjJdr3gtCc64wLqg1sFeKmsbPc8INPOYsVrMSD8W9uLzxep0RTqnpyx6kLTyaj+K64pGUmPw01oDNac2hZLyi"},
		"navUserAgent":     {"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36"},
		"navPlatform":      {"MacIntel"}}

	return commonLoginInfo
}

func (this *DHHClient) MergaValues(dst url.Values, src url.Values) url.Values {
	for k, _ := range src {
		if len(dst.Get(k)) == 0 {
			dst.Set(k, src.Get(k))
		}
	}
	return dst
}

func (this *DHHClient) InitClient(ua string) *DHHError {
	this.initViewData, _, _ = this.LoadingTaobaoInitViewData()

	csrf, errToken := this.getCsrfToken()

	if errToken != nil {
		return errToken
	}

	this.csrf = csrf
	this.ua = ua
	return nil
}

func (this *DHHClient) AccountCheck(loginId string) *DHHError {

	formData := url.Values{
		DHH_LOGIN_FORM_KEY_LOGINID: {loginId}}

	this.MergaValues(formData, this.getCommonLoginInfo())

	qs := url.Values{
		DHH_LOGIN_QS_KEY_APPNAME:  {this.initViewData[DHH_LOGIN_QS_KEY_APPNAME].(string)},
		DHH_LOGIN_QS_KEY_FROMSITE: {strconv.FormatFloat(this.initViewData[DHH_LOGIN_QS_KEY_FROMSITE].(float64), 'f', -1, 64)}}

	path := DHH_API_ENDPOINT_ACCOUNT_CHECK + "?" + qs.Encode()
	resp, err := http.PostForm(path, formData)

	if err != nil {
		return NewError(DHH_ERROR_CODE_HTTPERROR, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return NewError(DHH_ERROR_CODE_HTTPNOTOK, "AccountCheck response status("+resp.Status+"), url: "+path)
	}
	data, err2 := ioutil.ReadAll(resp.Body)

	if err2 != nil {
		return NewError(DHH_ERROR_CODE_HTTPRESP_READERR, err2.Error())
	}

	var respJson DHHLoginResponseJson

	err = json.Unmarshal(data, &respJson)

	if err != nil {
		zlog.Logger.Error("AccountCheck Response Unmarshal error, resp data: " + string(data))
		return NewError(DHH_ERROR_CODE_JSONUNMARSHAL, "AccountCheck Response Unmarshal error, resp data: "+string(data))
	}

	if !respJson.HasError && respJson.Content != nil {
		if respJson.Content.Success && respJson.Content.Status == 0 && respJson.Content.Data != nil {
			if respJson.Content.Data.ResultCode == 100 {
				return nil
			}
		}
	}
	return NewError(DHH_ERROR_CODE_UNKNOWN, "AccountCheck Response error, resp data: "+string(data))
}

func (this *DHHClient) Login(loginId string, password2 string) *DHHError {

	formData := url.Values{
		DHH_LOGIN_FORM_KEY_LOGINID:          {loginId},
		DHH_LOGIN_FORM_KEY_PASSWORD2:        {password2},
		DHH_LOGIN_FORM_KEY_KEEPLOGIN:        {"false"},
		DHH_LOGIN_FORM_KEY_UMIDGETSTATUSVAL: {"255"}}

	this.MergaValues(formData, this.getCommonLoginInfo())

	for k, v := range this.initViewData {
		switch v.(type) {
		case string:
			formData.Add(k, v.(string))
			break
		case bool:
			formData.Add(k, strconv.FormatBool(v.(bool)))
			break
		case int:
			formData.Add(k, strconv.Itoa(v.(int)))
			break
		case float64:
			formData.Add(k, strconv.FormatFloat(v.(float64), 'f', -1, 64))
			break
		}
	}

	qs := url.Values{
		DHH_LOGIN_QS_KEY_APPNAME:  {this.initViewData[DHH_LOGIN_QS_KEY_APPNAME].(string)},
		DHH_LOGIN_QS_KEY_FROMSITE: {strconv.FormatFloat(this.initViewData[DHH_LOGIN_QS_KEY_FROMSITE].(float64), 'f', -1, 64)}}

	path := DHH_API_ENDPOINT_LOGIN + "?" + qs.Encode()
	resp, err := http.PostForm(path, formData)
	if err != nil {
		return NewError(DHH_ERROR_CODE_HTTPERROR, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return NewError(DHH_ERROR_CODE_HTTPNOTOK, "Login response status("+resp.Status+"), url: "+path)
	}
	data, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		return NewError(DHH_ERROR_CODE_HTTPRESP_READERR, err2.Error())
	}

	var respJson DHHLoginResponseJson

	err = json.Unmarshal(data, &respJson)

	if err != nil {
		zlog.Logger.Error("Login Response Unmarshal error, resp data: " + string(data))
		return NewError(DHH_ERROR_CODE_JSONUNMARSHAL, "Login Response Unmarshal error, resp data: "+string(data))
	}

	if !respJson.HasError && respJson.Content != nil {
		if respJson.Content.Success && respJson.Content.Data != nil {
			if respJson.Content.Data.Redirect && strings.HasPrefix(respJson.Content.Data.RedirectUrl, DHH_LOGIN_REDIRECTURL_SUCCESS_PREFIX) {
				this.cookies = resp.Cookies()
				for _, v := range this.cookies {
					if v.Name == "XSRF-TOKEN" {
						v.Value = this.csrf
						return nil
					}
				}

				this.cookies = append(this.cookies, &http.Cookie{
					Name:       "XSRF-TOKEN",
					Value:      this.csrf,
					Path:       "/",
					Domain:     "",
					Expires:    time.Time{},
					RawExpires: "",
					MaxAge:     0,
					Secure:     false,
					HttpOnly:   true,
					SameSite:   0,
					Raw:        "",
					Unparsed:   nil,
				})

				return nil
			} else if respJson.Content.Data.TitleMsg != "" {
				zlog.Logger.Error("Login Response error, Login resp data: " + string(data))
				return NewError(DHH_ERROR_CODE_UNKNOWN, "Login Response error, TitleMsg: "+respJson.Content.Data.TitleMsg)
			} else {
				zlog.Logger.Error("Login Response error, Login resp data: " + string(data))
				return NewError(DHH_ERROR_CODE_UNKNOWN, "Login Response error, redirect failed. redirectUrl is :"+respJson.Content.Data.RedirectUrl)
			}
		}
	}

	return NewError(DHH_ERROR_CODE_UNKNOWN, "Login Response error, unknown.Login resp data: "+string(data))
}

func (this *DHHClient) LoadingTaobaoInitViewData() (map[string]interface{}, string, int) {
	mapResult := make(map[string]interface{})
	resp, _ := http.Get(DHH_API_ENDPOINT_INITVIEWDATA)
	html, _ := ioutil.ReadAll(resp.Body)

	reg := regexp.MustCompile("\"loginFormData\":([^}]*})")

	content := reg.FindStringSubmatch(string(html))
	if len(content) < 2 {
		log.Panic("the document be not matched")

	} else {
		json.Unmarshal([]byte(content[1]), &mapResult)
	}

	reg2 := regexp.MustCompile("rsaModulus\":\"([0-9a-z]+)")
	reg3 := regexp.MustCompile("rsaExponent\":\"([0-9a-z]+)")

	content2 := reg2.FindStringSubmatch(string(html))
	content3 := reg3.FindStringSubmatch(string(html))

	if len(content2) < 2 {
		return mapResult, "", 0
	}

	if len(content3) < 2 {
		return mapResult, "", 0
	}

	tmp, _ := strconv.ParseInt(content3[1], 16, 64)

	return mapResult, content2[1], int(tmp)

}

func (this *DHHClient) QueryTaskMaterial(taskId string, pageNum int, pageSize int) (int, *DHHError) {
	form := url.Values{
		DHH_QUERY_MATERIAL_FORM_APPID:      {"1"},
		DHH_QUERY_MATERIAL_FORM_SOURCE:     {""},
		DHH_QUERY_MATERIAL_FORM_TASKTYPEID: {taskId},
		DHH_QUERY_MATERIAL_FORM_STATUS:     {"1"},
		DHH_QUERY_MATERIAL_FORM_CSRF:       {this.csrf},
		DHH_QUERY_MATERIAL_FORM_PAGE:       {strconv.Itoa(pageNum)},
		DHH_QUERY_MATERIAL_FORM_PAGESIZE:   {strconv.Itoa(pageSize)}}

	url := DHH_API_ENDPOINT_QUERYMATERIAL

	bodyReader := strings.NewReader(form.Encode())

	req, _ := http.NewRequest("POST", url, bodyReader)

	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("x-xsrf-token", this.csrf)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("sec-fetch-site", "same-origin")
	for _, v := range this.cookies {
		req.AddCookie(v)
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		zlog.Logger.Warn(err.Error())
		return 0, NewError(DHH_ERROR_CODE_HTTPERROR, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		zlog.Logger.Warn(resp.Status)
		return 0, NewError(DHH_ERROR_CODE_HTTPNOTOK, "QueryTaskMaterial response status("+resp.Status+"), url: "+url)
	}

	var respJson TaskMaterialResponseJson

	data, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		zlog.Logger.Warn(err2.Error())
		return 0, NewError(DHH_ERROR_CODE_HTTPRESP_READERR, err2.Error())
	}

	err = json.Unmarshal(data, &respJson)

	if err != nil {
		zlog.Logger.Error("QueryTaskMaterial Response Unmarshal error, resp data: " + string(data))
		return 0, NewError(DHH_ERROR_CODE_JSONUNMARSHAL, "QueryTaskMaterial Response Unmarshal error, resp data: "+string(data))
	}

	if respJson.Successful && (respJson.List != nil) {
		println(respJson.List)
		for _, v := range respJson.List {
			if v.Status == 1 && strconv.Itoa(v.TaskTypeId) == taskId {
				timeStart, _ := time.Parse("2006-01-02", v.Start)
				timeEnd, _ := time.Parse("2006-01-02", v.End)
				timeNow := time.Now()

				if timeStart.Before(timeNow) && timeEnd.After(timeNow) {
					return v.MaterialId, nil
				}
			}
		}
	}
	return 0, NewError(DHH_ERROR_CODE_UNKNOWN, "QueryTaskMaterial Response unknown error, resp data: "+string(data))
}

func (this *DHHClient) QueryMaterialInfo(taskId string, materialId string, pageNum int, pageSize int, prefixFilter string) (string, int, int, *DHHError) {
	qs := url.Values{
		DHH_QUERY_MATERIALINFO_QS_APPID:      {"1"},
		DHH_QUERY_MATERIALINFO_QS_TASKTYPEID: {taskId},
		DHH_QUERY_MATERIALINFO_QS_MATERIALID: {materialId},
		DHH_QUERY_MATERIALINFO_QS_PAGE:       {strconv.Itoa(pageNum)},
		DHH_QUERY_MATERIALINFO_QS_PAGESIZE:   {strconv.Itoa(pageSize)},
		DHH_QUERY_MATERIALINFO_QS_CSRF:       {this.csrf}}

	req, _ := http.NewRequest("GET", DHH_API_ENDPOINT_QUERYMATERIALINFO+"?"+qs.Encode(), nil)

	req.Header.Set("Connection", "keep-alive")

	for _, v := range this.cookies {
		if v.Name == "XSRF-TOKEN" || v.Name == "cookie2" || v.Name == "sg" {
			req.AddCookie(v)
		}
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		zlog.Logger.Warn(err.Error())
		return "", 0, 0, NewError(DHH_ERROR_CODE_HTTPERROR, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		zlog.Logger.Warn(resp.Status)
		return "", 0, 0, NewError(DHH_ERROR_CODE_HTTPNOTOK, "QueryMaterialInfo response status("+resp.Status+"), url: "+DHH_API_ENDPOINT_QUERYMATERIALINFO+"?"+qs.Encode())
	}

	var respJson MaterialUrlResponseJson

	data, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		zlog.Logger.Warn(err2.Error())
		return "", 0, 0, NewError(DHH_ERROR_CODE_HTTPRESP_READERR, err2.Error())
	}

	err = json.Unmarshal(data, &respJson)

	if err != nil {
		zlog.Logger.Error("QueryMaterialInfo Response Unmarshal error, resp data: " + string(data))
		return "", 0, 0, NewError(DHH_ERROR_CODE_JSONUNMARSHAL, "QueryMaterialInfo Response Unmarshal error, resp data: "+string(data))
	}

	if respJson.Successful && (respJson.List != nil) {
		//		println(respJson.List)
		content := ""
		for _, v := range respJson.List {
			if len(prefixFilter) == 0 || strings.HasPrefix(v.Adname, prefixFilter) {
				line := strconv.Itoa(v.AdId) + "\t" + v.Adname + "\t" + v.ImgUrl + "\t" + v.H5Url + "\t" + v.DeepLinkUrl
				content += line + "\n"
			}
		}
		return content, respJson.Page, respJson.Total, nil
	}
	return "", 0, 0, NewError(DHH_ERROR_CODE_UNKNOWN, "QueryMaterialInfo Response unknown error, resp data: "+string(data))
}

func (this *DHHClient) ListAdSpaces(status int, pageNum int, pageSize int, prefix string) (string, int, int, *DHHError) {
	form := url.Values{
		DHH_QUERY_ADSPACES_FORM_STATUS:   {strconv.Itoa(status)},
		DHH_QUERY_ADSPACES_FORM_CSRF:     {this.csrf},
		DHH_QUERY_ADSPACES_FORM_PAGE:     {strconv.Itoa(pageNum)},
		DHH_QUERY_ADSPACES_FORM_PAGESIZE: {strconv.Itoa(pageSize)}}

	url := DHH_API_ENDPOINT_ADSPACES

	bodyReader := strings.NewReader(form.Encode())

	req, _ := http.NewRequest("POST", url, bodyReader)

	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("x-xsrf-token", this.csrf)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("sec-fetch-site", "same-origin")
	for _, v := range this.cookies {
		req.AddCookie(v)
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		zlog.Logger.Warn(err.Error())
		return "", 0, 0, NewError(DHH_ERROR_CODE_HTTPERROR, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		zlog.Logger.Warn(resp.Status)
		return "", 0, 0, NewError(DHH_ERROR_CODE_HTTPNOTOK, "ListAdSpaces response status("+resp.Status+"), url: "+url)
	}

	var respJson CommonResponseJson

	data, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		zlog.Logger.Warn(err2.Error())
		return "", 0, 0, NewError(DHH_ERROR_CODE_HTTPRESP_READERR, err2.Error())
	}

	if err = json.Unmarshal(data, &respJson); err != nil {
		zlog.Logger.Error("ListAdSpaces Response Unmarshal error, resp data: " + string(data))
		return "", 0, 0, NewError(DHH_ERROR_CODE_JSONUNMARSHAL, "ListAdSpaces Response Unmarshal error, resp data: "+string(data))
	}

	if respJson.Successful && (respJson.Data != nil) {
		var commonDataJson CommonDataJson
		if err2 := json.Unmarshal(*respJson.Data, &commonDataJson); err2 != nil {
			zlog.Logger.Error("ListAdSpaces Response Unmarshal error, resp data: " + string(data))
			return "", 0, 0, NewError(DHH_ERROR_CODE_JSONUNMARSHAL, "ListAdSpaces Response Unmarshal error, resp data: "+string(data))
		}

		if commonDataJson.Successful && (commonDataJson.List != nil) {
			content := ""
			for _, item := range commonDataJson.List {
				var adspaceJson AdSpaceJson

				if err3 := json.Unmarshal(*item, &adspaceJson); err3 != nil {
					zlog.Logger.Error("ListAdSpaces Response Unmarshal error, resp data: " + string(data))
					return "", 0, 0, NewError(DHH_ERROR_CODE_JSONUNMARSHAL, "ListAdSpaces Response Unmarshal error, resp data: "+string(data))
				}

				if len(prefix) == 0 || strings.HasPrefix(adspaceJson.Name, prefix) {
					line := strconv.Itoa(adspaceJson.Id) + "\t" + adspaceJson.Name + "\t" + strconv.Itoa(adspaceJson.Status)
					content += line + "\n"
				}
			}
			return content, commonDataJson.Page, commonDataJson.Total, nil
		}

	}
	return "", 0, 0, NewError(DHH_ERROR_CODE_UNKNOWN, "ListAdSpaces Response unknown error, resp data: "+string(data))
}

func (this *DHHClient) applyAdSpace(adSpaceName string, deliveryApp string, deliveryAddress string, effectExampleFileId int, effectExampleUrl string) *DHHError {
	form := url.Values{
		"channelId":           {this.channelId},
		"name":                {adSpaceName},
		"type":                {"5"},
		"deliveryApp":         {deliveryApp},
		"deliveryAddress":     {deliveryAddress},
		"effectExampleFileId": {strconv.Itoa(effectExampleFileId)},
		"effectExampleUrl":    {effectExampleUrl},
		"deliveryType":        {"2"},
		"_csrf":               {this.csrf}}

	bodyReader := strings.NewReader(form.Encode())

	req, _ := http.NewRequest("POST", DHH_API_ENDPOINT_APPLYADSPACE, bodyReader)

	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("x-xsrf-token", this.csrf)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("sec-fetch-site", "same-origin")

	for _, v := range this.cookies {
		req.AddCookie(v)
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		zlog.Logger.Warn(err.Error())
		return NewError(DHH_ERROR_CODE_HTTPERROR, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		zlog.Logger.Warn(resp.Status)
		return NewError(DHH_ERROR_CODE_HTTPNOTOK, "applyAdSpace response status("+resp.Status+"), url: "+DHH_API_ENDPOINT_APPLYADSPACE)
	}

	var respJson CommonItemResponseJson

	data, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		zlog.Logger.Warn(err2.Error())
		return NewError(DHH_ERROR_CODE_HTTPRESP_READERR, err2.Error())
	}

	err = json.Unmarshal(data, &respJson)

	if err != nil {
		zlog.Logger.Error("applyAdSpace Response Unmarshal error, resp data: " + string(data))
		return NewError(DHH_ERROR_CODE_JSONUNMARSHAL, "applyAdSpace Response Unmarshal error, resp data: "+string(data))
	}

	if respJson.Successful {
		return nil
	} else {
		zlog.Logger.Error("applyAdSpace Response Successful is false, resp data: " + string(data))
		return NewError(DHH_ERROR_CODE_NOTSUCCESS, "applyAdSpace Response Successful is false, resp data: "+string(data))
	}
}

func (this *DHHClient) AddAdspace(deliveryApp string, deliveryAddress string, deliveryAppFileName string, adName string) *DHHError {
	errBiz, fileId, fileUrl := this.bizFile(deliveryAppFileName)

	if errBiz != nil {
		return errBiz
	}

	return this.applyAdSpace(adName, deliveryApp, deliveryAddress, fileId, fileUrl)
}

func (this *DHHClient) BatchAdSpace(deliveryApp string, deliveryAddress string, deliveryAppFileName string, adNameFile string) *DHHError {

	errBiz, fileId, fileUrl := this.bizFile(deliveryAppFileName)

	if errBiz != nil {
		return errBiz
	}

	fi, err := os.Open(adNameFile)
	if err != nil {
		zlog.Logger.Error("open file failed: " + adNameFile + ",verbose:" + err.Error())
		return NewError(DHH_ERROR_CODE_OPENFILEFAILED, "open file failed: "+adNameFile+",verbose:"+err.Error())
	}
	defer fi.Close()
	// 读取内容
	buf := bufio.NewScanner(fi)
	for buf.Scan() {
		line := buf.Text()
		err2 := this.applyAdSpace(line, deliveryApp, deliveryAddress, fileId, fileUrl)
		if err2 != nil {
			zlog.Logger.Warn("apply AdSpace failed, adName(" + line + "), verbose:" + err2.Error())
			return err2
		} else {
			zlog.Logger.Info("apply AdSpace Success, adName(" + line + ")")
		}
	}
	return nil
}

func (this *DHHClient) bizFile(fileName string) (*DHHError, int, string) {

	url := DHH_API_ENDPOINT_BIZFILES

	var buff bytes.Buffer
	w := multipart.NewWriter(&buff)
	f, err := os.Open(fileName)
	if err != nil {
		return NewError(DHH_ERROR_CODE_OPENFILEFAILED, "open file failed: "+fileName), 0, ""
	}
	defer f.Close()

	fw, err := w.CreateFormFile("file", fileName)
	if err != nil {
		return NewError(DHH_ERROR_CODE_OPENFILEFAILED, "CreateFormFile failed: "+fileName), 0, ""
	}
	if _, err = io.Copy(fw, f); err != nil {
		return NewError(DHH_ERROR_CODE_IOERROR, "io.Copy failed"), 0, ""
	}
	w.WriteField("_csrf", this.csrf)
	w.WriteField("bizModule", "Contracts.Templates")

	w.Close()

	req, _ := http.NewRequest("POST", url, &buff)

	req.Header.Set("content-type", w.FormDataContentType())

	for _, v := range this.cookies {
		req.AddCookie(v)
	}

	resp, _ := http.DefaultClient.Do(req)

	if err != nil {
		zlog.Logger.Warn(err.Error())
		return NewError(DHH_ERROR_CODE_HTTPERROR, err.Error()), 0, ""
	}

	if resp.StatusCode != http.StatusOK {
		zlog.Logger.Warn(resp.Status)
		return NewError(DHH_ERROR_CODE_HTTPNOTOK, "bizFile response status("+resp.Status+"), url: "+DHH_API_ENDPOINT_BIZFILES), 0, ""
	}

	data, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		zlog.Logger.Warn(err2.Error())
		return NewError(DHH_ERROR_CODE_HTTPRESP_READERR, err2.Error()), 0, ""
	}

	var respJson CommonItemResponseJson

	err = json.Unmarshal(data, &respJson)

	if err != nil {
		zlog.Logger.Error("bizFile Response Unmarshal error, resp data: " + string(data))
		return NewError(DHH_ERROR_CODE_JSONUNMARSHAL, "bizFile Response Unmarshal error, resp data: "+string(data)), 0, ""
	}

	if respJson.Successful && (respJson.Data != nil) {
		if respJson.Data.FileProcessStateV1 == "SUCCESSFUL" {
			return nil, respJson.Data.FileID, respJson.Data.Url
		}
		zlog.Logger.Error("bizFile Response Successful is false, resp data: " + string(data))
		return NewError(DHH_ERROR_CODE_NOTSUCCESS, "bizFile Response Successful is false, resp data: "+string(data)), 0, ""
	}

	return NewError(DHH_ERROR_CODE_UNKNOWN, "bizFile Response unknown error, resp data: "+string(data)), 0, ""
}

package dhhclient

import (
	"bufio"
	"encoding/json"
	"bytes"
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
	"github.com/webview/webview"
)


type DHHClient struct {
	loginId		 string
	password2	string
	channelId    string
	csrf         string
	cookies      string
	ua           string
	bx_ua		 string
	initViewData map[string]interface{}
}

func NewClient(loginId string, password2 string, channelId string, ua string, bx_ua string) *DHHClient {
	return &DHHClient{
		loginId,
		password2,
		channelId,
		"",
		"",
		ua,
		bx_ua,
		nil}
}

func (this *DHHClient) InitClient()*DHHError{
	this.initViewData, _, _ = this.LoadingTaobaoInitViewData()

	csrf, errToken := this.getCsrfToken()

	if errToken != nil {
		return errToken
	}

	this.csrf = csrf

	cookies, errLogin := this.Login(this.loginId, this.password2)
	if errLogin != nil{
		return errLogin
	}

	for _, cookie := range cookies{
		this.cookies = this.cookies + cookie.Raw + "; "
	}

	return nil
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
	r.Header.Set("cookie", this.cookies)
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
		"ua":               {this.ua},
		"bx-ua":			{this.bx_ua},
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

	if respJson.RgvFlag587 == "sm" && len(respJson.AuthUrl) > 0{
		web := webview.New(true)
		defer web.Destroy()
		web.SetTitle("验证")
		web.SetSize(800, 600, webview.HintNone)
		web.Navigate(DHH_API_ENDPOINT_LOGINROOT + respJson.AuthUrl)
		println(DHH_API_ENDPOINT_LOGINROOT + respJson.AuthUrl)
		// go func(){
		// 	time.Sleep(time.Duration(30)*time.Second)
		// 	web.Terminate()
		// }()
		web.Run()
		return nil
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

func (this *DHHClient) Login(loginId string, password2 string) ([]*http.Cookie, *DHHError) {

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
		DHH_LOGIN_QS_KEY_FROMSITE: {strconv.FormatFloat(this.initViewData[DHH_LOGIN_QS_KEY_FROMSITE].(float64), 'f', -1, 64)},
		"_bx-v":{"2.0.31"}}

	path := DHH_API_ENDPOINT_LOGIN + "?" + qs.Encode()
	resp, err := http.PostForm(path, formData)
	if err != nil {
		return nil, NewError(DHH_ERROR_CODE_HTTPERROR, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return nil, NewError(DHH_ERROR_CODE_HTTPNOTOK, "Login response status("+resp.Status+"), url: "+path)
	}
	data, err2 := ioutil.ReadAll(resp.Body)
	if err2 != nil {
		return nil, NewError(DHH_ERROR_CODE_HTTPRESP_READERR, err2.Error())
	}

	var respJson DHHLoginResponseJson

	err = json.Unmarshal(data, &respJson)

	if err != nil {
		zlog.Logger.Error("Login Response Unmarshal error, resp data: " + string(data))
		return nil, NewError(DHH_ERROR_CODE_JSONUNMARSHAL, "Login Response Unmarshal error, resp data: "+string(data))
	}

	if respJson.RgvFlag587 == "sm" && len(respJson.AuthUrl) > 0{
		// web := webview.New(true)
		// defer web.Destroy()
		// web.SetTitle("验证")
		// web.SetSize(800, 600, webview.HintNone)
		// web.Navigate(DHH_API_ENDPOINT_LOGINROOT + respJson.AuthUrl)
		println(DHH_API_ENDPOINT_LOGINROOT + respJson.AuthUrl)
//		web.Run()
		
		
		// time.Sleep(time.Duration(30)*time.Second)
		// web.Terminate()
		return nil, nil
	}

	if !respJson.HasError && respJson.Content != nil {
		if respJson.Content.Success && respJson.Content.Data != nil {
			if respJson.Content.Data.Redirect && strings.HasPrefix(respJson.Content.Data.RedirectUrl, DHH_LOGIN_REDIRECTURL_SUCCESS_PREFIX) {
				cookies := resp.Cookies()
				for _, v := range cookies {
					if v.Name == "XSRF-TOKEN" {
						v.Value = this.csrf
						return cookies, nil
					}
				}

				cookies = append(cookies, &http.Cookie{
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

				return cookies, nil
			} else if respJson.Content.Data.TitleMsg != "" {
				zlog.Logger.Error("Login Response error, Login resp data: " + string(data))
				return nil, NewError(DHH_ERROR_CODE_UNKNOWN, "Login Response error, TitleMsg: "+respJson.Content.Data.TitleMsg)
			} else {
				zlog.Logger.Error("Login Response error, Login resp data: " + string(data))
				return nil, NewError(DHH_ERROR_CODE_UNKNOWN, "Login Response error, redirect failed. redirectUrl is :"+respJson.Content.Data.RedirectUrl)
			}
		}
	}

	return nil, NewError(DHH_ERROR_CODE_UNKNOWN, "Login Response error, unknown.Login resp data: "+string(data))
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
	
	this.setCookies(req)

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

	// for _, v := range this.cookies {
	// 	if v.Name == "XSRF-TOKEN" || v.Name == "cookie2" || v.Name == "sg" {
	// 		req.AddCookie(v)
	// 	}
	// }
	this.setCookies(req)

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
	
	this.setCookies(req)

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

	this.setCookies(req)

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

	this.setCookies(req)

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

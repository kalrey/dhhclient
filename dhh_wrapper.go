package dhhclient

import (
	"os"
	"strconv"
)

type DHHWrapper struct {
	client *DHHClient
}

func NewDHHWrapper(channelId string, loginId string, password string, ua string) (*DHHWrapper, *DHHError) {
	client := NewClient(channelId)

	if client == nil {
		return nil, nil
	}

	err := client.InitClient(ua)
	if err != nil {
		return nil, err
	}

	// err := client.AccountCheck(loginId)

	// if err != nil {
	// 	return nil, err
	// }

	err2 := client.Login(loginId, password)

	if err2 != nil {
		return nil, err2
	}

	return &DHHWrapper{client}, nil
}

func (this *DHHWrapper) AddAdSpaces(deliveryAppFileName string, adnameFile string) *DHHError {
	return this.client.BatchAdSpace("讯飞输入法",
		"https://sj.qq.com/myapp/detail.htm?apkName=com.iflytek.inputmethod",
		deliveryAppFileName,
		adnameFile)
}

func (this *DHHWrapper) GetTaskUrls(taskId string, prefixFilter string) *DHHError {
	materialId, err := this.client.QueryTaskMaterial(taskId, 1, 10)

	if err != nil {
		return err
	}
	println("MaterialId: ", materialId)

	fileContent := ""

	pageSize := 100
	pageNum := 1

	for {
		content, curNum, total, err2 := this.client.QueryMaterialInfo(taskId, strconv.Itoa(materialId), pageNum, pageSize, prefixFilter)

		if err2 != nil {
			return err2
		}

		fileContent += content

		if curNum*pageSize >= total {
			break
		}

		pageNum = curNum + 1

	}

	file, err3 := os.Create(taskId + "-urls.txt")
	if err3 != nil {
		return NewError(0, err3.Error())
	}
	file.WriteString(fileContent)
	file.Close()
	return nil
}

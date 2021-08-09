package dhhclient


type DHHClientV2 struct{
	DHHClient
	cookies string
}


func NewClientV2(channelId string, cookies string, csrf string)*DHHClientV2{
	c := &DHHClientV2{*NewClient("", "", channelId, "", ""), cookies}
	c.DHHClient.csrf = csrf
	c.DHHClient.cookies = cookies
	return c
}


func (this *DHHClientV2)InitClient()*DHHError{
	return nil
}




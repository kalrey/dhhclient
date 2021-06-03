package dhhclient


//Date			int		`json:"date"`
//TotalAmount		int		`json:"totalAmount"`
//TargetAmount	int		`json:"targetAmount"`
//ShowAccountFee	string	`json:"showAccountFee"`
//AdSpaceID		int		`json:"adSpaceID"`
//ShowQualityScore	string `json:"showQualityScore"`
//CustomOSName	string	`json:"customOS_Name"`
//AdSpaceName		string	`json:"adSpaceName"`


type TaskData struct {
	Date 			int		`Field:"日期"`
	AdSpaceName		string	`Field:"广告位名称"`
	AdSpaceID		string	`Field:"广告位ID"`
	ShowQualityScore	string	`Field:"质量分"`
	TotalAmount		int		`Field:"总计促活"`
	TargetAmount	int		`Field:"总计目标完成"`
	ShowAccountFee	string	`Field:"预估佣金"`
}









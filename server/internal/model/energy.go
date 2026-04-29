package model

type Energy struct {
	ID       uint   `json:"id" gorm:"column:id;primaryKey"`
	DATETIME string `json:"datetime" gorm:"column:DATETIME"`
	Zqmc     string `json:"Zqmc" gorm:"column:Zqmc"`
	Cbjgex   string `json:"Cbjgex" gorm:"column:Cbjgex"`
	Cbjg     string `json:"Cbjg" gorm:"column:Cbjg"`
	Ckcb     string `json:"Ckcb" gorm:"column:Ckcb"`
	Ckcbj    string `json:"Ckcbj" gorm:"column:Ckcbj"`
	Ckyk     string `json:"Ckyk" gorm:"column:Ckyk"`
	Ykbl     string `json:"Ykbl" gorm:"column:Ykbl"`
	Dryk     string `json:"Dryk" gorm:"column:Dryk"`
	Drykbl   string `json:"Drykbl" gorm:"column:Drykbl"`
	Cwbl     string `json:"Cwbl" gorm:"column:Cwbl"`
	Djsl     string `json:"Djsl" gorm:"column:Djsl"`
	Dqcb     string `json:"Dqcb" gorm:"column:Dqcb"`
	Gddm     string `json:"Gddm" gorm:"column:Gddm"`
	Gfmcdj   string `json:"Gfmcdj" gorm:"column:Gfmcdj"`
	Gfmrjd   string `json:"Gfmrjd" gorm:"column:Gfmrjd"`
	Gfssmmce string `json:"Gfssmmce" gorm:"column:Gfssmmce"`
	Gfye     string `json:"Gfye" gorm:"column:Gfye"`
	Jgbm     string `json:"Jgbm" gorm:"column:Jgbm"`
	Khdm     string `json:"Khdm" gorm:"column:Khdm"`
	Ksssl    string `json:"Ksssl" gorm:"column:Ksssl"`
	Kysl     string `json:"Kysl" gorm:"column:Kysl"`
	Ljyk     string `json:"Ljyk" gorm:"column:Ljyk"`
	Market   string `json:"Market" gorm:"column:Market"`
	Mrssc    string `json:"Mrssc" gorm:"column:Mrssc"`
	Sssl     string `json:"Sssl" gorm:"column:Sssl"`
	Szjsbs   string `json:"Szjsbs" gorm:"column:Szjsbs"`
	Zjzh     string `json:"Zjzh" gorm:"column:Zjzh"`
	Zqdm     string `json:"Zqdm" gorm:"column:Zqdm"`
	Zqlx     string `json:"Zqlx" gorm:"column:Zqlx"`
	Zqlxmc   string `json:"Zqlxmc" gorm:"column:Zqlxmc"`
	Zqsl     string `json:"Zqsl" gorm:"column:Zqsl"`
	Ztmc     string `json:"Ztmc" gorm:"column:Ztmc"`
	Ztmr     string `json:"Ztmr" gorm:"column:Ztmr"`
	Zxjg     string `json:"Zxjg" gorm:"column:Zxjg"`
	Zxsz     string `json:"Zxsz" gorm:"column:Zxsz"`
	Bz       string `json:"Bz" gorm:"column:Bz"`
}

func (Energy) TableName() string {
	return "energy"
}

type Bond struct {
	ID         uint   `json:"id" gorm:"column:id;primaryKey"`
	DATETIME   string `json:"datetime" gorm:"column:DATETIME"`
	Cbjg       string `json:"Cbjg" gorm:"column:Cbjg"`
	Market     string `json:"Market" gorm:"column:Market"`
	MarketName string `json:"MarketName" gorm:"column:MarketName"`
	Zqdm       string `json:"Zqdm" gorm:"column:Zqdm"`
	Zqlx       string `json:"Zqlx" gorm:"column:Zqlx"`
	Zqlxmc     string `json:"Zqlxmc" gorm:"column:Zqlxmc"`
	Zqmc       string `json:"Zqmc" gorm:"column:Zqmc"`
	Zxsz       string `json:"Zxsz" gorm:"column:Zxsz"`
}

type Assets struct {
	Djzj          string   `json:"djzj"`
	Drhz          string   `json:"drhj"`
	Dryk          string   `json:"dryk"`
	Kqzj          string   `json:"kqzj"`
	Kyzj          string   `json:"kyzj"`
	Ljyk          string   `json:"ljyk"`
	MoneyType     string   `json:"money_Type"`
	RMBZzc        string   `json:"RMBZzc"`
	Zjye          string   `json:"zjye"`
	Zxsz          string   `json:"zxsz"`
	Zzc           string   `json:"zzc"`
	Bonds         []Bond   `json:"bonds"`
	TotalSecMkval string   `json:"totalSecMkval"`
	Positions     []Energy `json:"positions"`
}

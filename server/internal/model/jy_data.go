package model

type JyData struct {
	ID       uint   `json:"id" gorm:"column:id;primaryKey"`
	NOEDATE  string `json:"NOEDATE" gorm:"column:NOEDATE"`
	Zqmc     string `json:"Zqmc" gorm:"column:Zqmc"`
	Ckcb     string `json:"Ckcb" gorm:"column:Ckcb"`
	Ckcbj    string `json:"Ckcbj" gorm:"column:Ckcbj"`
	Ckyk     string `json:"Ckyk" gorm:"column:Ckyk"`
	Ykbl     string `json:"Ykbl" gorm:"column:Ykbl"`
	Dryk     string `json:"Dryk" gorm:"column:Dryk"`
	Drykbl   string `json:"Drykbl" gorm:"column:Drykbl"`
	Cwbl     string `json:"Cwbl" gorm:"column:Cwbl"`
	Djjsl    string `json:"Djjsl" gorm:"column:Djjsl"`
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
	Lyjk     string `json:"Lyjk" gorm:"column:Lyjk"`
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
	Cjbgex   string `json:"Cjbgex" gorm:"column:Cjbgex"`
	Cjbjg    string `json:"Cjbjg" gorm:"column:Cjbjg"`
}

func (JyData) TableName() string {
	return "jy_data"
}

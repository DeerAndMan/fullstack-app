package service

// These summaries are tied to the cited disclosures or F10 entries. Dates refer to the source, not to a current contract.
var conceptReasonSources = map[string]map[string]MarketConceptReason{
	"gn_hwhs": {
		"sh600584": {
			Kind:    "封装测试",
			Text:    "芯源微2019年年报援引当时公开资料称，长电科技等境内企业承接了海思部分芯片封测。该资料未说明目前订单情况。",
			Sources: []MarketReasonSource{{Title: "芯源微2019年年报", URL: "https://money.finance.sina.com.cn/corp/view/vCB_AllBulletinDetail.php?id=6125850&stockid=688037"}},
		},
		"sh603236": {
			Kind:    "通信模组",
			Text:    "据同花顺转引的2020年5月26日公司互动回复，移远通信曾与海思在NB-IoT模组和5G模组上开展合作。",
			Sources: []MarketReasonSource{{Title: "同花顺F10 · 移远通信", URL: "https://basic.10jqka.com.cn/603236/concept.html"}},
		},
		"sh603700": {
			Kind: "物联网水表",
			Text: "宁水集团官网记载，公司曾与华为开发NB-IoT物联网水表。界面刊载的报道转引公司负责人称，部分水表所用模组内含海思芯片。",
			Sources: []MarketReasonSource{
				{Title: "宁水集团官网 · 智能水表业务", URL: "https://www.chinawatermeter.com/newsd/ning-shui-ji-tuan-tou-zi-xin-yi-xin-xi-ke-ji-jia-chi-chan-ye-lian-xie-tong-fa-zhan"},
				{Title: "界面新闻 · 水表芯片报道", URL: "https://www.jiemian.com/article/5136512.html"},
			},
		},
		"sh603738": {
			Kind:    "器件认证",
			Text:    "同花顺转引的2021年调研记录将海思列为泰晶科技石英晶体谐振器的认证方之一，原文未单列海思认证型号或订单规模。",
			Sources: []MarketReasonSource{{Title: "同花顺F10 · 泰晶科技", URL: "https://basic.10jqka.com.cn/603738/concept.html?cid=301459"}},
		},
		"sh688037": {
			Kind:    "半导体设备",
			Text:    "芯源微2019年招股意向书将海思光电子列为单片式湿法设备的代表客户。该类设备包括清洗、去胶和湿法刻蚀设备。",
			Sources: []MarketReasonSource{{Title: "芯源微2019年招股意向书", URL: "https://money.finance.sina.com.cn/corp/view/vCB_AllBulletinDetail.php?id=5761049&stockid=688037"}},
		},
		"sh688170": {
			Kind:    "半导体设备",
			Text:    "据同花顺转引的2024年8月公司互动回复，德龙激光的半导体精密激光微加工设备已进入海思供应链。",
			Sources: []MarketReasonSource{{Title: "同花顺 · 德龙激光互动回复摘录", URL: "https://field.10jqka.com.cn/20260814/c678968465.shtml"}},
		},
		"sz000062": {
			Kind:    "授权分销",
			Text:    "深圳华强于2025年1月披露，子公司深圳淇诺科技和深圳华强半导体集团都是海思代理商，业务属于电子元器件授权分销。",
			Sources: []MarketReasonSource{{Title: "深圳华强2025年1月调研记录", URL: "https://static.cninfo.com.cn/finalpage/2025-01-14/1222334006.PDF"}},
		},
		"sz000801": {
			Kind:    "芯片方案应用",
			Text:    "四川九洲2021年年报披露，基于海思HI3796CV300方案的8K超高清机顶盒已完成研发，达到批量出货状态。",
			Sources: []MarketReasonSource{{Title: "四川九洲2021年年报", URL: "https://disc.static.szse.cn/disc/disk03/finalpage/2022-04-20/ba632ba8-76cd-41be-8666-0c9cd5769f32.PDF"}},
		},
		"sz002156": {
			Kind:    "封装测试",
			Text:    "同花顺F10将海思列为通富微电拓展的芯片设计客户之一。通富微电主营集成电路封测，所引资料未披露海思的具体项目。",
			Sources: []MarketReasonSource{{Title: "同花顺F10 · 通富微电", URL: "https://basic.10jqka.com.cn/002156/concept.html?cid=309092"}},
		},
		"sz002185": {
			Kind:    "封装测试",
			Text:    "据同花顺F10，华天科技及部分子公司曾通过海思的质量体系审核或可靠性测试，并为海思提供过芯片封装测试服务。",
			Sources: []MarketReasonSource{{Title: "同花顺F10 · 华天科技", URL: "https://basic.10jqka.com.cn/002185/concept.html?cid=307816"}},
		},
		"sz002371": {
			Kind:    "半导体设备",
			Text:    "同花顺F10在集成电路设备与服务业务中列海思为北方华创的客户，所引资料未细分海思使用的设备型号。",
			Sources: []MarketReasonSource{{Title: "同花顺F10 · 北方华创", URL: "https://basic.10jqka.com.cn/002371/concept.html?cid=308527"}},
		},
		"sz002436": {
			Kind:    "历史客户",
			Text:    "兴森科技2012年披露IC封装载板项目时，把海思列为长期客户之一。该公告说明了客户资源，未披露海思采购载板的具体订单。",
			Sources: []MarketReasonSource{{Title: "兴森科技2012年董事会公告", URL: "https://epaper.stcn.com/paper/zqsb/html/2012-09/05/content_403338.htm"}},
		},
		"sz002819": {
			Kind:    "测试服务",
			Text:    "据同花顺转引的2025年5月21日公司回复，海思是东方中科的客户之一，东方中科为其提供综合测试方案。",
			Sources: []MarketReasonSource{{Title: "同花顺F10 · 东方中科", URL: "https://basic.10jqka.com.cn/002819/concept.html?cid=308527"}},
		},
		"sz002881": {
			Kind:    "通信模组",
			Text:    "同花顺F10记录，美格智能曾于2019年和海思推出采用国产芯片的SLM790 4G模组。",
			Sources: []MarketReasonSource{{Title: "同花顺F10 · 美格智能", URL: "https://basic.10jqka.com.cn/002881/concept.html?cid=308527"}},
		},
		"sz002902": {
			Kind:    "供应关系",
			Text:    "同花顺转引铭普光磁2024年5月公众号，将海思列为其企业客户之一。铭普光磁生产光电子和磁性元件，所引资料未披露海思对应的产品型号。",
			Sources: []MarketReasonSource{{Title: "同花顺 · 铭普光磁业务摘要", URL: "https://field.10jqka.com.cn/20260911/c679832511.shtml"}},
		},
		"sz002967": {
			Kind:    "芯片检测",
			Text:    "广电计量2021年调研记录将海思列为芯片检测业务的主要客户之一。相关服务侧重芯片验证和元器件应用环节。",
			Sources: []MarketReasonSource{{Title: "广电计量2021年调研记录", URL: "https://pdf.dfcfw.com/pdf/H22_AN202101281455401647_1.pdf"}},
		},
		"sz300050": {
			Kind:    "产品适配",
			Text:    "世纪鼎利2023年披露，公司购买了海思芯片接口控制文档的非独家授权，用于路测仪表适配华为手机。公司明确否认联合开发。",
			Sources: []MarketReasonSource{{Title: "世纪鼎利2023年关注函回复", URL: "https://notice.10jqka.com.cn/api/pdf/6c56843919ea90f5.pdf"}},
		},
		"sz300079": {
			Kind:    "产品预集成",
			Text:    "同花顺F10将数码视讯的数字版权保护方案与海思芯片预集成列为关联依据。",
			Sources: []MarketReasonSource{{Title: "同花顺F10 · 数码视讯", URL: "https://basic.10jqka.com.cn/300079/concept.html?cid=301997"}},
		},
		"sz300184": {
			Kind:    "授权分销",
			Text:    "力源信息2019年公告称，其自2009年起代理海思芯片，披露的产品包括安防视频编码、Wi-Fi和NB-IoT芯片、宽带载波芯片。",
			Sources: []MarketReasonSource{{Title: "力源信息2019-046号公告", URL: "https://pdf.dfcfw.com/pdf/H2_AN201905201331211670_1.pdf"}},
		},
		"sz300264": {
			Kind:    "研发线索",
			Text:    "2020年1月转引的公司互动回复称，在研智慧家庭产品可能集成海思芯片，当时尚未开展相关物联网应用合作。",
			Sources: []MarketReasonSource{{Title: "通信行业周报 · 佳创视讯互动回复摘录", URL: "https://pdf.dfcfw.com/pdf/H3_AP202001151374206338_1.pdf"}},
		},
		"sz300292": {
			Kind: "物联网模组",
			Text: "吴通控股2018年年报披露子公司宽翼通信的NB-IoT模组WT208曾中标集采。同期行业产品资料标注该模组采用海思芯片。",
			Sources: []MarketReasonSource{
				{Title: "吴通控股2018年年报", URL: "https://static.cninfo.com.cn/finalpage/2019-04-26/1206101453.PDF"},
				{Title: "小树开源 · NB-IoT模组资料", URL: "https://www.eetree.cn/doc/1149"},
			},
		},
		"sz300339": {
			Kind:    "开发平台",
			Text:    "同花顺的业务资料显示，润和软件曾基于海思芯片推出开发平台，应用方向包括智能安防、数字媒体和智能支付。",
			Sources: []MarketReasonSource{{Title: "同花顺 · 润和软件业务摘要", URL: "https://field.10jqka.com.cn/20260911/c679832511.shtml"}},
		},
		"sz300371": {
			Kind:    "物联网模组",
			Text:    "据第一财经转引的2019年公司互动回复，汇中股份的超声水表、热量表、流量计等NB-IoT产品曾使用搭载海思芯片的通信模组。",
			Sources: []MarketReasonSource{{Title: "第一财经 · 汇中股份互动回复", URL: "https://www.yicai.com/news/100203601.html"}},
		},
		"sz300448": {
			Kind:    "间接线索",
			Text:    "浩云科技2019年年报披露，它与华为开展金融物联合作，并使用华为提供的高算力芯片开发银行边缘物联方案。年报未指名海思或芯片型号。",
			Sources: []MarketReasonSource{{Title: "浩云科技2019年年报摘要", URL: "https://static.cninfo.com.cn/finalpage/2020-04-24/1207583559.PDF"}},
		},
		"sz300456": {
			Kind:    "IC设计服务",
			Text:    "赛微电子2025年年报披露，其控股子公司展诚科技从事IC设计服务，并将海思列为累计服务过的客户之一。",
			Sources: []MarketReasonSource{{Title: "赛微电子2025年年报", URL: "https://static.cninfo.com.cn/finalpage/2026-03-27/1225036416.PDF"}},
		},
		"sz300460": {
			Kind:    "器件认证",
			Text:    "同花顺资料将海思列为ST惠伦频率控制元器件的认证平台之一，具体获认证型号和采购量未在所引资料中单列。",
			Sources: []MarketReasonSource{{Title: "同花顺 · ST惠伦业务摘要", URL: "https://field.10jqka.com.cn/20260911/c679832511.shtml"}},
		},
		"sz300496": {
			Kind:    "参考设计",
			Text:    "中科创达2020年再融资反馈回复披露，公司曾与海思共同开发机器人参考设计。",
			Sources: []MarketReasonSource{{Title: "中科创达2020年再融资反馈回复", URL: "https://pdf.dfcfw.com/pdf/H2_AN202004301379113849_1.pdf"}},
		},
		"sz300504": {
			Kind:    "网络设备",
			Text:    "据同花顺转引的2025年9月17日公司回复，天邑股份与海思的合作涉及部分宽带网关和路由器产品。",
			Sources: []MarketReasonSource{{Title: "同花顺 · 天邑股份互动回复摘录", URL: "https://field.10jqka.com.cn/20260911/c679832511.shtml"}},
		},
		"sz300531": {
			Kind:    "终端芯片方案",
			Text:    "优博讯2025年半年报披露，公司在智能终端研发中使用海思国产芯片，优化终端、边缘设备与云端的算力协作。",
			Sources: []MarketReasonSource{{Title: "优博讯2025年半年报", URL: "https://static.cninfo.com.cn/finalpage/2025-08-26/1224566106.PDF"}},
		},
		"sz300598": {
			Kind:    "系统适配",
			Text:    "诚迈科技2020年问询函回复显示，当时披露的桌面操作系统研发已适配包括海思麒麟在内的多种芯片平台。",
			Sources: []MarketReasonSource{{Title: "诚迈科技2020年审核问询函回复", URL: "https://pdf.dfcfw.com/pdf/H2_AN202012241443565132_1.pdf"}},
		},
		"sz300602": {
			Kind:    "屏蔽与热管理",
			Text:    "据同花顺转引的2023年9月4日公司回复，飞荣达曾向海思提供电磁屏蔽和热管理材料及器件。",
			Sources: []MarketReasonSource{{Title: "同花顺F10 · 飞荣达", URL: "https://basic.10jqka.com.cn/144/300602/concept.html"}},
		},
		"sz300609": {
			Kind:    "客流分析终端",
			Text:    "汇纳科技2019年半年报披露，采用海思AI芯片并集成自有算法的新一代客流分析产品已开始规模化销售。",
			Sources: []MarketReasonSource{{Title: "汇纳科技2019年半年报", URL: "https://static.cninfo.com.cn/finalpage/2019-08-16/1206534697.PDF"}},
		},
		"sz300682": {
			Kind:    "历史采购",
			Text:    "朗新科技2019年收购易视腾的问询回复将海思列为易视腾的历史供应商。易视腾从事智能电视终端业务，相关采购包括主控和存储芯片等。",
			Sources: []MarketReasonSource{{Title: "朗新科技2019年重组反馈回复", URL: "https://pdf.dfcfw.com/pdf/H2_AN201903191307082538_1.pdf"}},
		},
		"sz300726": {
			Kind:    "器件供货",
			Text:    "同花顺F10披露，宏达电子控股子公司宏达恒芯曾向海思小批量供货，相关器件用于光通信模块和激光模块。",
			Sources: []MarketReasonSource{{Title: "同花顺F10 · 宏达电子", URL: "https://basic.10jqka.com.cn/300726/concept.html?cid=300105"}},
		},
		"sz301312": {
			Kind:    "设备客户",
			Text:    "据同花顺转引的2023年4月公司回复，海思光电子是智立方的客户之一。该份回复未列明采购的设备型号。",
			Sources: []MarketReasonSource{{Title: "同花顺 · 智立方互动回复摘录", URL: "https://field.10jqka.com.cn/20260814/c678968465.shtml"}},
		},
		"sz301316": {
			Kind: "芯片测试",
			Text: "据同花顺转引的2023年9月5日公司回复，慧博云通为海思提供芯片外场测试服务。其2022年招股说明书也将海思列为报告期主要客户之一。",
			Sources: []MarketReasonSource{
				{Title: "慧博云通2022年招股说明书", URL: "https://static.cninfo.com.cn/finalpage/2022-10-10/1214724186.PDF"},
				{Title: "同花顺F10 · 慧博云通", URL: "https://basic.10jqka.com.cn/301316/concept.html?cid=300163"},
			},
		},
	},
}

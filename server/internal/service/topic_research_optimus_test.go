package service

import "testing"

func TestOptimusEvidenceRequiresCompanyOriginalForStatus(t *testing.T) {
	t.Parallel()
	const company = "样本股份"
	const code = "300123"
	info := ResearchSource{ID: "S1", Provider: "cninfo.com.cn", Title: "样本股份投资者关系活动记录"}
	const context = "样本股份 证券代码：300123 投资者关系活动记录。"
	for _, tt := range []struct {
		name, body, quote, answer, want string
		open, media                     bool
	}{
		{"company shipment", "公司回复：公司已批量供货特斯拉人形机器人关节组件。", "公司已批量供货特斯拉人形机器人关节组件。", "confirmed", "confirmed", true, false},
		{"shipment with customer before verb", "公司答复：本公司已向特斯拉人形机器人供货。", "本公司已向特斯拉人形机器人供货。", "confirmed", "confirmed", true, false},
		{"company sample", "公司答复：特斯拉人形机器人项目已经完成送样验证。", "特斯拉人形机器人项目已经完成送样验证。", "sample", "sample", true, false},
		{"sample with customer before verb", "公司答复：公司已向特斯拉Optimus项目送样。", "公司已向特斯拉Optimus项目送样。", "sample", "sample", true, false},
		{"company denial", "回复：公司PEEK部件尚未应用于特斯拉人形机器人上。", "公司PEEK部件尚未应用于特斯拉人形机器人上。", "denied", "denied", true, false},
		{"company denial in separate short reply", "问题1：公司已进入特斯拉Optimus的供应商名录，是否属实？回答：公司暂未与特斯拉开展合作。", "公司已进入特斯拉Optimus的供应商名录，是否属实？回答：公司暂未与特斯拉开展合作。", "denied", "denied", true, false},
		{"negation outside quote", "答：尚未实现批量供货，特斯拉人形机器人合作仍需核实。", "实现批量供货，特斯拉人形机器人合作仍需核实。", "confirmed", "reported", true, false},
		{"investor question", "投资者提问：本公司已批量供货特斯拉人形机器人关节组件，望回复。", "本公司已批量供货特斯拉人形机器人关节组件", "confirmed", "reported", true, false},
		{"media only", "公司已批量供货特斯拉人形机器人关节组件。", "公司已批量供货特斯拉人形机器人关节组件。", "confirmed", "reported", true, true},
		{"search snippet", "公司已批量供货特斯拉人形机器人关节组件。", "公司已批量供货特斯拉人形机器人关节组件。", "confirmed", "unverified", false, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			doc := optimusDocument{source: info, text: context + tt.body, fullText: tt.open}
			doc.source.Excerpt = context + tt.body
			if tt.media {
				doc.source.Provider = "stcn.com"
			}
			ev, ok := optimusEvidenceFor(doc, optimusFinding{Quote: tt.quote, Answer: tt.answer}, company, code)
			if !ok || ev.Status != tt.want {
				t.Fatalf("evidence = %+v, accepted = %t, want %q", ev, ok, tt.want)
			}
		})
	}
}

func TestOptimusQueryRouting(t *testing.T) {
	t.Parallel()
	for _, query := range []string{"查询机器人、特斯拉链", "Optimus 的执行器", "人形机器人 T链"} {
		if !isOptimusQuestion(query) {
			t.Errorf("missed Optimus query %q", query)
		}
	}
	if isOptimusQuestion("查询特斯拉汽车零部件") {
		t.Fatal("automotive query was misclassified")
	}
}

func TestOptimusDoesNotInferStageFromSearchQuery(t *testing.T) {
	t.Parallel()
	if got := optimusStageFor("制造设备与检测", "特斯拉机器人团队到样本股份开启审厂。"); got != optimusUnspecifiedStage {
		t.Fatalf("stage = %q for an audit without a named component", got)
	}
	if got := optimusStageFor("制造设备与检测", "样本股份向特斯拉人形机器人供应丝杠。"); got != "丝杠与减速器" {
		t.Fatalf("stage = %q for a named screw component", got)
	}
}

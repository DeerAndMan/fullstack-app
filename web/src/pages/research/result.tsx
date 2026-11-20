import { Link } from "react-router-dom";
import { Alert, Card, Collapse, Empty, Space, Table, Tag } from "antd";
import { updateTime } from "../hotspots/data";
import styles from "./research.module.css";
import type { ResearchCandidate, ResearchReport, ResearchRun, ResearchSource } from "./research-data";

const evidenceLabels = { double: "双站收录", combined: "跨站补齐题材", single: "单站收录" };
const skillCheckLabels = { orders: "订单兑现", industry: "行业景气", risks: "风险反证" };
const chainLabels = {
  confirmed: ["已确认供货", "green"],
  sample: ["送样验证", "blue"],
  denied: ["公司明确否认", "red"],
  reported: ["报道或传闻", "gold"],
  unverified: ["待核验原文", "default"],
};

function SourceLink({ source, compact = false }: { source: ResearchSource; compact?: boolean }) {
  if (/^\/hotspots\/review(?:\?[^#\\\s]*)?$/.test(source.url)) {
    return <Link to={source.url}>{source.title}</Link>;
  }
  let valid = false;
  try {
    const url = new URL(source.url);
    valid = ["https:", "http:"].includes(url.protocol) && !url.username && !url.password;
  } catch {
    // 记录中缺少可点击的源链接。
  }
  const text = compact ? `${source.provider} · ${source.title}` : source.title;
  return valid ? <a href={source.url} target="_blank" rel="noreferrer">{text}</a> : <span>{text}</span>;
}

function OptimusResults({ report }: { report: ResearchReport }) {
  if (!report.optimus) return null;
  const { stages, stocks } = report.optimus;
  const sources = new Map(report.sources.map(source => [source.id, source]));
  const total = new Set(stocks.map(stock => stock.code)).size;
  return <Card title={`Optimus 供应链 · ${total} 只股票`}>
    {stages.map(stage => {
      const stageStocks = stocks.filter(stock => stock.stage === stage);
      return <section key={stage} className={styles.chainStage}>
        <h3>{stage}<span>{stageStocks.length} 只</span></h3>
        {stageStocks.length === 0 && <p className={styles.muted}>本次未取得可核验的个股证据。</p>}
        {stageStocks.map(stock => <article className={styles.chainStock} key={stock.code}>
          <header>
            <a href={stock.quoteUrl} target="_blank" rel="noreferrer"><strong>{stock.name}</strong> {stock.code}</a>
            <Tag color={chainLabels[stock.status][1]}>{chainLabels[stock.status][0]}</Tag>
          </header>
          <ul className={styles.evidenceList}>
            {stock.evidence.map(evidence => {
              const source = sources.get(evidence.sourceId);
              if (!source) return null;
              return <li key={evidence.sourceId}>
                <div className={styles.evidenceMeta}>
                  <Tag color={chainLabels[evidence.status][1]}>{chainLabels[evidence.status][0]}</Tag>
                  <span>{source.date?.slice(0, 10)}{evidence.fullText ? "" : " · 检索摘要"}</span>
                </div>
                <p>{evidence.quote}</p>
                <SourceLink source={source} compact />
              </li>;
            })}
          </ul>
        </article>)}
      </section>;
    })}
  </Card>;
}

export default function ResearchResult({ run }: { run: ResearchRun }) {
  const report = run.result;
  if (!report) {
    return run.status === "running"
      ? <Alert showIcon type="info" message="正在交叉核对资料。可以离开页面，完成后会保存到查询历史。" />
      : <Alert showIcon type="warning" message={run.error || "没有可显示的结果。"} />;
  }

  if (report.topics.length === 1 && report.topics[0] === "强势股复盘" &&
      (report.sources.length === 0 || report.sources[0].provider === "复盘归档")) {
    return <Card title="复盘归档">
      <p>{report.conclusion}</p>
      {report.notes.filter(note => !report.conclusion?.includes(note)).map(note => <p key={note}>{note}</p>)}
      {report.sources.map(source => <SourceLink key={source.id} source={source} />)}
    </Card>;
  }
  const sources = new Map(report.sources.map(source => [source.id, source]));
  const observations = new Map(report.observations?.map(row => [row.code, row]));
  const relatedSources = (ids: string[]) => ids.flatMap(id => sources.has(id) ? [sources.get(id)!] : []);
  return (
    <div className={styles.stack}>
      <Card title="查询结果">
        <div className={styles.researchMeta}>
          <span>{updateTime(run.created_at)} · {run.model}</span>
          {run.referenceId && <span>引用历史 #{run.referenceId}</span>}
          <Tag>{report.optimus ? "按供应链环节" : report.match === "any" ? "任一题材" : "同时匹配"}</Tag>
        </div>
        {report.topics.map(topic => <Tag key={topic}>{topic}</Tag>)}
        {report.conclusion && <section className={styles.researchSummary}>
          <h3>{report.optimus ? "核验概况" : "AI 解读"}</h3>
          <p>{report.conclusion}</p>
        </section>}
        {report.notes.map((note, i) => <p className={styles.note} key={i}>{note}</p>)}
        {!report.optimus && <Table<ResearchCandidate>
          size="small" rowKey="code" dataSource={report.candidates}
          scroll={{ x: "max-content", y: report.candidates.length > 15 ? 540 : undefined }}
          pagination={false}
          columns={[
            { title: "股票", render: (_, row) => <strong>{row.name} {row.code}</strong> },
            { title: "覆盖题材", dataIndex: "topics", render: (topics: string[]) => topics.join("、") },
            { title: "核验范围", dataIndex: "status", render: (status: ResearchCandidate["status"]) => (
              <Tag color={status === "double" ? "green" : status === "combined" ? "blue" : undefined}>{evidenceLabels[status]}</Tag>
            ) },
            { title: "成分股证据", render: (_, row) => <Space wrap>{relatedSources(row.sourceIds).map(source =>
              <SourceLink key={source.id} source={source} compact />,
            )}</Space> },
          ]}
        />}
        {observations.size > 0 && <section className={styles.researchSummary}>
          <h3>待复核的个股线索</h3>
          {report.observations?.map(row => <div key={row.code} className={styles.researchObservation}>
            <strong>{row.name} {row.code}</strong> {row.note}
            <Space wrap>
              {relatedSources(row.sourceIds).map(source => <SourceLink key={source.id} source={source} />)}
              {row.quoteUrl && <a href={row.quoteUrl} target="_blank" rel="noreferrer">核对证券代码</a>}
            </Space>
          </div>)}
        </section>}
      </Card>

      <OptimusResults report={report} />

      {Boolean(report.skillChecks?.length) && <Card title="Skill 原文核验">
        {report.skillChecks?.map(check => <section className={styles.researchSummary} key={check.method}>
          <h3>{skillCheckLabels[check.method]}</h3>
          {check.findings.length === 0 && <p className={styles.muted}>本轮未取得足以支持结论的原文。</p>}
          {check.findings.map((finding, index) => <div className={styles.researchObservation} key={index}>
            {finding.subject && <strong>{finding.subject}</strong>}
            <p>{finding.note}</p>
            <ul className={styles.evidenceList}>{finding.evidence.map(({ sourceId, quote }) => {
              const source = sources.get(sourceId);
              return source && <li key={`${sourceId}:${quote}`}>
                <p><q>{quote}</q></p>
                <span>{sourceId} {source.date?.slice(0, 10)} · </span>
                <SourceLink source={source} compact />
              </li>;
            })}</ul>
          </div>)}
        </section>)}
      </Card>}

      <Card title="检索来源与取数范围">
        {report.sources.length === 0 && <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="本次暂无可查阅来源" />}
        <ul className={styles.researchSources}>
          {report.sources.map(source => <li key={source.id}>
            <SourceLink source={source} />
            <div className={styles.muted}>{source.id} · {source.provider}{source.date ? ` · ${source.date.slice(0, 10)}` : ""}</div>
            {source.excerpt && <p className={styles.note}>{source.excerpt}</p>}
          </li>)}
        </ul>
      </Card>

      {Boolean(run.skills?.length) && <Card title="本次使用的 Skill 快照">
        <Collapse items={run.skills?.map(skill => ({
          key: skill.id, label: skill.name, children: <p className={styles.skillText}>{skill.instructions}</p>,
        }))} />
      </Card>}
    </div>
  );
}


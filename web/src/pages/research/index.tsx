import { useEffect, useRef, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Alert, Button, Input, Select, Skeleton, theme } from "antd";
import { useAuthStore } from "@/stores/auth";
import { withTokenColors } from "../hotspots/panels";
import { updateTime } from "../hotspots/data";
import {
  listResearchRuns,
  listSkillCatalog,
  groupSkills,
  readResearchRun,
  startResearchRun,
  researchError,
} from "./research-data";
import type { ResearchRun } from "./research-data";
import Turn from "./turn";
import styles from "./research.module.css";

export default function ResearchPage() {
  const account = useAuthStore(s => s.user?.id);
  return <Chat key={account} account={account} />;
}

function Chat({ account }: { account?: number }) {
  const { token } = theme.useToken();
  const [params, setParams] = useSearchParams();
  const id = Number(params.get("run")) || 0;
  const [question, setQuestion] = useState("");
  const [selected, setSelected] = useState<string[]>();
  const [filter, setFilter] = useState("");
  const [editing, setEditing] = useState<{ id: number; skillIds: string[] }>();
  const [historyOpen, setHistoryOpen] = useState(false);
  const timeline = useRef<HTMLDivElement>(null);
  const client = useQueryClient();
  const key = ["topic-research", account, "runs"];
  const runs = useInfiniteQuery({
    queryKey: key,
    queryFn: ({ pageParam }) => listResearchRuns(pageParam),
    initialPageParam: 0,
    getNextPageParam: last => (last.length === 20 ? last.at(-1)?.id : undefined),
    refetchInterval: q => (q.state.data?.pages.some(p => p.some(r => r.status === "running")) ? 10000 : false),
  });
  const skills = useQuery({ queryKey: ["topic-research", account, "skill-catalog"], queryFn: listSkillCatalog });
  const selectedSkills = (editing?.skillIds ?? selected ?? skills.data?.skills.filter(skill => skill.defaultSelected).map(skill => skill.id) ?? [])
    .filter(id => !skills.data || skills.data.skills.some(skill => skill.id === id));
  const { hasNextPage, isFetching, isFetchNextPageError, fetchNextPage } = runs;
  useEffect(() => {
    if (hasNextPage && !isFetching && !isFetchNextPageError) void fetchNextPage();
  }, [hasNextPage, isFetching, isFetchNextPageError, fetchNextPage]);
  const records = runs.data?.pages.flat() ?? [];
  const direct = useQuery({
    queryKey: ["topic-research", account, "run", id],
    queryFn: () => readResearchRun(id),
    enabled: !!id && !records.some(r => r.id === id),
  });
  const all = records.slice();
  if (id && direct.data && !all.some(r => r.id === id)) all.push(direct.data);
  const byId = new Map(all.map(r => [r.id, r]));
  const rootById = new Map<number, number>();
  const groups = new Map<number, ResearchRun[]>();
  for (const row of all.sort((a, b) => a.id - b.id)) {
    const root = rootById.get(row.referenceId ?? 0) ?? row.referenceId ?? row.id;
    rootById.set(row.id, root);
    if (!groups.has(root)) groups.set(root, []);
    groups.get(root)!.push(row);
  }
  const path: ResearchRun[] = [];
  for (
    let row = byId.get(id);
    row;
    row = row.referenceId && row.referenceId < row.id ? byId.get(row.referenceId) : undefined
  )
    path.unshift(row);
  const current = byId.get(id);
  const busy = records.some(r => r.status === "running") || current?.status === "running";
  function choose(next: number) {
    setParams(next ? { run: String(next) } : {});
    setQuestion("");
    setEditing(undefined);
    setHistoryOpen(false);
  }
  const send = useMutation({
    mutationFn: startResearchRun,
    onSuccess: row => {
      client.setQueryData(["topic-research", account, "run", row.id], row);
      void client.invalidateQueries({ queryKey: key });
      choose(row.id);
    },
  });
  function submit() {
    if (question.trim() && !skills.isPending && !send.isPending && !busy && (!id || current))
      send.mutate({ question: question.trim(), skillIds: selectedSkills, referenceId: editing?.id ?? (id || undefined) });
  }
  useEffect(() => {
    if (timeline.current) timeline.current.scrollTop = timeline.current.scrollHeight;
  }, [id, path.length]);
  const grouped = [...groups.entries()]
    .sort((a, b) => b[1].at(-1)!.id - a[1].at(-1)!.id)
    .filter(([, turns]) => turns.some(t => t.question.toLowerCase().includes(filter.trim().toLowerCase())));
  return (
    <main className={styles.page} style={withTokenColors(token)}>
      <header className={styles.pageHeader}>
        <h1>AI查询</h1>
        <Link to="/research/skills" state={{ run: id }}>
          管理 Skill →
        </Link>
      </header>
      <div className={styles.workspace}>
        <aside className={styles.sidebar}>
          <header>
            <h2>历史对话</h2>
            <Button
              disabled={send.isPending}
              onClick={() => {
                choose(0);
                setSelected(undefined);
              }}
            >
              ＋ 新对话
            </Button>
            <button className={styles.mobileToggle} onClick={() => setHistoryOpen(!historyOpen)}>
              {historyOpen ? "收起" : "展开"}
            </button>
          </header>
          <div className={historyOpen ? styles.historyOpen : styles.historyBody}>
            <Input
              allowClear
              aria-label="搜索历史对话"
              placeholder="搜索问题或追问"
              value={filter}
              onChange={e => setFilter(e.target.value)}
            />
            {runs.isPending && <Skeleton active />}
            {runs.isError && (
              <Button onClick={() => void (isFetchNextPageError ? fetchNextPage() : runs.refetch())}>
                重新加载历史
              </Button>
            )}
            {!runs.isPending && !grouped.length && <p>{filter ? "没有匹配的对话" : "暂无历史对话"}</p>}
            {grouped.map(([root, turns]) => {
              const last = turns.at(-1)!;
              return (
                <div key={root} className={rootById.get(id) === root ? styles.activeSession : styles.session}>
                  <button disabled={send.isPending} onClick={() => choose(last.id)}>
                    <strong>{turns[0].question}</strong>
                    <span>
                      {turns.length} 轮 · {updateTime(last.created_at)}
                      {last.status === "running" ? " · 研究中" : ""}
                    </span>
                  </button>
                  {turns.length > 1 && (
                    <details>
                      <summary>查看各轮问题</summary>
                      {turns.map((row, i) => (
                        <button
                          disabled={send.isPending}
                          key={row.id}
                          aria-current={id === row.id}
                          onClick={() => choose(row.id)}
                        >
                          {i + 1}. {row.question}
                        </button>
                      ))}
                    </details>
                  )}
                </div>
              );
            })}
          </div>
        </aside>
        <section className={styles.chat}>
          <div className={styles.timeline} ref={timeline}>
            {!id && (
              <div className={styles.welcome}>
                <h2>想研究哪个题材？</h2>
                <p>从一个题材、公司或供应链环节开始，收到回答后可继续追问。</p>
              </div>
            )}
            {!!id &&
              !current &&
              (direct.isError ? (
                <Alert
                  type="error"
                  message="无法读取这段对话"
                  action={<Button onClick={() => void direct.refetch()}>重试</Button>}
                />
              ) : (
                <Skeleton active />
              ))}
            {path.map(row => (
              <Turn
                key={row.id}
                row={row}
                account={account}
                expanded={row.id === id}
                edit={run => {
                  setQuestion(run.question);
                  setEditing({ id: run.id, skillIds: (run.skills ?? []).map(skill => skill.id) });
                  document.getElementById("research-message")?.focus();
                }}
              />
            ))}
          </div>
          <form
            className={styles.composer}
            onSubmit={e => {
              e.preventDefault();
              submit();
            }}
          >
            {editing && (
              <div>
                从所选问题重新提问{" "}
                <button
                  type="button"
                  onClick={() => {
                    setEditing(undefined);
                    setQuestion("");
                  }}
                >
                  取消
                </button>
              </div>
            )}
            <textarea
              id="research-message"
              className={styles.questionEditor}
              aria-label="对话问题"
              aria-describedby="research-editor-hint"
              placeholder={id ? "继续追问这段对话…" : "输入想研究的题材…"}
              rows={3}
              maxLength={120}
              value={question}
              onChange={e => setQuestion(e.target.value)}
              onKeyDown={e => {
                if (e.key === "Enter" && (e.metaKey || e.ctrlKey) && !e.nativeEvent.isComposing) {
                  e.preventDefault();
                  submit();
                }
              }}
            />
            <div className={styles.editorMeta}>
              <span id="research-editor-hint">拖动右下角调整大小 · Ctrl / ⌘ + Enter 发送</span>
              <span>{question.length} / 120</span>
            </div>
            <div className={styles.composeActions}>
              <Select
                mode="multiple"
                aria-label="本轮 Skill"
                placeholder="选择本轮 Skill（可选）"
                maxCount={8}
                value={selectedSkills}
                optionFilterProp="label"
                onChange={skillIds => editing ? setEditing({ ...editing, skillIds }) : setSelected(skillIds)}
                options={groupSkills(skills.data).filter(group => group.skills.length).map(group => ({
                  label: group.name, options: group.skills.map(skill => ({ label: skill.name, value: skill.id })),
                }))}
              />
              <Button
                type="primary"
                htmlType="submit"
                loading={send.isPending}
                disabled={!question.trim() || skills.isPending || busy || (!!id && !current)}
              >
                {busy ? "研究中" : "发送"}
              </Button>
            </div>
            {skills.isError && (
              <Alert
                type="warning"
                message="Skill 暂时无法加载"
                action={<Button onClick={() => void skills.refetch()}>重试</Button>}
              />
            )}
            {send.isError && <Alert type="error" message={researchError(send.error, "发送失败，请重试")} />}
          </form>
        </section>
      </div>
    </main>
  );
}

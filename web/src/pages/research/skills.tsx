import { useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { Alert, Button, Form, Input, Skeleton, theme } from "antd";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useAuthStore } from "@/stores/auth";
import { withTokenColors } from "../hotspots/panels";
import { listResearchSkills, saveResearchSkill, researchError } from "./research-data";
import type { ResearchSkill } from "./research-data";
import styles from "./research.module.css";

export default function ResearchSkills() {
  const previousRun = Number(useLocation().state?.run) || 0;
  const account = useAuthStore(s => s.user?.id);
  const { token } = theme.useToken();
  const client = useQueryClient();
  const key = ["topic-research", account, "skills"];
  const skills = useQuery({ queryKey: key, queryFn: listResearchSkills });
  const [editing, setEditing] = useState<ResearchSkill | "new">();
  const save = useMutation({
    mutationFn: (values: { name: string; instructions: string }) =>
      saveResearchSkill({ ...values, id: editing && editing !== "new" ? Number(editing.id.split(":")[1]) : undefined }),
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: key });
      setEditing(undefined);
    },
  });
  function edit(skill: ResearchSkill | "new") {
    save.reset();
    setEditing(skill);
  }
  return (
    <main className={styles.settings} style={withTokenColors(token)}>
      <Link to={previousRun ? `/research?run=${previousRun}` : "/research"}>← 返回题材查询</Link>
      <header className={styles.pageHeader}>
        <h1>Skill 管理</h1>
        <Button disabled={save.isPending} onClick={() => edit("new")}>
          新建 Skill
        </Button>
      </header>
      <p>在这里配置研究方法，每轮对话可以自由选用。修改会在下次发送时生效。</p>
      {skills.isPending && <Skeleton active />}
      {skills.isError && (
        <Alert
          type="error"
          message="读取 Skill 失败"
          action={<Button onClick={() => void skills.refetch()}>重试</Button>}
        />
      )}
      {editing && (
        <Form
          key={editing === "new" ? "new" : editing.id}
          layout="vertical"
          className={styles.skillEditor}
          initialValues={editing === "new" ? {} : editing}
          onFinish={v => save.mutate({ name: v.name, instructions: v.instructions })}
        >
          <Form.Item name="name" label="名称" rules={[{ required: true, whitespace: true }]}>
            <Input maxLength={80} />
          </Form.Item>
          <Form.Item name="instructions" label="研究指令" rules={[{ required: true, whitespace: true }]}>
            <Input.TextArea rows={6} maxLength={3000} showCount />
          </Form.Item>
          {save.isError && <Alert type="error" message={researchError(save.error, "保存失败")} />}
          <Button htmlType="submit" type="primary" loading={save.isPending}>
            保存
          </Button>{" "}
          <Button disabled={save.isPending} onClick={() => setEditing(undefined)}>
            取消
          </Button>
        </Form>
      )}
      {skills.data?.map(skill => (
        <article className={styles.skillItem} key={skill.id}>
          <header>
            <h3>{skill.name}</h3>
            <Button size="small" disabled={save.isPending} onClick={() => edit(skill)}>
              编辑
            </Button>
          </header>
          <p>{skill.instructions}</p>
        </article>
      ))}
    </main>
  );
}

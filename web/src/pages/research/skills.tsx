import { useEffect, useRef, useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { Alert, Button, Form, Input, Modal, Skeleton, Space, theme } from "antd";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useAuthStore } from "@/stores/auth";
import SkillInstructions from "./skill-instructions";
import SkillList from "./skill-list";
import { withTokenColors } from "../hotspots/panels";
import { listSkillCatalog, saveSkillGroup, deleteSkillGroup, groupSkills, orderResearchSkills, saveResearchSkill, researchError } from "./research-data";
import type { ResearchSkill, SkillCatalog, SkillGroup, SkillOrder } from "./research-data";
import styles from "./research.module.css";

export default function ResearchSkills() {
  const previousRun = Number(useLocation().state?.run) || 0;
  const account = useAuthStore(s => s.user?.id);
  const { token } = theme.useToken();
  const client = useQueryClient();
  const key = ["topic-research", account, "skill-catalog"];
  const skills = useQuery({ queryKey: key, queryFn: listSkillCatalog });
  const [editing, setEditing] = useState<ResearchSkill | "new">();
  const [groupDraft, setGroupDraft] = useState<SkillGroup>();
  const groupSave = useMutation({
    mutationFn: saveSkillGroup,
    onSuccess: () => setGroupDraft(undefined),
    onSettled: () => client.invalidateQueries({ queryKey: key }),
  });
  const groupDelete = useMutation({
    mutationFn: deleteSkillGroup,
    onSettled: () => client.invalidateQueries({ queryKey: key }),
  });
  function editGroup(group: SkillGroup) {
    groupSave.reset();
    setGroupDraft(group);
  }
  const formArea = useRef<HTMLDivElement>(null);
  useEffect(() => {
    if (editing) formArea.current?.scrollIntoView({ block: "start" });
  }, [editing]);

  const save = useMutation({
    mutationFn: (values: { name: string; instructions: string }) => {
      const original = editing && editing !== "new" ? editing : undefined;
      return saveResearchSkill({
        ...values,
        id: original ? Number(original.id.split(":")[1]) : undefined,
        markdown: !!original?.markdown || values.instructions !== original?.instructions,
      });
    },
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: key });
      setEditing(undefined);
    },
  });
  const order = useMutation({
    mutationFn: orderResearchSkills,
    onMutate: async (next: SkillOrder) => {
      await client.cancelQueries({ queryKey: key });
      const previous = client.getQueryData<SkillCatalog>(key);
      if (previous) client.setQueryData(key, { ...previous, skills: next.skills });
      return { previous };
    },
    onError: (_error, _next, context) => client.setQueryData(key, context?.previous),
    onSettled: () => client.invalidateQueries({ queryKey: key }),
  });
  function edit(skill: ResearchSkill | "new") {
    save.reset();
    setEditing(skill);
  }
  return (
    <main className={styles.settings} style={withTokenColors(token)}>
      <Link to={previousRun ? `/research?run=${previousRun}` : "/research"}>← 返回AI查询</Link>
      <header className={styles.pageHeader}>
        <h1>Skill 管理</h1>
        <Space wrap>
          <Button onClick={() => editGroup({ id: 0, name: "" })}>新建分组</Button>
          <Button disabled={save.isPending} onClick={() => edit("new")}>
            新建 Skill
          </Button>
        </Space>
      </header>
      <p>拖动左侧手柄可调整组内顺序，也可将 Skill 移到其他分组。</p>
      {skills.isPending && <Skeleton active />}
      {skills.isError && (
        <Alert
          type="error"
          message="读取 Skill 失败"
          action={<Button onClick={() => void skills.refetch()}>重试</Button>}
        />
      )}
      <Modal title={groupDraft?.id ? "重命名分组" : "新建分组"} open={!!groupDraft}
        onCancel={() => { if (!groupSave.isPending) setGroupDraft(undefined); }}
        onOk={() => { if (groupDraft?.name.trim() && !groupSave.isPending) groupSave.mutate(groupDraft); }}
        okText="保存" cancelText="取消" confirmLoading={groupSave.isPending} okButtonProps={{ disabled: !groupDraft?.name.trim() }}>
        <Input aria-label="分组名称" placeholder="分组名称" maxLength={60} value={groupDraft?.name ?? ""}
          disabled={groupSave.isPending} onChange={e => groupDraft && setGroupDraft({ ...groupDraft, name: e.target.value })}
          onPressEnter={e => { if (!e.nativeEvent.isComposing && groupDraft?.name.trim() && !groupSave.isPending) groupSave.mutate(groupDraft); }} />
        {groupSave.isError && <Alert type="error" message={researchError(groupSave.error, "保存分组失败")} />}
      </Modal>
      <div ref={formArea} className={styles.editorAnchor}>
        {editing && (
          <Form
            key={editing === "new" ? "new" : editing.id}
            layout="vertical"
            className={styles.skillEditor}
            initialValues={editing === "new" ? {} : editing}
            disabled={save.isPending}
            onFinish={v => save.mutate({ name: v.name, instructions: v.instructions })}
          >
            <Form.Item name="name" label="名称" rules={[{ required: true, whitespace: true }]}>
              <Input maxLength={80} />
            </Form.Item>
            <Form.Item name="instructions" label="研究指令" rules={[
              { required: true, whitespace: true, message: "请输入研究指令" },
              { validator: (_, value: string = "") => Array.from(value).length <= 3000 ? Promise.resolve() : Promise.reject(new Error("含格式标记最多 3000 字")) },
            ]}>
              <SkillInstructions markdown={editing !== "new" && !!editing.markdown} disabled={save.isPending} />
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
      </div>
      {order.isError && <Alert type="error" message={researchError(order.error, "保存排序失败，请重试")} />}
      {groupDelete.isError && <Alert type="error" message={researchError(groupDelete.error, "删除分组失败")} />}
      {skills.data && <SkillList groups={groupSkills(skills.data)} disabled={save.isPending || order.isPending || groupDelete.isPending} edit={edit}
        rename={editGroup} remove={id => groupDelete.mutate(id)}
        reorder={next => order.mutate(next)} />}
    </main>
  );
}

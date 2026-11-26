import { DndContext, KeyboardSensor, PointerSensor, closestCenter, pointerWithin, useDroppable, useSensor, useSensors } from "@dnd-kit/core";
import type { CollisionDetection } from "@dnd-kit/core";
import { SortableContext, arrayMove, rectSortingStrategy, sortableKeyboardCoordinates, useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { HolderOutlined } from "@ant-design/icons";
import { Button, Popconfirm } from "antd";
import { TextPreview } from "@/components/text-preview";
import type { ResearchSkill, SkillGroup, SkillOrder } from "./research-data";
import styles from "./research.module.css";

type ItemProps = { skill: ResearchSkill; disabled: boolean; edit: (skill: ResearchSkill) => void };

function SkillItem({ skill, disabled, edit }: ItemProps) {
  const { attributes, listeners, setNodeRef, setActivatorNodeRef, transform, transition, isDragging } =
    useSortable({ id: skill.id, disabled });
  return <article ref={setNodeRef} className={styles.skillItem} data-dragging={isDragging || undefined}
    style={{ transform: CSS.Transform.toString(transform), transition }}>
    <button type="button" ref={setActivatorNodeRef} className={styles.sortGrip} {...attributes} {...listeners}
      disabled={disabled} aria-label={`拖动排序：${skill.name}`} title="拖动排序"><HolderOutlined /></button>
    <h3><TextPreview text={skill.name} /></h3>
    <TextPreview className={styles.skillDescription} text={skill.instructions} />
    <Button size="small" disabled={disabled} onClick={() => edit(skill)}>编辑</Button>
  </article>;
}

type Group = SkillGroup & { skills: ResearchSkill[] };
type GroupActions = {
  rename: (group: SkillGroup) => void;
  remove: (id: number) => void;
};

function SkillSection({ group, disabled, edit, rename, remove }: Omit<ItemProps, "skill"> & GroupActions & { group: Group }) {
  const { setNodeRef, isOver } = useDroppable({ id: `group:${group.id}`, disabled });
  return <section ref={setNodeRef} aria-label={group.name} className={styles.skillGroup} data-over={isOver || undefined}>
    <header className={styles.groupHeader}>
      <h2>{group.name} <span>{group.skills.length}</span></h2>
      {group.id !== 0 && <div>
        <Button type="text" size="small" onClick={() => rename(group)} disabled={disabled}>重命名</Button>
        {!group.skills.length && <Popconfirm title="删除这个空分组？" okText="删除" cancelText="取消" onConfirm={() => remove(group.id)}>
          <Button type="text" size="small" disabled={disabled}>删除</Button>
        </Popconfirm>}
      </div>}
    </header>
    <SortableContext items={group.skills.map(skill => skill.id)} strategy={rectSortingStrategy}>
      {group.skills.length ? <div className={styles.skillList}>
        {group.skills.map(skill => <SkillItem key={skill.id} skill={skill} disabled={disabled} edit={edit} />)}
      </div> : <div className={styles.emptyGroup}>拖动 Skill 到这里</div>}
    </SortableContext>
  </section>;
}

// Prefer the card under the pointer; the surrounding section also accepts drops in its empty space.
const collision: CollisionDetection = args => {
  if (!args.pointerCoordinates) return closestCenter(args);
  const hits = pointerWithin(args);
  const card = hits.find(hit => String(hit.id).startsWith("custom:"));
  return card ? [card] : hits;
};

export function placeSkill(groups: Group[], activeId: string, overId: string): SkillOrder | undefined {
  const source = groups.find(group => group.skills.some(skill => skill.id === activeId));
  const target = groups.find(group => `group:${group.id}` === overId || group.skills.some(skill => skill.id === overId));
  if (!source || !target || activeId === overId) return;
  const from = source.skills.findIndex(skill => skill.id === activeId);
  const over = target.skills.findIndex(skill => skill.id === overId);
  const moved = { ...source.skills[from], groupId: target.id };
  let ordered: ResearchSkill[];
  if (source.id === target.id) {
    const to = over < 0 ? source.skills.length - 1 : over;
    if (from === to) return;
    ordered = arrayMove(source.skills, from, to);
  } else {
    ordered = target.skills.slice();
    ordered.splice(over < 0 ? ordered.length : over, 0, moved);
  }
  return {
    moved,
    skills: groups.flatMap(group => group.id === target.id ? ordered : group.skills.filter(skill => skill.id !== activeId)),
  };
}

export default function SkillList({ groups, disabled, edit, reorder, rename, remove }: Omit<ItemProps, "skill"> & GroupActions & {
  groups: Group[]; reorder: (next: SkillOrder) => void;
}) {
  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }));
  return <DndContext sensors={sensors} collisionDetection={collision}
    accessibility={{ screenReaderInstructions: { draggable: "按空格开始排序，用方向键移动，再按空格确认或按 Esc 取消。" } }}
    onDragEnd={({ active, over }) => {
      if (!over || disabled) return;
      const next = placeSkill(groups, String(active.id), String(over.id));
      if (next) reorder(next);
    }}>
    {groups.map(group => <SkillSection key={group.id} group={group} disabled={disabled} edit={edit} rename={rename} remove={remove} />)}
  </DndContext>;
}

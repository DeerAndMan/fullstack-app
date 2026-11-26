import { EditorContent, useEditor } from "@tiptap/react";
import type { JSONContent } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import { Markdown } from "@tiptap/markdown";
import { BoldOutlined, OrderedListOutlined, UnorderedListOutlined, UndoOutlined, RedoOutlined } from "@ant-design/icons";
import styles from "./research.module.css";

function paragraph(text: string): JSONContent {
  return { type: "paragraph", content: text ? [{ type: "text", text }] : [] };
}

// Older Skills are plain text. Keep literal markup and show numbered checklists one rule per paragraph.
export function plainInstructions(text: string): JSONContent {
  const content: JSONContent[] = [];
  for (const line of text.split(/\r?\n/)) {
    const section = line.match(/^(\d+ [^。]{1,20}。)\s*(.*)$/u);
    if (section) {
      content.push({ type: "heading", attrs: { level: 3 }, content: [{ type: "text", text: section[1] }] });
      if (!section[2]) continue;
    }
    const body = section ? section[2] : line;
    const rules = /^\d+\.\d+\s/u.test(body) ? body.split(/(?<=。)\s*(?=\d+\.\d+\s)/u) : [body];
    content.push(...rules.map(paragraph));
  }
  return { type: "doc", content };
}

export default function SkillInstructions({ value = "", onChange, markdown, id, disabled = false }: {
  value?: string;
  onChange?: (markdown: string) => void;
  markdown: boolean;
  id?: string;
  disabled?: boolean;
}) {
  const editor = useEditor({
    immediatelyRender: false,
    shouldRerenderOnTransaction: true,
    extensions: [StarterKit.configure({ heading: { levels: [3] }, link: false }), Markdown],
    content: markdown ? value : plainInstructions(value),
    contentType: markdown ? "markdown" : "json",
    editorProps: { attributes: { id: id ?? "skill-instructions", role: "textbox", "aria-label": "研究指令", "aria-multiline": "true" } },
    onUpdate: ({ editor }) => onChange?.(editor.isEmpty ? "" : editor.getMarkdown()),
  });
  if (!editor) return null;
  const tools = [
    { label: "加粗", icon: <BoldOutlined />, active: editor.isActive("bold"), run: () => editor.chain().focus().toggleBold().run() },
    { label: "小标题", icon: "标题", active: editor.isActive("heading"), run: () => editor.chain().focus().toggleHeading({ level: 3 }).run() },
    { label: "编号列表", icon: <OrderedListOutlined />, active: editor.isActive("orderedList"), run: () => editor.chain().focus().toggleOrderedList().run() },
    { label: "项目列表", icon: <UnorderedListOutlined />, active: editor.isActive("bulletList"), run: () => editor.chain().focus().toggleBulletList().run() },
    { label: "撤销", icon: <UndoOutlined />, unavailable: !editor.can().undo(), run: () => editor.chain().focus().undo().run() },
    { label: "重做", icon: <RedoOutlined />, unavailable: !editor.can().redo(), run: () => editor.chain().focus().redo().run() },
  ];
  return <>
    <div className={styles.documentEditor} inert={disabled}>
      <div role="group" aria-label="文本格式" className={styles.formatBar}>
        {tools.map(tool => <button type="button" key={tool.label} title={tool.label} aria-label={tool.label}
          aria-pressed={tool.active} disabled={tool.unavailable}
          onMouseDown={event => event.preventDefault()} onClick={tool.run}>{tool.icon}</button>)}
      </div>
      <EditorContent editor={editor} className={styles.documentBody} />
    </div>
    <div className={styles.formatCount}>含格式标记 {Array.from(value).length} / 3000</div>
  </>;
}

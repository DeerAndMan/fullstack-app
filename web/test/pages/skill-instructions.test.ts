// @vitest-environment jsdom
import { expect, it } from "vitest";
import { Editor } from "@tiptap/react";
import { Markdown } from "@tiptap/markdown";
import StarterKit from "@tiptap/starter-kit";
import { plainInstructions } from "@/pages/research/skill-instructions";

it("旧规则在编辑时分段，加粗和编号列表保存后可继续编辑", () => {
  const extensions = [StarterKit.configure({ heading: { levels: [3] }, link: false }), Markdown];
  const editor = new Editor({ extensions, content: plainInstructions("核查口径\n0 门槛。0.1 利率<3%。0.2 文字保留。") });
  try {
    expect(editor.getJSON().content?.map(block => block.type)).toEqual(["paragraph", "heading", "paragraph", "paragraph"]);
    editor.commands.setTextSelection({ from: 1, to: 5 });
    editor.commands.toggleBold();
    editor.commands.toggleOrderedList();
    const reopened = new Editor({ extensions, content: editor.getMarkdown(), contentType: "markdown" });
    try {
      expect(reopened.getJSON()).toEqual(editor.getJSON());
    } finally {
      reopened.destroy();
    }
  } finally {
    editor.destroy();
  }
});

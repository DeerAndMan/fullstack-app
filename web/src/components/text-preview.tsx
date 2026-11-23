import { Tooltip } from "antd";
import { useRef, useState } from "react";

import styles from "./text-preview.module.css";

import type { ReactNode } from "react";

export function TextPreview({ text, detail, maxWidth, className, children }: {
  text: string;
  detail?: string;
  maxWidth?: number | string;
  className?: string;
  children?: ReactNode;
}) {
  const trigger = useRef<HTMLSpanElement>(null);
  const [open, setOpen] = useState(false);
  return <Tooltip
    trigger={["hover", "focus"]}
    open={open}
    onOpenChange={visible => {
      const element = trigger.current;
      setOpen(visible && Boolean(detail || (element && element.scrollWidth > element.clientWidth)));
    }}
    title={<div className={styles.tooltip}>{text}{detail && <div className={styles.detail}>{detail}</div>}</div>}
    styles={{ root: { maxWidth: "min(36rem, calc(100vw - 24px))" } }}
  >
    <span ref={trigger} className={className ? `${styles.preview} ${className}` : styles.preview}
      style={{ maxWidth }}
      tabIndex={children == null ? 0 : undefined}>
      {children ?? text}
    </span>
  </Tooltip>;
}

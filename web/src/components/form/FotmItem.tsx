import React from "react";
import { Form } from "antd";

import type { FormItemProps } from "antd";

export type BaseFormItemProps = FormItemProps & {
  children: React.ReactNode;
};

/**
 * FotmItem 组件
 * @returns {React.FunctionComponent}
 */
export const FotmItem = (props: BaseFormItemProps) => {
  const { children, ...restFormItemProps } = props;
  return <Form.Item {...restFormItemProps}>{children}</Form.Item>;
};

export default FotmItem;

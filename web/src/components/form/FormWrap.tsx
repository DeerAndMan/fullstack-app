import React, { forwardRef, useImperativeHandle } from "react";
import { Form } from "antd";

import type { FormInstance, FormProps } from "antd";

export interface FormWrapRef<T = any> {
  form: FormInstance<T>;
}

export type BaseFormProps<T = any> = FormProps<T> & {
  children: React.ReactNode;
};

/**
 * FormWrap form表单组件
 * @returns {React.FunctionComponent}
 */
function InnerFormWrap<T>(props: BaseFormProps<T>, ref: React.Ref<FormWrapRef<T>>) {
  const [form] = Form.useForm();

  const { children, ...restFormProps } = props;

  useImperativeHandle(ref, () => ({
    form,
  }));

  return (
    <Form form={form} name="menuForm" layout="vertical" {...restFormProps}>
      {children}
    </Form>
  );
}

export const FormWrap = forwardRef(InnerFormWrap) as <T>(
  props: BaseFormProps<T> & { ref?: React.Ref<FormWrapRef<T>> }
) => React.ReactElement;

export default FormWrap;

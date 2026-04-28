import { Button, Form, InputNumber, Input, Modal } from "antd";

import { useQuerySubscribeHome } from "@/api/subscribe";
import { useBoolean } from "@/hooks/customHooks/useBoolean";

export default function OperableDialog() {
  const { value: isModalOpen, onTrue: showModal, onFalse: hideModal } = useBoolean();
  const { addMutation } = useQuerySubscribeHome();
  const [form] = Form.useForm();

  const handleOk = async () => {
    const values = await form.validateFields();
    addMutation.mutate(values, {
      onSuccess: () => {
        form.resetFields();
        hideModal();
      },
    });
  };

  const handleCancel = () => {
    form.resetFields();
    hideModal();
  };

  return (
    <>
      <Button type="primary" onClick={showModal}>
        添加订阅
      </Button>
      <Modal
        title="添加订阅"
        open={isModalOpen}
        onOk={handleOk}
        onCancel={handleCancel}
        confirmLoading={addMutation.isPending}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            label="用户ID"
            name="user_id"
            rules={[{ required: true, message: "请输入用户ID" }]}
          >
            <InputNumber placeholder="请输入用户ID" style={{ width: "100%" }} controls={false} />
          </Form.Item>
          <Form.Item label="曾用名" name="description">
            <Input placeholder="请输入曾用名（选填）" />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}

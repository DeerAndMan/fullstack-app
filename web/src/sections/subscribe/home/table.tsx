import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button, Form, Input, Modal, Popconfirm, Switch, Table } from "antd";

import { useQuerySubscribeHome } from "@/api/subscribe";
import { useAuthStore } from "@/stores/auth";
import { RoleKey } from "@/types/enum";

import type { SubscribeItem } from "@/types/xq/subscribe/home";

export default function HomeTable() {
  const { queryList, toggleMutation, updateMutation, deleteMutation } = useQuerySubscribeHome();
  const [editingRecord, setEditingRecord] = useState<SubscribeItem | null>(null);
  const [form] = Form.useForm();
  const navigate = useNavigate();
  const role = useAuthStore(s => s.role);
  const isSuperAdmin = role?.role_key === RoleKey.SUPER_ADMIN;

  const handleEdit = (record: SubscribeItem) => {
    setEditingRecord(record);
    form.setFieldsValue({
      description: record.description || "",
      newDescription: "",
    });
  };

  const handleEditOk = async () => {
    if (!editingRecord) return;
    const values = await form.validateFields();
    const newDesc = values.newDescription?.trim();
    if (!newDesc) {
      setEditingRecord(null);
      return;
    }
    updateMutation.mutate(
      { id: editingRecord.id, user_id: editingRecord.user_id, former_name: newDesc },
      {
        onSuccess: () => {
          setEditingRecord(null);
        },
      }
    );
  };

  const handleEditCancel = () => {
    setEditingRecord(null);
  };

  return (
    <>
      <Table
        size="small"
        bordered
        dataSource={queryList.data?.data || []}
        columns={[
          { title: "ID", dataIndex: "id", key: "id" },
          { title: "用户ID", dataIndex: "user_id", key: "user_id" },
          { title: "曾用名", dataIndex: "description", key: "description" },
          {
            title: "状态",
            key: "enabled",
            width: 120,
            align: "center",
            render: (_, record: SubscribeItem) => (
              <Switch
                checked={record.enabled}
                onChange={checked => {
                  toggleMutation.mutate({ user_id: record.user_id, enabled: checked });
                }}
              />
            ),
          },
          {
            title: "操作",
            key: "action",
            width: 220,
            render: (_, record: SubscribeItem) => (
              <>
                <Button type="link" size="small" onClick={() => handleEdit(record)}>
                  编辑
                </Button>
                <Button
                  type="link"
                  size="small"
                  onClick={() => navigate(`/subscribe/detail/${record.id}/${record.user_id}`)}
                >
                  详情
                </Button>
                {isSuperAdmin && (
                  <Popconfirm
                    title="删除订阅"
                    description={`确认删除订阅「${record.description || record.user_id}」？该操作不可恢复。`}
                    okText="确认删除"
                    cancelText="取消"
                    okButtonProps={{ danger: true, loading: deleteMutation.isPending }}
                    onConfirm={() => deleteMutation.mutate({ id: String(record.id), user_id: String(record.user_id) })}
                  >
                    <Button type="link" size="small" danger>
                      删除
                    </Button>
                  </Popconfirm>
                )}
              </>
            ),
          },
        ]}
      />

      <Modal
        title="编辑订阅"
        open={!!editingRecord}
        onOk={handleEditOk}
        onCancel={handleEditCancel}
        confirmLoading={updateMutation.isPending}
      >
        <Form form={form} layout="vertical">
          <Form.Item label="ID">
            <Input value={editingRecord?.id} disabled />
          </Form.Item>
          <Form.Item label="用户ID">
            <Input value={editingRecord?.user_id} disabled />
          </Form.Item>
          <Form.Item label="状态">
            <Input value={editingRecord?.enabled ? "启用" : "禁用"} disabled />
          </Form.Item>
          <Form.Item label="已有曾用名">
            <Input.TextArea value={editingRecord?.description || "无"} disabled autoSize />
          </Form.Item>
          <Form.Item label="添加曾用名" name="newDescription">
            <Input placeholder="请输入新的曾用名" />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}

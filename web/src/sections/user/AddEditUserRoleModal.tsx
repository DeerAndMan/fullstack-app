import { useEffect, useMemo, useState } from "react";
import { Button, Form, message, Modal, Select } from "antd";

import { useAuthStore } from "@/stores/auth";
import { useEnumStore } from "@/stores/enum";
import { getQueryUserList, userRoleMutation } from "@/api/user/user.query";

import type { FormProps } from "antd";

type UserRoleFieldType = {
  roleId: number;
};

interface Props {
  showDefaultTitle?: boolean;
  userId: number | undefined;
  setSelectUserId?: React.Dispatch<React.SetStateAction<number | undefined>>;
}

/**
 * AddEditUserRoleModal 组件 - 用于分配用户角色
 * @returns {React.FunctionComponent}
 */
export const AddEditUserRoleModal = ({
  showDefaultTitle = false,
  userId,
  setSelectUserId,
}: Props) => {
  const roleList = useEnumStore(s => s.roleList);

  const [messageApi, contextHolder] = message.useMessage();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [form] = Form.useForm();

  const queryList = getQueryUserList();

  const { mutate: updateRole } = userRoleMutation({
    successCb: data => {
      useAuthStore.getState().setRole(data);
      messageApi.success("角色分配成功");
      setSelectUserId?.(undefined);
      setIsModalOpen(false);
      form.resetFields();
    },
    errorCb: ({ msg }) => {
      messageApi.error(msg);
    },
  });

  const handleOk = () => {
    form.submit();
  };

  const handleCancel = () => {
    form.resetFields();
    setSelectUserId?.(undefined);
    setIsModalOpen(false);
  };

  const onFinish: FormProps<UserRoleFieldType>["onFinish"] = async values => {
    if (!userId || !values.roleId) return;
    try {
      updateRole({ user_id: userId!, role_id: values.roleId! });
    } catch (error) {
      messageApi.error("角色分配失败");
    }
  };

  const onFinishFailed: FormProps<UserRoleFieldType>["onFinishFailed"] = errorInfo => {
    console.log("Failed:", errorInfo);
  };

  const userInfo = useMemo(() => queryList.find(item => item.id === userId), [userId, queryList]);
  const roleId = useMemo(() => userInfo?.role?.role_id || 3, [userInfo, roleList]);

  useEffect(() => {
    if (!userInfo) return;
    form.setFieldsValue({ roleId: roleId });
    setIsModalOpen(true);
  }, [userInfo, roleId, form]);

  return (
    <div>
      {showDefaultTitle ? <Button type="primary">分配角色</Button> : null}
      {contextHolder}

      <Modal
        title="分配用户角色"
        open={isModalOpen}
        onOk={handleOk}
        onCancel={handleCancel}
        destroyOnClose
      >
        <Form
          form={form}
          name="userRoleForm"
          labelCol={{ span: 8 }}
          wrapperCol={{ span: 16 }}
          style={{ maxWidth: 500 }}
          onFinish={onFinish}
          onFinishFailed={onFinishFailed}
          autoComplete="off"
        >
          <Form.Item<UserRoleFieldType> label="用户名">
            <span className="text-gray-600">{userInfo?.name || "未知用户"}</span>
          </Form.Item>

          <Form.Item<UserRoleFieldType>
            label="角色"
            name="roleId"
            rules={[{ required: true, message: "请选择角色!" }]}
          >
            <Select
              placeholder="请选择角色"
              allowClear
              showSearch
              filterOption={(input, option) =>
                (option?.label ?? "").toLowerCase().includes(input.toLowerCase())
              }
              options={roleList.map(role => ({
                value: role.id,
                label: `${role.name} (${role.role_key})`,
              }))}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default AddEditUserRoleModal;

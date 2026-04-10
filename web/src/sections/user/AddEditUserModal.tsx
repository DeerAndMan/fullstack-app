import { useEffect, useState } from "react";
import { Button, Form, Input, message, Modal, Upload } from "antd";
import { PlusOutlined } from "@ant-design/icons";
import { nanoid } from "nanoid";

import { encryptPassword } from "@/utils";
import { base64ToImg } from "@/utils/img";
import { userApi } from "@/api";
import { uploadAvatar } from "@/api/user";
import { userQuery } from "@/hooks/query/user-query-user";

import type { FormProps, UploadFile } from "antd";
import type { Account, AddAccount } from "@/types/user";

type UserFieldType = {
  avatar?: string;
  name?: string;
  password?: string;
  age?: number;
  email?: string;
  description?: string;
};

interface Props {
  editAccount?: Account | null;
  setEditAccount?: React.Dispatch<React.SetStateAction<Account | null>>;
}

/**
 * AddEditUserModal 组件
 * @returns {React.FunctionComponent}
 */
export const AddEditUserModal = ({ editAccount, setEditAccount }: Props) => {
  const [form] = Form.useForm();
  const [messageApi, contextHolder] = message.useMessage();
  const [isModalOpen, setIsModalOpen] = useState(false);

  const { refreshQuery } = userQuery(false);

  const showModal = () => {
    setIsModalOpen(true);
  };

  const handleOk = () => {
    form.submit();
  };

  const handleCancel = () => {
    form.resetFields();
    setEditAccount?.(null);
    setIsModalOpen(false);
  };

  const onFinish: FormProps<UserFieldType>["onFinish"] = async values => {
    const { avatar } = values;
    delete values.avatar;

    const newValues = {
      ...values,
      age: Number(values.age),
      password: encryptPassword(values.password!),
    } as AddAccount;

    const res = editAccount
      ? await userApi.updateUser({ ...newValues, id: editAccount.id })
      : await userApi.addUser(newValues);

    if (res.code === 200 || res.code === 0) {
      const userId = editAccount ? editAccount?.id : res.data.id;
      if (avatar && userId) {
        await addAvatar(userId, avatar as unknown as UploadFile[]);
      }
      messageApi.success("添加成功");
      refreshQuery();
      setIsModalOpen(false);
    }
  };

  const addAvatar = async (userId: number, files: UploadFile[]) => {
    if (!files.length) return;
    const file = files[0];
    const formData = new FormData();
    formData.append("userId", userId.toString());
    formData.append("name", (file as UploadFile).name);
    formData.append("file", file.response);
    await uploadAvatar(formData);
  };

  const onFinishFailed: FormProps<UserFieldType>["onFinishFailed"] = errorInfo => {
    console.log("Failed:", errorInfo);
  };

  useEffect(() => {
    if (editAccount) {
      let avatar = null;
      if (editAccount.avatar) {
        avatar = {
          uid: nanoid(),
          name: "avatar",
          status: "done",
          url: editAccount.avatar ? base64ToImg(editAccount.avatar) : "",
        };
      }

      form.setFieldsValue({
        name: editAccount.name,
        password: editAccount.password,
        age: editAccount.age,
        email: editAccount.email,
        description: editAccount.description,
        avatar: avatar ? [avatar] : [],
      });
      setIsModalOpen(true);
    }
  }, [editAccount]);

  return (
    <div>
      <Button type="primary" onClick={showModal}>
        {editAccount ? "编辑用户" : "添加用户"}
      </Button>
      {contextHolder}

      <Modal
        title={editAccount ? "编辑用户" : "添加用户"}
        closable={{ "aria-label": "Custom Close Button" }}
        open={isModalOpen}
        onOk={handleOk}
        onCancel={handleCancel}
      >
        <Form
          form={form}
          // clearOnDestroy
          name="basic"
          labelCol={{ span: 8 }}
          wrapperCol={{ span: 16 }}
          style={{ maxWidth: 600 }}
          initialValues={{ password: editAccount ? "" : "112233TT__TT" }}
          onFinish={onFinish}
          onFinishFailed={onFinishFailed}
          autoComplete="off"
        >
          <Form.Item<UserFieldType>
            label="头像"
            name="avatar"
            valuePropName="fileList"
            getValueFromEvent={e => {
              if (Array.isArray(e)) {
                return e;
              }
              return e?.fileList;
            }}
          >
            <Upload
              listType="picture-card"
              maxCount={1}
              showUploadList={{ showDownloadIcon: false }}
              beforeUpload={file => {
                const isJpgOrPng = file.type === "image/jpeg" || file.type === "image/png";
                if (!isJpgOrPng) {
                  messageApi.error("仅支持 JPG/PNG 格式");
                  return false;
                }
                const isLt2M = file.size / 1024 / 1024 < 2;
                if (!isLt2M) {
                  messageApi.error("图片大小不能超过 2MB!");
                  return false;
                }
                return true;
              }}
              customRequest={({ file, onSuccess }) => {
                if (onSuccess) onSuccess(file);
              }}
            >
              <div>
                <PlusOutlined />
                <div style={{ marginTop: 8 }}>上传</div>
              </div>
            </Upload>
          </Form.Item>

          <Form.Item<UserFieldType>
            label="用户名"
            name="name"
            rules={[{ required: true, message: "请输入用户名!" }]}
            normalize={value => value?.trim()}
          >
            <Input />
          </Form.Item>

          <Form.Item<UserFieldType>
            label="密码"
            name="password"
            rules={[{ required: true, message: "请输入密码!" }]}
            normalize={value => value?.trim()}
          >
            <Input.Password />
          </Form.Item>

          <Form.Item<UserFieldType>
            label="年龄"
            name="age"
            rules={[{ required: true, message: "请输入年龄!" }]}
          >
            <Input type="number" />
          </Form.Item>

          <Form.Item<UserFieldType>
            label="邮箱"
            name="email"
            rules={[
              { required: false, message: "请输入邮箱!" },
              { type: "email", message: "请输入有效的邮箱地址!" },
            ]}
            normalize={value => value?.trim()}
          >
            <Input />
          </Form.Item>

          <Form.Item<UserFieldType>
            label="描述"
            name="description"
            rules={[{ required: true, message: "请输入描述!" }]}
          >
            <Input.TextArea rows={4} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default AddEditUserModal;

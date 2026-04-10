import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Avatar, Button, Card, Form, Input, Modal, Space, Upload, message } from "antd";
import { UserOutlined, UploadOutlined, LockOutlined, LogoutOutlined } from "@ant-design/icons";

import { useAuthStore } from "@/stores/auth";
import MenuSetting from "./MenuSetting";
import { base64ToImg } from "@/utils/img";
import { uploadAvatar } from "@/api/user";

import type { UploadProps } from "antd";
import type { RcFile } from "antd/es/upload/interface";

export default function User() {
  const navigate = useNavigate();
  const { user, storeLogout, setUserAvatar } = useAuthStore();

  const [file, setFile] = useState<RcFile | null>(null);
  const [avatar, setAvatar] = useState<string>("");
  const [isModalOpen, setIsModalOpen] = useState<boolean>(false);
  const [isPasswordModalOpen, setIsPasswordModalOpen] = useState<boolean>(false);

  const [userData] = useState({ username: "test_user", createTime: "2023-01-01" });

  const beforeUpload = (file: RcFile) => {
    const isJpgOrPng = file.type === "image/jpeg" || file.type === "image/png";
    if (!isJpgOrPng) {
      message.error("只能上传 JPG/PNG 格式的图片!");
    }
    const isLt1M = file.size / 1024 / 1024 < 1;
    if (!isLt1M) {
      message.error("图片必须小于 1MB!");
    }
    return isJpgOrPng && isLt1M;
  };

  const customUpload: UploadProps["customRequest"] = option => {
    const { file } = option;
    const imageUrl = URL.createObjectURL(file as Blob);
    setAvatar(imageUrl);
    setFile(file as RcFile);
  };

  const handleOk = () => {
    if (!file) return message.error("请选择图片");
    const formData = new FormData();
    formData.append("name", (file as RcFile).name);
    formData.append("file", file);
    uploadAvatar(formData).then(res => {
      if (res.code === 200) {
        setUserAvatar(res.data);
        message.success("上传成功");
        setIsModalOpen(false);
      }
    });
  };

  const handleCancel = () => {
    setIsModalOpen(false);
    setFile(null);
    setAvatar("");
  };

  const handlePasswordChange = (values: {
    oldPassword: string;
    newPassword: string;
    confirmPassword: string;
  }) => {
    if (values.newPassword !== values.confirmPassword) {
      message.error("两次输入的新密码不一致");
      return;
    }
    console.log("修改密码:", values);
    message.success("密码修改成功");
    setIsPasswordModalOpen(false);
  };

  const handleLogout = () => {
    Modal.confirm({
      title: "确认退出",
      content: "确定要退出登录吗？",
      onOk: async () => {
        try {
          await storeLogout();
          navigate("/login");
        } catch {
          message.error("退出登录失败");
        }
      },
    });
  };

  return (
    <div className="p-6">
      <div className="flex flex-col items-center mb-8">
        <Avatar
          size={100}
          src={<img src={user?.avatar ? base64ToImg(user.avatar) : ""} />}
          icon={<UserOutlined />}
          className="cursor-pointer mb-4"
          onClick={() => setIsModalOpen(true)}
        />
      </div>

      <Card title="用户信息" className="mb-6">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <p className="text-gray-500">用户名</p>
            <p className="font-medium">{userData.username}</p>
          </div>
          <div>
            <p className="text-gray-500">注册时间</p>
            <p className="font-medium">{userData.createTime}</p>
          </div>
          <Space direction="horizontal">
            <MenuSetting />
          </Space>
        </div>
      </Card>

      <Card title="安全设置">
        <div className="flex justify-between items-center">
          <div>
            <p className="font-medium">账户密码</p>
            <p className="text-gray-500">定期修改密码可以保护账户安全</p>
          </div>
          <Button
            type="primary"
            icon={<LockOutlined />}
            onClick={() => setIsPasswordModalOpen(true)}
          >
            修改密码
          </Button>
        </div>

        <div className="flex justify-between items-center">
          <div>
            <p className="font-medium">退出登录</p>
            <p className="text-gray-500">退出当前账号</p>
          </div>
          <Button danger icon={<LogoutOutlined />} onClick={handleLogout}>
            退出登录
          </Button>
        </div>
      </Card>

      <Modal
        title="修改头像"
        maskClosable={false}
        open={isModalOpen}
        onOk={handleOk}
        onCancel={handleCancel}
      >
        <div className="flex justify-center items-center">
          <Upload
            name="avatar"
            listType="picture-card"
            className="avatar-uploader"
            showUploadList={false}
            beforeUpload={beforeUpload}
            customRequest={customUpload}
          >
            {avatar ? (
              <img src={avatar} style={{ width: "100%", height: "100%", objectFit: "cover" }} />
            ) : (
              <div className="w-full h-full flex flex-col items-center justify-center">
                <UploadOutlined style={{ fontSize: "24px" }} />
                <div style={{ marginTop: 8 }}>上传</div>
              </div>
            )}
          </Upload>
        </div>
      </Modal>

      <Modal
        title="修改密码"
        open={isPasswordModalOpen}
        onCancel={() => setIsPasswordModalOpen(false)}
        footer={null}
      >
        <Form name="passwordForm" layout="vertical" onFinish={handlePasswordChange}>
          <Form.Item
            name="oldPassword"
            label="当前密码"
            rules={[{ required: true, message: "请输入当前密码" }]}
          >
            <Input.Password />
          </Form.Item>
          <Form.Item
            name="newPassword"
            label="新密码"
            rules={[
              { required: true, message: "请输入新密码" },
              { min: 6, message: "密码长度不能少于6位" },
            ]}
          >
            <Input.Password />
          </Form.Item>
          <Form.Item
            name="confirmPassword"
            label="确认新密码"
            rules={[
              { required: true, message: "请再次输入新密码" },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue("newPassword") === value) {
                    return Promise.resolve();
                  }
                  return Promise.reject(new Error("两次输入的密码不一致"));
                },
              }),
            ]}
          >
            <Input.Password />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" block>
              确认修改
            </Button>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}

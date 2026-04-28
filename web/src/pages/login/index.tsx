import { useNavigate } from "react-router-dom";
import { Button, Checkbox, Form, Input } from "antd";
import { toast } from "react-toastify";
import { setAuthCookie } from "@/utils";
import { useAuthStore } from "@/stores/auth";
import { userApi } from "@/api";

import type { FormProps } from "antd";

type FieldType = {
  username?: string;
  password?: string;
  remember?: string;
};

export const Login = () => {
  const navigator = useNavigate();
  const { setToken, setUser } = useAuthStore();

  const login = (data: FieldType) => {
    if (!data.username || !data.password) return;

    userApi
      .login({ name: data.username.trim(), password: data.password.trim() }, {})
      .then((res) => {
        if (res.code === 0 && res.data) {
          setToken(res.data.token.access_token);
          setUser(res.data.user);
          setAuthCookie(res.data.token.access_token);
          navigator("/", { replace: true });
        } else {
          toast.error(res.msg);
        }
      })
      .catch(() => {});
  };

  const onFinish: FormProps<FieldType>["onFinish"] = (values) => {
    login(values);
  };

  const onFinishFailed: FormProps<FieldType>["onFinishFailed"] = (
    errorInfo,
  ) => {
    console.log("Failed:", errorInfo);
  };

  return (
    <div className="flex h-screen w-full">
      <div className="hidden md:flex md:w-1/2 bg-emerald-600 items-center justify-center">
        <div className="text-center text-white p-8">
          <h1 className="text-4xl font-bold mb-4">欢迎使用系统</h1>
          <p className="text-xl opacity-90">高效、安全的管理平台</p>
        </div>
      </div>

      <div className="w-full md:w-1/2 flex items-center justify-center p-8">
        <div className="w-full max-w-md bg-gray-50 rounded-lg shadow-md p-8 ">
          <h2 className="text-xl font-bold text-center text-gray-800 mb-6">
            用户登录
          </h2>

          <Form
            name="basic"
            labelCol={{ span: 8 }}
            wrapperCol={{ span: 16 }}
            initialValues={{
              username: "test_name",
              password: "112233TT__TT",
              remember: true,
            }}
            onFinish={onFinish}
            onFinishFailed={onFinishFailed}
            autoComplete="off"
            className="w-full dark:text-white"
          >
            <Form.Item<FieldType>
              label="用户名"
              name="username"
              rules={[{ required: true, message: "请输入用户名!" }]}
              normalize={(value) => value?.trim()}
            >
              <Input className="rounded dark:bg-gray-700 dark:text-white dark:border-gray-600" />
            </Form.Item>

            <Form.Item<FieldType>
              label="密码"
              name="password"
              rules={[{ required: true, message: "请输入密码!" }]}
              normalize={(value) => value?.trim()}
            >
              <Input.Password className="rounded dark:bg-gray-700 dark:text-white dark:border-gray-600" />
            </Form.Item>

            <Form.Item<FieldType>
              name="remember"
              valuePropName="checked"
              wrapperCol={{ offset: 8, span: 16 }}
            >
              <Checkbox className="dark:text-gray-300">记住我</Checkbox>
            </Form.Item>

            <Form.Item wrapperCol={{ offset: 8, span: 16 }}>
              <Button
                type="primary"
                htmlType="submit"
                className="w-full bg-blue-600 hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800"
              >
                登录
              </Button>
            </Form.Item>
          </Form>
        </div>
      </div>
    </div>
  );
};

export default Login;

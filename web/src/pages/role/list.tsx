import { Avatar, Button, Table } from "antd";
import { useNavigate } from "react-router-dom";
import { UserOutlined } from "@ant-design/icons";

import ROUTER_PATH from "@/router/section/router-path";
import { base64ToImg } from "@/utils/img";
import { userQuery } from "@/hooks/query/user-query-user";

import type { TableColumnsType } from "antd";
import type { Account, Role } from "@/types/user";

export default function list() {
  const navigate = useNavigate();

  const { queryList } = userQuery();

  const columns: TableColumnsType<Account> = [
    { title: "ID", dataIndex: "id", key: "id" },
    {
      title: "头像",
      dataIndex: "avatar",
      key: "avatar",
      render: (value: string) => (
        <Avatar alt="用户头像" icon={<UserOutlined />} src={base64ToImg(value)} />
      ),
    },
    { title: "账号", dataIndex: "name", key: "name" },
    { title: "年龄", dataIndex: "age", key: "age" },
    { title: "邮箱", dataIndex: "email", key: "email" },
    { title: "描述", dataIndex: "description", key: "description" },
    { title: "权限", dataIndex: "role", key: "role", render: (value: Role) => value?.role_name },
    {
      title: "操作",
      key: "__operation__",
      fixed: "right",
      width: 200,
      render: () => (
        <>
          <Button type="link" size="small" onClick={() => navigate(ROUTER_PATH.user.operation)}>
            编辑
          </Button>
          <Button type="link" size="small" onClick={() => navigate(ROUTER_PATH.role.menu)}>
            权限
          </Button>
        </>
      ),
    },
  ];

  return (
    <div>
      list 权限列表
      <Table
        bordered
        size="small"
        loading={queryList.isLoading}
        columns={columns}
        dataSource={(queryList.data?.data || []).map(l => ({ ...l, key: l.id }))}
      />
    </div>
  );
}

import React, { useState } from "react";
import { Avatar, Button, Space, Table } from "antd";

import AddEditUserModal from "@/sections/user/AddEditUserModal";
import AddEditUserRoleModal from "@/sections/user/AddEditUserRoleModal";
import { base64ToImg } from "@/utils/img";
import { userApi } from "@/api";
import { userQuery } from "@/api/user/user.query";
import { useAuthStore } from "@/stores/auth";
import { useGlobalStore } from "@/stores/global";
import { RoleKey } from "@/types/enum";

import type { Account, Role } from "@/types/user";

export const Operation = () => {
  const { user: stateUser, role: stateRole } = useAuthStore();
  const messageApi = useGlobalStore(s => s.messageApi);

  const [editAccount, setEditAccount] = useState<Account | null>(null);
  const [selectUserId, setSelectUserId] = useState<number | undefined>(undefined);

  const { queryList, refreshQuery } = userQuery();

  const handleChangeRole = (recordId: number) => {
    if (recordId === stateUser?.id || stateRole?.role_key === RoleKey.SUPER_ADMIN) {
      setSelectUserId(recordId);
      return;
    }
    messageApi?.warning("只有本身或管理员才能操作角色");
  };

  const handleEdit = (record: Account) => {
    if (record.id === stateUser?.id || stateRole?.role_key === RoleKey.SUPER_ADMIN) {
      setEditAccount(record);
      return;
    }
    messageApi?.warning("只有本身或管理员才能编辑");
  };

  const handleDelete = (id: number) => {
    if (id === stateUser?.id || stateRole?.role_key === RoleKey.SUPER_ADMIN) {
      userApi.deleteUser(id).then(res => {
        if (res.code === 200) {
          refreshQuery();
        }
      });
      return;
    }
    messageApi?.warning("只有本身或管理员才能删除角色");
  };

  return (
    <React.StrictMode>
      <AddEditUserModal editAccount={editAccount} setEditAccount={setEditAccount} />
      <AddEditUserRoleModal userId={selectUserId} setSelectUserId={setSelectUserId} />

      <Table
        size="small"
        dataSource={(queryList.data?.data || []).map(l => ({ ...l, key: l.id }))}
        columns={[
          { key: "id", title: "ID", dataIndex: "id" },
          {
            key: "avatar",
            title: "头像",
            dataIndex: "avatar",
            render: (text: string) => (
              <Avatar src={<img src={text ? base64ToImg(text) : ""} />} className="cursor-pointer">
                无
              </Avatar>
            ),
          },
          { key: "name", title: "姓名", dataIndex: "name" },
          { key: "age", title: "年龄", dataIndex: "age" },
          { key: "email", title: "邮箱", dataIndex: "email" },
          { key: "description", title: "描述", dataIndex: "description" },
          { key: "createdAt", title: "创建时间", dataIndex: "createdAt" },
          { key: "updatedAt", title: "更新时间", dataIndex: "updatedAt" },
          {
            key: "role",
            title: "角色",
            dataIndex: "role",
            render: (text: Role, record: Account) => {
              if (text && text.role_id) {
                return (
                  <div className="flex items-center gap-2">
                    <Button size="small" type="text" onClick={() => handleChangeRole(record.id)}>
                      {text.role_name || "普通用户"}
                    </Button>
                  </div>
                );
              }
              return (
                <Button size="small" type="text" onClick={() => handleChangeRole(record.id)}>
                  普通用户
                </Button>
              );
            },
          },
          {
            key: "action",
            title: "操作",
            render: data => (
              <Space>
                <Button size="small" type="primary" onClick={() => handleEdit(data)}>
                  编辑
                </Button>
                <Button size="small" type="default" onClick={() => handleDelete(data.id)}>
                  删除
                </Button>
              </Space>
            ),
          },
        ]}
      />
    </React.StrictMode>
  );
};

export default Operation;

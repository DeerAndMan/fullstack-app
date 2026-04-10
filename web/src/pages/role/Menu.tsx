import { useEffect, useRef, useState } from "react";
import { Space, Select, Button, Tree } from "antd";

import { useAuthStore } from "@/stores/auth";
import { useEnumStore } from "@/stores/enum";
import { useGlobalStore } from "@/stores/global";
import {
  getRoleRoutingByRoleIdQuery,
  menuListQuery,
  roleRoutingMutation,
} from "@/hooks/query/use-query-menu";
import { FormItem, FormWrap } from "@/components/form";
import { buildMenuTree } from "@/utils/treeTransform";

import type { TreeProps, SelectProps } from "antd";
import type { FormWrapRef } from "@/components/form/FormWrap";

interface FormType {
  roleId: number;
  menuIds: string[];
}

/**
 * Menu 组件
 * @returns {React.FunctionComponent}
 */
const Menu = () => {
  // const menuRoles = useAuthStore(s => s.menuRoles);
  const messageApi = useGlobalStore(s => s.messageApi);
  const userRoleStore = useAuthStore(s => s.role);
  const roleList = useEnumStore(s => s.roleList);
  const formRef = useRef<FormWrapRef<FormType>>(null);

  const [roldId, setroldId] = useState(0);
  const [autoExpandParent, setAutoExpandParent] = useState<boolean>(true);
  const [checkedKeys, setCheckedKeys] = useState<React.Key[]>([]);
  const [selectedKeys, setSelectedKeys] = useState<React.Key[]>([]);

  const { queryList } = menuListQuery();
  const { mutate } = roleRoutingMutation();
  const queryData = getRoleRoutingByRoleIdQuery(roldId);
  // console.log("queryData", queryData.data);

  const handleSelectChange: SelectProps["onChange"] = value => {
    setroldId(value);
    formRef.current?.form.setFieldsValue({ menuIds: [] });
    setCheckedKeys([]);
    setSelectedKeys([]);
  };

  const onExpand: TreeProps["onExpand"] = _expandedKeysValue => {
    setAutoExpandParent(false);
  };

  const onCheck: TreeProps["onCheck"] = (checked, _halfChecked) => {
    formRef.current?.form.setFieldsValue({ menuIds: checked as string[] });
    setCheckedKeys(checked as React.Key[]);
  };

  const onSelect: TreeProps["onSelect"] = (selectedKeysValue, info) => {
    console.log("onSelect", selectedKeysValue, info);
    setSelectedKeys(selectedKeysValue);
  };
  const onFinish = (values: FormType) => {
    console.log("onFinish values", values);
    const menuIds = values.menuIds
      .map(item => queryList.data?.find(s => s.menu_code === item)?.id || "")
      .filter(id => id);

    mutate({ ...values, menuIds });
  };

  useEffect(() => {
    if (!userRoleStore?.role_id) return;
    setroldId(userRoleStore.role_id);
    formRef.current?.form.setFieldsValue({ roleId: userRoleStore.role_id });
  }, [userRoleStore?.role_id]);

  useEffect(() => {
    messageApi?.open({
      key: "role_menu_select_list",
      type: "loading",
      content: "加载角色菜单列表...",
      duration: 0,
    });

    if (!queryData.data) return;
    const keys = queryData.data.data.map(item => item.menu_code);
    formRef.current?.form.setFieldsValue({ menuIds: keys });
    setCheckedKeys(keys);
    messageApi?.destroy("role_menu_select_list");
  }, [queryData.data, messageApi]);

  return (
    <Space wrap direction="vertical" className="w-full">
      <h1 className="text-xl font-bold">角色菜单列表设置</h1>

      <FormWrap<FormType> ref={formRef} onFinish={onFinish}>
        <FormItem name="roleId" label="角色" rules={[{ required: true }]}>
          <Select
            // defaultValue={1}
            style={{ width: 160 }}
            options={roleList.map(l => ({ value: l.id, label: l.name }))}
            onChange={handleSelectChange}
          />
        </FormItem>
        <FormItem name="menuIds" label="菜单列表" rules={[{ required: true }]}>
          <Tree
            defaultExpandAll
            checkable
            onExpand={onExpand}
            autoExpandParent={autoExpandParent}
            onCheck={onCheck}
            checkedKeys={checkedKeys}
            onSelect={onSelect}
            selectedKeys={selectedKeys}
            treeData={buildMenuTree(queryList.data || [])}
          />
        </FormItem>
        <FormItem>
          <Button type="primary" htmlType="submit">
            确认
          </Button>
        </FormItem>
      </FormWrap>
    </Space>
  );
};

export default Menu;

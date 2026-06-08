import { useCallback, useEffect, useRef, useState } from "react";
import { Button, Modal, Tree } from "antd";
import { MenuOutlined } from "@ant-design/icons";

import { navRouter } from "@/router";
import { addMenuRouter } from "@/api/menu/menu.service";
import { menuListQuery } from "@/api/menu";

import type { TreeDataNode, TreeProps } from "antd";
import type { NavRouter, TreeDataType } from "@/router";

/**
 * MenuSetting 组件
 * @returns {React.FunctionComponent}
 */
export const MenuSetting = () => {
  const selectedData = useRef<React.Key[]>([]);

  const [open, setOpen] = useState(false);
  const [autoExpandParent, setAutoExpandParent] = useState<boolean>(true);
  const [checkedKeys, setCheckedKeys] = useState<React.Key[]>([]);
  const [selectedKeys, setSelectedKeys] = useState<React.Key[]>([]);

  const { queryList, refresh } = menuListQuery();

  const transformNode = useCallback(
    (routeData: NavRouter[]): TreeDataType[] => {
      if (queryList.isPending) return [];
      selectedData.current = [];

      const transform = (node: NavRouter): TreeDataType => {
        const key = node.id;
        const findItem = (queryList.data || []).find(s => s.menu_code === key);

        if (findItem) {
          selectedData.current.push(key);
        }

        const menuItem: TreeDataType = {
          key,
          title: node.name,
          link_url: node.path === "" ? `menu-${node.name}` : node.path,
        };
        if (node.children && node.children.length > 0) {
          return { ...menuItem, children: node.children.map(child => transform(child)) };
        }
        return menuItem;
      };

      return routeData.map(node => transform(node));
    },
    [queryList.isPending, queryList.data]
  );
  const menuTreeData = transformNode(navRouter);

  const onExpand: TreeProps["onExpand"] = _expandedKeysValue => {
    setAutoExpandParent(false);
  };

  const onCheck: TreeProps["onCheck"] = checkedKeysValue => {
    setCheckedKeys(checkedKeysValue as React.Key[]);
  };

  const onSelect: TreeProps["onSelect"] = (selectedKeysValue, _info) => {
    setSelectedKeys(selectedKeysValue);
  };

  const handleOk = () => {
    const findList = checkedKeys
      .map(c => menuTreeData.find(t => t.key === c))
      .filter(t => t) as TreeDataNode[];

    addMenuRouter(findList)
      .then(res => {
        if (res.code !== 0) return;
        setOpen(false);
        setCheckedKeys([]);
        refresh();
      })
      .catch(rej => {
        console.error("rej", rej);
      });
  };

  useEffect(() => {
    if (!open) return;
    setCheckedKeys(selectedData.current);

    return () => {
      selectedData.current = [];
    };
  }, [open, queryList.data]);

  return (
    <>
      <Button
        type="dashed"
        color="default"
        icon={<MenuOutlined />}
        onClick={() => setOpen(prev => !prev)}
      >
        菜单列表设置
      </Button>

      <Modal title="菜单列表设置" open={open} onCancel={() => setOpen(false)} onOk={handleOk}>
        <div className="mb-3">
          <p className="font-medium">菜单列表</p>
          <p className="text-gray-500">选择菜单列表，用户登录后将显示该菜单</p>
        </div>

        <Tree
          // showLine
          defaultExpandAll
          checkable
          onExpand={onExpand}
          autoExpandParent={autoExpandParent}
          onCheck={onCheck}
          checkedKeys={checkedKeys}
          onSelect={onSelect}
          selectedKeys={selectedKeys}
          treeData={transformNode(navRouter)}
          // treeData={treeData}
        />
      </Modal>
    </>
  );
};

export default MenuSetting;

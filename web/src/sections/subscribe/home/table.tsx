import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Alert, Button, Checkbox, Form, Input, InputNumber, Modal, Popconfirm, Select, Space, Switch, Table, Tooltip, Typography } from "antd";
import { ReloadOutlined, SearchOutlined, StarFilled, StarOutlined } from "@ant-design/icons";

import { useQuerySubscribeHome } from "@/api/subscribe";
import { useAuthStore } from "@/stores/auth";
import { RoleKey } from "@/types/enum";

import type { SubscribeItem } from "@/types/xq/subscribe/home";
import { selectSubscriptions, type SubscriptionSort, type SubscriptionStatus } from "./list-utils";

export default function HomeTable() {
  const { queryList, toggleMutation, updateMutation, deleteMutation, preferencesMutation } = useQuerySubscribeHome();
  const [editingRecord, setEditingRecord] = useState<SubscribeItem | null>(null);
  const [sortingRecord, setSortingRecord] = useState<SubscribeItem | null>(null);
  const [sortValue, setSortValue] = useState<number | null>(0);
  const [search, setSearch] = useState("");
  const [status, setStatus] = useState<SubscriptionStatus>("all");
  const [starredOnly, setStarredOnly] = useState(false);
  const [sort, setSort] = useState<SubscriptionSort>("priority");
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [form] = Form.useForm();
  const navigate = useNavigate();
  const role = useAuthStore(s => s.role);
  const isSuperAdmin = role?.role_key === RoleKey.SUPER_ADMIN;
  const items = useMemo(
    () => selectSubscriptions(queryList.data?.data || [], search, status, starredOnly, sort),
    [queryList.data, search, status, starredOnly, sort]
  );
  const currentPage = Math.min(page, Math.max(1, Math.ceil(items.length / pageSize)));
  const validSortValue = sortValue !== null && Number.isInteger(sortValue) && sortValue >= 0 && sortValue <= 999999;

  const resetFilters = () => {
    setSearch("");
    setStatus("all");
    setStarredOnly(false);
    setSort("priority");
    setPage(1);
  };

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
      <div className="flex flex-wrap items-center gap-3">
        <Input
          aria-label="搜索订阅"
          placeholder="搜索 ID、用户ID、曾用名"
          prefix={<SearchOutlined />}
          allowClear
          value={search}
          onChange={event => { setSearch(event.target.value); setPage(1); }}
          style={{ width: 280, maxWidth: "100%" }}
        />
        <Select<SubscriptionStatus>
          aria-label="订阅状态"
          value={status}
          style={{ width: 120 }}
          options={[
            { value: "all", label: "全部状态" },
            { value: "enabled", label: "已启用" },
            { value: "disabled", label: "已禁用" },
          ]}
          onChange={value => { setStatus(value); setPage(1); }}
        />
        <Checkbox checked={starredOnly} onChange={event => { setStarredOnly(event.target.checked); setPage(1); }}>
          只看星标
        </Checkbox>
        <Select<SubscriptionSort>
          aria-label="列表排序方式"
          value={sort}
          style={{ width: 160 }}
          options={[
            { value: "priority", label: "星标 / 手动排序" },
            { value: "idAsc", label: "ID 从小到大" },
            { value: "idDesc", label: "ID 从大到小" },
            { value: "name", label: "曾用名排序" },
          ]}
          onChange={value => { setSort(value); setPage(1); }}
        />
        <Button onClick={resetFilters}>重置</Button>
        <Button icon={<ReloadOutlined />} loading={queryList.isFetching} onClick={() => void queryList.refetch()}>
          刷新
        </Button>
      </div>
      <Typography.Text type="secondary">星标订阅优先，同组内排序值越大越靠前；相同排序值按 ID 排列。</Typography.Text>
      {queryList.isError && (
        <Alert type="error" showIcon message="订阅列表加载失败" description={queryList.error.message} />
      )}
      <Table
        size="small"
        bordered
        rowKey={record => `${record.id}-${record.user_id}`}
        loading={queryList.isFetching}
        dataSource={items}
        scroll={{ x: 1100 }}
        locale={{ emptyText: search || status !== "all" || starredOnly ? "没有符合条件的订阅" : "暂无订阅，请先添加订阅" }}
        pagination={{
          current: currentPage,
          pageSize,
          total: items.length,
          showSizeChanger: true,
          showQuickJumper: true,
          pageSizeOptions: [10, 20, 50, 100],
          showTotal: total => `共 ${total} 条，全部订阅 ${queryList.data?.data.length ?? 0} 条`,
          onChange: (nextPage, nextPageSize) => {
            setPage(nextPageSize !== pageSize ? 1 : nextPage);
            setPageSize(nextPageSize);
          },
        }}
        columns={[
          {
            title: "星标", key: "is_starred", width: 64, align: "center",
            render: (_, record: SubscribeItem) => (
              <Tooltip title={record.is_starred ? "取消星标" : "标记为重点订阅"}>
                <Button
                  type="text"
                  aria-label={`${record.is_starred ? "取消星标" : "添加星标"}：${record.description || record.user_id}`}
                  icon={record.is_starred ? <StarFilled style={{ color: "#faad14" }} /> : <StarOutlined />}
                  disabled={preferencesMutation.isPending}
                  onClick={() => preferencesMutation.mutate({
                    id: record.id, user_id: record.user_id, is_starred: !record.is_starred,
                  }, {
                    onSuccess: () => { setSort("priority"); setPage(1); },
                  })}
                />
              </Tooltip>
            ),
          },
          {
            title: "排序值", key: "sort_order", width: 100, align: "center",
            render: (_, record: SubscribeItem) => (
              <Button type="link" size="small" disabled={preferencesMutation.isPending} onClick={() => {
                setSortingRecord(record);
                setSortValue(record.sort_order);
              }}>
                {record.sort_order} · 设置
              </Button>
            ),
          },
          { title: "ID", dataIndex: "id", key: "id", width: 210, render: (value: string) => <Typography.Text copyable>{value}</Typography.Text> },
          { title: "用户ID", dataIndex: "user_id", key: "user_id", width: 160, render: (value: string) => <Typography.Text copyable>{value}</Typography.Text> },
          { title: "曾用名", dataIndex: "description", key: "description", render: (value: string | null) => value || "-" },
          {
            title: "状态",
            key: "enabled",
            width: 120,
            align: "center",
            render: (_, record: SubscribeItem) => (
              <Switch
                checked={record.enabled}
                checkedChildren="启用"
                unCheckedChildren="禁用"
                disabled={toggleMutation.isPending}
                loading={toggleMutation.isPending && toggleMutation.variables?.user_id === record.user_id}
                onChange={checked => {
                  toggleMutation.mutate({ user_id: record.user_id, enabled: checked });
                }}
              />
            ),
          },
          {
            title: "操作",
            key: "action",
            width: 200,
            fixed: "right",
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
                    <Button type="link" size="small" danger disabled={deleteMutation.isPending}>
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
        title="设置手动排序"
        open={!!sortingRecord}
        confirmLoading={preferencesMutation.isPending}
        okButtonProps={{ disabled: !validSortValue }}
        onCancel={() => { if (!preferencesMutation.isPending) setSortingRecord(null); }}
        onOk={() => {
          if (!sortingRecord || !validSortValue || sortValue === null) return;
          preferencesMutation.mutate({
            id: sortingRecord.id, user_id: sortingRecord.user_id, sort_order: sortValue,
          }, {
            onSuccess: () => {
              setSortingRecord(null);
              setSort("priority");
              setPage(1);
            },
          });
        }}
      >
        <Space direction="vertical" size="middle" style={{ width: "100%" }}>
          <Typography.Text>{sortingRecord?.description || sortingRecord?.user_id}</Typography.Text>
          <Typography.Text type="secondary">输入 0–999999 的整数，数值越大越靠前；0 为默认排序。星标订阅始终优先于未星标订阅。</Typography.Text>
          <InputNumber
            aria-label="排序值"
            min={0} max={999999} precision={0}
            value={sortValue} onChange={setSortValue}
            disabled={preferencesMutation.isPending}
            style={{ width: "100%" }}
          />
        </Space>
      </Modal>

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

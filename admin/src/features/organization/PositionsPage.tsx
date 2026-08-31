import { useQueryClient } from "@tanstack/react-query";
import { Button, Card, Form, Input, message, Space, Table, Tag } from "antd";
import {
  getListPositionsQueryKey,
  useCreatePosition,
  useListPositions,
} from "../../api/generated/modura";
import { useAuth } from "../auth/auth-context";
import { usePermissions } from "../workspace/use-permissions";

export function PositionsPage() {
  const auth = useAuth();
  const granted = usePermissions();
  const client = useQueryClient();
  const query = useListPositions({ fetch: auth.fetchOptions });
  const positions = query.data?.status === 200 ? query.data.data : [];
  const create = useCreatePosition({
    fetch: {
      ...auth.fetchOptions,
      headers: { ...auth.fetchOptions.headers, "X-CSRF-Token": auth.csrfToken },
    },
    mutation: {
      onSuccess: async (response) => {
        if (response.status === 201) {
          message.success("岗位已创建");
          await client.invalidateQueries({
            queryKey: getListPositionsQueryKey(),
          });
        } else message.error("创建失败");
      },
    },
  });
  return (
    <Space direction="vertical" size="large" className="workspace">
      {granted.has("organization.positions/create") && (
        <Card title="创建岗位">
          <Form layout="inline" onFinish={(data) => create.mutate({ data })}>
            <Form.Item name="name" label="名称" rules={[{ required: true }]}>
              <Input maxLength={128} />
            </Form.Item>
            <Button type="primary" htmlType="submit" loading={create.isPending}>
              创建
            </Button>
          </Form>
        </Card>
      )}
      <Card title="岗位">
        <Table
          rowKey="id"
          loading={query.isLoading}
          dataSource={positions}
          pagination={false}
          columns={[
            { title: "名称", dataIndex: "name" },
            {
              title: "状态",
              render: (_, item) => (
                <Tag color={item.status === "active" ? "green" : "default"}>
                  {item.status === "active" ? "启用" : "停用"}
                </Tag>
              ),
            },
          ]}
        />
      </Card>
    </Space>
  );
}

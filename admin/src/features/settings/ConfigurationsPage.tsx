import { Card, Table, Tag } from "antd";

import { useListConfigurations } from "../../api/generated/modura";
import { useAuth } from "../auth/auth-context";

export function ConfigurationsPage() {
  const auth = useAuth();
  const query = useListConfigurations({ fetch: auth.fetchOptions });
  const configurations = query.data?.status === 200 ? query.data.data : [];

  return (
    <Card title="系统配置">
      <Table
        loading={query.isLoading}
        rowKey="key"
        dataSource={configurations}
        pagination={false}
        columns={[
          { title: "配置", dataIndex: "name" },
          { title: "键", dataIndex: "key" },
          { title: "类型", dataIndex: "valueType" },
          {
            title: "当前值",
            dataIndex: "value",
            render: (value) => JSON.stringify(value),
          },
          {
            title: "来源",
            dataIndex: "source",
            render: (source) => (
              <Tag>{source === "tenant" ? "租户" : "全局"}</Tag>
            ),
          },
        ]}
      />
    </Card>
  );
}

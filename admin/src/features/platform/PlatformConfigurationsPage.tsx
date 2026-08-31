import { Card, Table, Tag } from "antd";
import { useListPlatformConfigurations } from "../../api/generated/modura";
import { usePlatformAuth } from "./platform-auth-context";

export function PlatformConfigurationsPage() {
  const auth = usePlatformAuth();
  const query = useListPlatformConfigurations({ fetch: auth.fetchOptions });
  const rows = query.data?.status === 200 ? query.data.data : [];
  return (
    <Card title="全局非密配置">
      <Table
        rowKey="key"
        loading={query.isLoading}
        dataSource={rows}
        pagination={false}
        columns={[
          { title: "名称", dataIndex: "name" },
          { title: "键", dataIndex: "key" },
          { title: "类型", dataIndex: "valueType" },
          { title: "值", render: (_, row) => JSON.stringify(row.value) },
          {
            title: "租户可覆盖",
            render: (_, row) => (
              <Tag color={row.tenantOverridable ? "blue" : "default"}>
                {row.tenantOverridable ? "是" : "否"}
              </Tag>
            ),
          },
        ]}
      />
    </Card>
  );
}

import { Card, Empty, List, Space, Tag, Typography } from "antd";
import { useListPlatformDictionaries } from "../../api/generated/modura";
import { usePlatformAuth } from "./platform-auth-context";

export function PlatformDictionariesPage() {
  const auth = usePlatformAuth();
  const query = useListPlatformDictionaries({ fetch: auth.fetchOptions });
  const rows = query.data?.status === 200 ? query.data.data : [];
  return (
    <Card title="全局字典" loading={query.isLoading}>
      {rows.length === 0 ? (
        <Empty />
      ) : (
        <List
          dataSource={rows}
          renderItem={(row) => (
            <List.Item>
              <List.Item.Meta
                title={
                  <Space>
                    <Typography.Text>{row.name}</Typography.Text>
                    <Tag>{row.code}</Tag>
                  </Space>
                }
                description={row.items
                  .map((item) => `${item.label} (${item.code})`)
                  .join(" · ")}
              />
            </List.Item>
          )}
        />
      )}
    </Card>
  );
}

import { Card, Empty, List, Space, Tag, Typography } from "antd";

import { useListDictionaries } from "../../api/generated/modura";
import { useAuth } from "../auth/auth-context";

export function DictionariesPage() {
  const auth = useAuth();
  const query = useListDictionaries({ fetch: auth.fetchOptions });
  const dictionaries = query.data?.status === 200 ? query.data.data : [];

  return (
    <Card loading={query.isLoading} title="字典管理">
      {dictionaries.length === 0 ? (
        <Empty description="暂无有效字典" />
      ) : (
        <List
          dataSource={dictionaries}
          renderItem={(dictionary) => (
            <List.Item>
              <List.Item.Meta
                title={
                  <Space>
                    <Typography.Text>{dictionary.name}</Typography.Text>
                    <Tag>{dictionary.code}</Tag>
                    <Tag
                      color={
                        dictionary.source === "tenant" ? "blue" : "default"
                      }
                    >
                      {dictionary.source === "tenant" ? "租户" : "全局"}
                    </Tag>
                  </Space>
                }
                description={dictionary.items
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

import { Card, Table, Tag } from "antd";

import { useListAuditEvents } from "../../api/generated/modura";
import { useAuth } from "../auth/auth-context";

export function AuditPage() {
  const auth = useAuth();
  const query = useListAuditEvents(
    { limit: 50, offset: 0 },
    { fetch: auth.fetchOptions },
  );
  const events = query.data?.status === 200 ? query.data.data : [];

  return (
    <Card title="审计日志">
      <Table
        loading={query.isLoading}
        rowKey="id"
        dataSource={events}
        pagination={false}
        columns={[
          { title: "时间", dataIndex: "occurredAt" },
          { title: "操作", dataIndex: "action" },
          { title: "资源", dataIndex: "resource" },
          { title: "原因", dataIndex: "reason" },
          {
            title: "结果",
            dataIndex: "result",
            render: (result) => (
              <Tag color={result === "succeeded" ? "green" : "red"}>
                {result}
              </Tag>
            ),
          },
        ]}
      />
    </Card>
  );
}

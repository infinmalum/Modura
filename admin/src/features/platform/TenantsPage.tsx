import { useQueryClient } from "@tanstack/react-query";
import {
  Button,
  Card,
  Form,
  Input,
  message,
  Modal,
  Space,
  Table,
  Tag,
} from "antd";
import {
  getListPlatformTenantsQueryKey,
  useListPlatformTenants,
  useProvisionPlatformTenant,
  useReactivatePlatformTenant,
  useSuspendPlatformTenant,
} from "../../api/generated/modura";
import { usePlatformAuth } from "./platform-auth-context";

export function TenantsPage() {
  const auth = usePlatformAuth();
  const client = useQueryClient();
  const query = useListPlatformTenants({ fetch: auth.fetchOptions });
  const tenants = query.data?.status === 200 ? query.data.data : [];
  const writeFetch = {
    ...auth.fetchOptions,
    headers: {
      ...auth.fetchOptions.headers,
      "X-CSRF-Token": auth.csrfToken,
      "Idempotency-Key": crypto.randomUUID(),
    },
  };
  const refresh = () =>
    client.invalidateQueries({ queryKey: getListPlatformTenantsQueryKey() });
  const provision = useProvisionPlatformTenant({
    fetch: writeFetch,
    mutation: {
      onSuccess: async (response) => {
        if (response.status === 200 || response.status === 201) {
          message.success(
            response.data.created ? "租户已创建" : "幂等请求已返回原结果",
          );
          await refresh();
        } else message.error("创建租户失败");
      },
    },
  });
  const suspend = useSuspendPlatformTenant({
    fetch: writeFetch,
    mutation: {
      onSuccess: async (response) => {
        if (response.status === 204) {
          message.success("租户已暂停");
          await refresh();
        } else message.error("操作失败");
      },
    },
  });
  const reactivate = useReactivatePlatformTenant({
    fetch: writeFetch,
    mutation: {
      onSuccess: async (response) => {
        if (response.status === 204) {
          message.success("租户已恢复");
          await refresh();
        } else message.error("操作失败");
      },
    },
  });
  const lifecycle = (tenantId: string, action: "suspend" | "reactivate") => {
    let reason = "";
    Modal.confirm({
      title: action === "suspend" ? "暂停租户" : "恢复租户",
      content: (
        <Input.TextArea
          placeholder="请输入审计原因"
          onChange={(event) => {
            reason = event.target.value;
          }}
        />
      ),
      onOk: () => {
        if (!reason.trim()) {
          message.warning("必须填写原因");
          return Promise.reject();
        }
        return action === "suspend"
          ? suspend.mutateAsync({ tenantId, data: { reason } })
          : reactivate.mutateAsync({ tenantId, data: { reason } });
      },
    });
  };
  return (
    <Space direction="vertical" size="large" className="workspace">
      <Card title="创建租户">
        <Form layout="vertical" onFinish={(data) => provision.mutate({ data })}>
          <Space wrap align="start">
            <Form.Item
              name="slug"
              label="租户标识"
              rules={[{ required: true }]}
            >
              <Input />
            </Form.Item>
            <Form.Item
              name="displayName"
              label="显示名称"
              rules={[{ required: true }]}
            >
              <Input />
            </Form.Item>
            <Form.Item
              name="rootDepartmentName"
              label="根部门"
              rules={[{ required: true }]}
            >
              <Input />
            </Form.Item>
            <Form.Item
              name="administratorUsername"
              label="首位管理员"
              rules={[{ required: true }]}
            >
              <Input />
            </Form.Item>
            <Form.Item name="administratorEmail" label="管理员邮箱">
              <Input />
            </Form.Item>
            <Form.Item
              name="reason"
              label="创建原因"
              rules={[{ required: true }]}
            >
              <Input />
            </Form.Item>
          </Space>
          <Button
            type="primary"
            htmlType="submit"
            loading={provision.isPending}
          >
            原子创建租户
          </Button>
        </Form>
      </Card>
      <Card title="租户">
        <Table
          rowKey="id"
          loading={query.isLoading}
          dataSource={tenants}
          pagination={false}
          columns={[
            { title: "名称", dataIndex: "displayName" },
            { title: "标识", dataIndex: "slug" },
            {
              title: "状态",
              render: (_, tenant) => (
                <Tag
                  color={
                    tenant.status === "active"
                      ? "green"
                      : tenant.status === "suspended"
                        ? "orange"
                        : "blue"
                  }
                >
                  {tenant.status}
                </Tag>
              ),
            },
            {
              title: "操作",
              render: (_, tenant) =>
                tenant.status === "active" ? (
                  <Button
                    danger
                    onClick={() => lifecycle(tenant.id, "suspend")}
                  >
                    暂停
                  </Button>
                ) : tenant.status === "suspended" ? (
                  <Button onClick={() => lifecycle(tenant.id, "reactivate")}>
                    恢复
                  </Button>
                ) : null,
            },
          ]}
        />
      </Card>
    </Space>
  );
}

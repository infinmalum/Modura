import {
  Button,
  Card,
  Form,
  Input,
  message,
  Select,
  Space,
  Typography,
} from "antd";
import { useState } from "react";
import {
  useAssignUserOrganization,
  useGetUserRoleGrants,
  useListDepartments,
  useListPositions,
  useListRoles,
  useReplaceUserRoleGrants,
} from "../../api/generated/modura";
import { useAuth } from "../auth/auth-context";
import { usePermissions } from "../workspace/use-permissions";

export function UserAssignmentsPage() {
  const [userId, setUserId] = useState("");
  const auth = useAuth();
  const granted = usePermissions();
  const departmentsQuery = useListDepartments({ fetch: auth.fetchOptions });
  const positionsQuery = useListPositions({ fetch: auth.fetchOptions });
  const rolesQuery = useListRoles({ fetch: auth.fetchOptions });
  const grantsQuery = useGetUserRoleGrants(userId, {
    fetch: auth.fetchOptions,
    query: { enabled: Boolean(userId) },
  });
  const departments =
    departmentsQuery.data?.status === 200 ? departmentsQuery.data.data : [];
  const positions =
    positionsQuery.data?.status === 200 ? positionsQuery.data.data : [];
  const roles =
    rolesQuery.data?.status === 200
      ? rolesQuery.data.data.filter((role) => !role.reserved)
      : [];
  const grants =
    grantsQuery.data?.status === 200 ? grantsQuery.data.data : undefined;
  const writeFetch = {
    ...auth.fetchOptions,
    headers: { ...auth.fetchOptions.headers, "X-CSRF-Token": auth.csrfToken },
  };
  const assign = useAssignUserOrganization({
    fetch: writeFetch,
    mutation: {
      onSuccess: (response) =>
        response.status === 204
          ? message.success("组织归属已更新")
          : message.error("更新失败"),
    },
  });
  const replace = useReplaceUserRoleGrants({
    fetch: writeFetch,
    mutation: {
      onSuccess: async (response) => {
        if (response.status === 200) {
          message.success("角色授权已更新");
          await grantsQuery.refetch();
        } else if (response.status === 409)
          message.warning("授权已被其他管理员修改，请刷新后重试");
        else message.error("更新失败");
      },
    },
  });
  return (
    <Space direction="vertical" size="large" className="workspace">
      <Card title="选择用户">
        <Typography.Paragraph type="secondary">
          当前契约尚无用户目录接口，请输入从邀请或审计记录获得的用户
          UUID。此处不会信任或切换租户。
        </Typography.Paragraph>
        <Input.Search
          placeholder="用户 UUID"
          enterButton="读取授权"
          onSearch={setUserId}
        />
      </Card>
      {userId && granted.has("organization.user-organization/update") && (
        <Card title="组织归属">
          <Form
            layout="inline"
            onFinish={(data) => assign.mutate({ userId, data })}
          >
            <Form.Item
              name="departmentId"
              label="主部门"
              rules={[{ required: true }]}
            >
              <Select
                style={{ width: 220 }}
                options={departments.map((item) => ({
                  value: item.id,
                  label: item.name,
                }))}
              />
            </Form.Item>
            <Form.Item name="positionId" label="岗位">
              <Select
                allowClear
                style={{ width: 220 }}
                options={positions.map((item) => ({
                  value: item.id,
                  label: item.name,
                }))}
              />
            </Form.Item>
            <Button type="primary" htmlType="submit" loading={assign.isPending}>
              保存
            </Button>
          </Form>
        </Card>
      )}
      {userId && grants && granted.has("authorization.user-roles/update") && (
        <Card title="角色授权">
          <Form
            key={grants.version}
            initialValues={{ roleIds: grants.roleIds }}
            onFinish={(data: { roleIds: string[] }) =>
              replace.mutate({
                userId,
                data: {
                  expectedVersion: grants.version,
                  roleIds: data.roleIds ?? [],
                },
              })
            }
          >
            <Form.Item name="roleIds" label={`角色（版本 ${grants.version}）`}>
              <Select
                mode="multiple"
                options={roles.map((role) => ({
                  value: role.id,
                  label: role.name,
                }))}
              />
            </Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              loading={replace.isPending}
            >
              保存期望状态
            </Button>
          </Form>
        </Card>
      )}
    </Space>
  );
}

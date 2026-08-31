export const navigation = [
  { key: "/", label: "概览" },
  {
    key: "/organization/departments",
    label: "部门管理",
    permission: "organization.departments/read",
  },
  {
    key: "/organization/positions",
    label: "岗位管理",
    permission: "organization.positions/read",
  },
  {
    key: "/organization/users",
    label: "用户授权",
    permission: "authorization.user-roles/read",
  },
  {
    key: "/authorization/roles",
    label: "角色与策略",
    permission: "authorization.roles/read",
  },
  {
    key: "/settings/dictionaries",
    label: "字典管理",
    permission: "settings.dictionaries/read",
  },
  {
    key: "/settings/configurations",
    label: "系统配置",
    permission: "settings.configurations/read",
  },
  { key: "/audit", label: "审计日志", permission: "audit.events/read" },
];

export function visibleNavigation(granted: ReadonlySet<string>) {
  return navigation.filter(
    (item) => !item.permission || granted.has(item.permission),
  );
}

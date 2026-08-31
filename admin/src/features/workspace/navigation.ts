export const navigation = [
  { key: "/", label: "概览" },
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

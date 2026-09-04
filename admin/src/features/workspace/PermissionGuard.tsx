import { Result } from "antd";
import type { ReactNode } from "react";

import { usePermissions } from "./use-permissions";

export function PermissionGuard({
  permission,
  children,
}: {
  permission: string;
  children: ReactNode;
}) {
  const granted = usePermissions();

  if (!granted.has(permission)) {
    return <Result status="403" title="无权访问此功能" />;
  }

  return children;
}

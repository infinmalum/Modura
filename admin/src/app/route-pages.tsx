import { Spin } from "antd";
import { lazy, Suspense, type ComponentType } from "react";

function deferred<T extends object>(
  load: () => Promise<T>,
  select: (module: T) => ComponentType,
) {
  const Page = lazy(async () => ({ default: select(await load()) }));
  return function DeferredPage() {
    return (
      <Suspense fallback={<Spin fullscreen />}>
        <Page />
      </Suspense>
    );
  };
}

export const AuditRoute = deferred(
  () => import("../features/audit/AuditPage"),
  (module) => module.AuditPage,
);
export const ConfigurationsRoute = deferred(
  () => import("../features/settings/ConfigurationsPage"),
  (module) => module.ConfigurationsPage,
);
export const DictionariesRoute = deferred(
  () => import("../features/settings/DictionariesPage"),
  (module) => module.DictionariesPage,
);
export const WorkspaceRoute = deferred(
  () => import("../features/workspace/Workspace"),
  (module) => module.Workspace,
);

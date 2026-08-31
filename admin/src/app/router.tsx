import { createBrowserRouter } from "react-router-dom";

import { LoginPage } from "../features/auth/LoginPage";
import { PlatformLayout } from "../features/platform/PlatformLayout";
import { PlatformLoginPage } from "../features/platform/PlatformLoginPage";
import { AdminLayout } from "../features/workspace/AdminLayout";
import {
  AuditRoute,
  ConfigurationsRoute,
  DictionariesRoute,
  DepartmentsRoute,
  PositionsRoute,
  PlatformConfigurationsRoute,
  PlatformDictionariesRoute,
  PlatformTenantsRoute,
  RolePoliciesRoute,
  RolesRoute,
  UserAssignmentsRoute,
  WorkspaceRoute,
} from "./route-pages";

export const router = createBrowserRouter([
  { path: "/login", element: <LoginPage /> },
  { path: "/platform/login", element: <PlatformLoginPage /> },
  {
    path: "/platform",
    element: <PlatformLayout />,
    children: [
      { index: true, element: <PlatformTenantsRoute /> },
      { path: "settings/dictionaries", element: <PlatformDictionariesRoute /> },
      {
        path: "settings/configurations",
        element: <PlatformConfigurationsRoute />,
      },
    ],
  },
  {
    path: "/",
    element: <AdminLayout />,
    children: [
      { index: true, element: <WorkspaceRoute /> },
      { path: "organization/departments", element: <DepartmentsRoute /> },
      { path: "organization/positions", element: <PositionsRoute /> },
      { path: "organization/users", element: <UserAssignmentsRoute /> },
      { path: "authorization/roles", element: <RolesRoute /> },
      {
        path: "authorization/roles/:roleId/policies",
        element: <RolePoliciesRoute />,
      },
      {
        path: "settings/dictionaries",
        element: <DictionariesRoute />,
      },
      {
        path: "settings/configurations",
        element: <ConfigurationsRoute />,
      },
      { path: "audit", element: <AuditRoute /> },
    ],
  },
]);

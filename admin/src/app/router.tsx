import { createBrowserRouter } from "react-router-dom";

import { LoginPage } from "../features/auth/LoginPage";
import { AdminLayout } from "../features/workspace/AdminLayout";
import {
  AuditRoute,
  ConfigurationsRoute,
  DictionariesRoute,
  WorkspaceRoute,
} from "./route-pages";

export const router = createBrowserRouter([
  { path: "/login", element: <LoginPage /> },
  {
    path: "/",
    element: <AdminLayout />,
    children: [
      { index: true, element: <WorkspaceRoute /> },
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

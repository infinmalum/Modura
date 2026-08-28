import { createBrowserRouter } from "react-router-dom";

import { Workspace } from "../features/workspace/Workspace";

export const router = createBrowserRouter([
  { path: "/", element: <Workspace /> },
]);
